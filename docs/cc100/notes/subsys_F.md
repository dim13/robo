# CC100 firmware 2.30: system subsystems in 0x11000–0x1bfff (wave 2, fork F)

This file covers the functions in 0x11000–0x1bfff that were still unnamed after the first wave. The
GP-GL parser, reply path, plotting commands and F*/TB handlers in the same range were already
documented in `input.md`, `control.md`, `plot.md` and `device.md` and are not repeated here.
Symbols are in `sym/wave2_F.txt`, readable C in `src/subsys_F.c`.

Assumptions used for timing below: the system clock φ is 16 MHz. This is consistent with two
independent observations (SCI0 BRR = 51 gives exactly 9600 baud, and the 8-bit tick timer setup
gives exactly 1 kHz), but the crystal itself was not seen.

## 1. Boot, restart and factory test modes

| Address | Name | What it does |
|---|---|---|
| 0x13940 | `main` | Called from `reset` (0x400, SP = 0xFFEFB0). Peripheral setup: 0x1f644, 0x739c(0), 0x238ec (C runtime data/bss init), 0x1f660, 0x1f656, `tick_timer_init`, `eeprom_check_format`, USB init 0x6c3c, `rx_init`, 0x1f66c, SCI0 init 0x1f458. Then `motion_init`, `settings_apply`, `solenoid_init`, `pen_reset`, 0x1054c, 0x2195c. If any of the jumper/DIP bits 2..5 of port register 0xFFFFB2 are set, runs `factory_test_select`. Finally calls `power_off(0)`: the plotter boots into the *off* state and waits for the power key. |
| 0x12ff2 | `power_key_isr` | IRQ handler for the power key (vector 18 wrapper at 0x462). Key input is bit 0 of 0xFFFFBE. If the machine is on: waits for release (debounced in 100-loop chunks) and calls `power_off(0)`. If off: waits for release while counting 100-loop chunks (saturating at 51), powers the hardware up (`power_up_hw`, USB re-enable at 0xC000C4/0xC00086, IRQ re-enable), and if the key was held longer than 50 chunks it homes the carriage and runs the ROM demo (`demo_run`, 0xd512). Then `warm_restart`. |
| 0x139ce | `warm_restart` | Re-runs the software part of init (SCI, receive buffer, motion, settings, solenoid, pen) and enters `enter_gpgl`. |
| 0x137c4 | `enter_gpgl` | Clears `service_mode` and `rx_starved`, homes the carriage, reloads the cutter offset, and jumps into `gpgl_loop` (never returns). |
| 0x137ec | `boot_wait_key` | Alternative boot path with no static callers (TODO: probably reached through a function pointer or left over). It waits until the key has been pressed for 100 ticks, then samples it for another 1000 ticks; if it is still pressed it runs `demo_run`, otherwise the endless motion test `test_motion_loop`. |
| 0x13736 | `factory_test_select` | Reads a 4-bit selector from port 0xFFFFB2 bits 2..5 (jumpers or DIP switches on the board): 1 = `eeprom_format(0)` then blink forever, 3 = key mirrors LED, 4 = `test_motion_xy`, 5 = `test_circle` (draws a 0..108° arc each key press using sin/cos), 6 = `test_jog`, 7 = `test_sincos_bench`, 8 = sensor calibration 0x21a26, 9 = key mirrors port 0xFFFF06 bit 4. 2 and other values do nothing. |

## 2. Power management and thermal protection

* `power_off(error)` (0x12ec0) is the single "off" path. It sets the off flag (0xFFB397,
  named `usb_suspended` in `symbols.txt`; despite the name it is the global *power off* flag), stops
  motion state, de-energizes the hardware (`power_down_hw`), disconnects USB (0xC00080 |= 1), and
  then spins forever. Only the power-key interrupt gets it out. With `error != 0` the busy LED
  (port 0xFFFF09 bit 3) blinks a double-flash pattern forever instead, which is the overheat
  shutdown.
* `power_down_hw` (0x12f64) writes 0x0F to the motor current latch 0x600010 (all windings off),
  zeros the solenoid D/A (0xFFFDAC), turns the LED off and stops both axis timers.
  `power_up_hw` (0x12fb8) writes 0x0A (both motors at standby current) and clears the idle counters.
* `thermal_protect` (0x12b18, called every 100 ticks by the tick ISR 0x12d7a) keeps three integrators:
  X motor (0xFFB38A), Y motor (0xFFB38E) and pen solenoid (0xFFB392). The motor integrators grow
  or shrink by a step chosen from the current setting in the 0x600010 shadow (0xFFB5DE: bits 3..2
  = X, bits 1..0 = Y; 0 = full current +4, 1 = +2, 2 = +2 or −2 around 37801, 3 = −4/−6). The solenoid
  integrator grows by (DAC − 180)/10, minimum −2. If X or Y exceed 54000, or the solenoid integrator exceeds 43200, it calls `power_off(1)`,
  which is the blinking overheat shutdown. Service command `j` reports these three integrators (`device.md`).
* Idle power save: the tick ISR counts idle minutes. When they reach the value in 0xFFB388
  (EEPROM word 0x1a, default 5, service `g10`), it drops the motor current (0x600010 = 0x0F).

## 3. Tick timer and serial port

* `tick_timer_init` (0x12a2e): 8-bit timer 0 with TCR (0xFFFF68) = 0x4A, i.e. clock φ/64, cleared
  on compare match A, compare-A interrupt enabled. TCORA (0xFFFF6C) = 250, so the tick is
  16 MHz / 64 / 250 = **1 kHz (1 ms)**. Also clears the idle and thermal counters.
* SCI0 / RS-232 (init 0x1f458, outside this range but noted for completeness): SMR = 0x08,
  BRR = 51 (0x33), SCMR = 0xF2, SCR = 0x70. That is **9600 baud, 8 data bits, no parity, 2 stop
  bits**, asynchronous, receive interrupt enabled. No hardware handshake signals are touched.

## 4. USB descriptors (ROM)

The USB controller is external (registers at 0xC00000; USB stack mostly at 0x6xxx/0xfxxx, outside
this range). The descriptor table at 0x29136 lists (length, pointer) pairs:

* Device descriptor at 0x29166: `12 01 00 01 00 00 00 40 4d 0b 0a 11 00 01 00 00 00 01`
  * bcdUSB 1.00, class/subclass/protocol 0 (defined per interface), EP0 max packet 64
  * **VID 0x0B4D (Graphtec), PID 0x110A**, bcdDevice 1.00
  * iManufacturer, iProduct and iSerialNumber are all 0 (no strings); 1 configuration
* Configuration at 0x29178 (total 32 bytes): 1 interface, configuration value 1, attributes 0xC0
  (self-powered), max power 0x10 (32 mA).
  * Interface 0: 2 endpoints, **class 7 (Printer), subclass 1, protocol 2 (bidirectional)**.
  * EP 0x01 OUT, bulk, 64 bytes (host to plotter, GP-GL data).
  * EP 0x82 IN, bulk, 64 bytes (replies).
* String 0 (language IDs) at 0x29198: 0x0409 (US English).
* String at 0x2919c: `"123"` (UTF-16). Not referenced by the device descriptor (all string indices are 0).
* Then four 5-byte ASCII entries `12342`, `12343`, `qrstt`, `qrstu` (TODO: purpose unknown, they look like test data for the descriptor/IEEE-1284 ID path).

## 5. Settings EEPROM driver (Microwire, 93Cxx type)

The settings live in a serial EEPROM accessed by bit-banging port register 0xFFFF00: bit 5 = SK
(clock, `uwire_sk`), bit 6 = CS (`uwire_cs`), bit 7 = DI (`uwire_di`). DO / ready is read on
port 0xFFFFB0 bit 4. The firmware addresses EEPROM bytes as 0x300000 + offset; the word/byte
access functions (`eeprom_read_word` etc., named in the first wave) live in 0x126xx–0x128xx.

* `uwire_shift_out(nbits, bits)` (0x123e0) shifts out MSB first: clock low, DI = bit, clock high.
* `uwire_command` (0x12440) wraps it in CS high/low.
* Opcodes used: 0x98 << 24 with 11 bits = **EWEN** (`eeprom_ewen` 0x1246a), 0x80 << 24 with 11 bits =
  **EWDS** (`eeprom_ewds` 0x12492), 0x88 << 24 = **ERAL** (`eeprom_erase_all` 0x124ba, which waits for
  DO to go ready). The 11-bit command length (start bit + 2 opcode bits + 8 address bits) fits a
  93C56/93C66 in x8 or a 93C46 in x8 organisation (TODO: exact part).
* Helpers: `eeprom_read_long`/`eeprom_write_long` (two words, low word first),
  `eeprom_set_bit`/`clear_bit`/`toggle_bit` (bit read-modify-write).
* Format check (`eeprom_check_format`, 0x12332): word 0x02 is a format marker. If it reads < 1 it
  is treated as 9999. If it is below 10001 (0x2711), `eeprom_format(marker + 1)` is run and the
  marker is set to 10001. `eeprom_format(0)` (factory test) erases everything, writes defaults and
  blinks the LED forever. Formats 10000/10001 only write defaults; 10002..10004 skip that and also
  skip the factory reset of user settings (`eeprom_factory_reset`, 0x11f7e).
* `eeprom_load_defaults` (0x11e3e) writes: 0x02 = 10001, 0x13 = 14, 0x14 = 0, 0x16 = 0,
  0x08/0x0a/0x0c = 164/177/203 (force table, see §6), 0x12 = 0, 0x29 = 'x' (0x78: nibbles 8 and 7),
  0x18 = 0, 0x28 = 5, 0x20 = 0x22 = 0, 0x1a = 5 (idle minutes), 0x2a = 0, 0x19 = 0, 0x2c = 0,
  0x39 = 3, and the config word 0x04 (written through `config_word_save`).
* Config word: EEPROM long at 0x04, mirrored in RAM at 0xFFB368 (`load_config_word` 0x1228e also
  sets 0xFFB500 = 1000000 and reloads the step size). Bit meanings as far as known: bits 3..2 = step
  size index (`load_step_size`, `g6`), bit 1/4/6/7 = flags set by `g0`/`g5`/`g1`/`g3` (TODO: meaning),
  bit 5 = "do not lift pen when idle" (checked in `rx_peek`).
* `settings_load` (0x121da) at boot and after `TT`: force table, config word, force, reg-marks,
  axis correction (`load_axis_correction` 0x12040: distance correction words 0x1c/0x1e converted to
  double scale factors, plus axis tweak bytes 0x13/0x14 clamped to ±60 and word 0x16, passed to
  0xf096), 0x10332, idle minutes, solenoid profile select (0x2c, clamped 0..1), media mode = 1.

## 6. Pen / cutter solenoid

The tool is pressed by a solenoid driven from the D/A converter register **0xFFFDAC** (8 bit).

* `solenoid_init` (0x17900): D/A enabled (0xFFFDAE = 0x40), output 0, loads three delays from
  EEPROM: 0x2e = settle after press (default 20 ticks), 0x30 = delay before lift (0), 0x32 = settle
  after lift (default 70 ticks).
* `solenoid_set(level)` (0x17c1c):
  * `level == 0` (pen up): optional pre-lift delay, then plays a lift profile of up to five
    (ticks, D/A value) steps from RAM table 0xFFC368, sets D/A to 0 and waits the lift settle time.
    During the profile the USB bulk-OUT interrupt is masked and 0xFFB2F2 is set, so no host data
    is accepted while the solenoid is switching.
  * `level != 0` (pen down): unless the force was given explicitly by `FX`, the level is scaled by
    the media-type percentage table at 0xFFC385 and clamped to 0..255. If the pen is already down
    only the D/A value changes. Otherwise a press profile (five (ticks, value) steps from
    0xFFC2F0[profile 0..1][speed−1][step], profile from EEPROM 0x2c / service `g7`) is played as a
    kick, then the D/A is set to the hold level and the press settle delay runs.
* `force_level_value(n)` (0x181da) returns entry n (0..3, clamped) of the force table at 0xFFC37C;
  entries 1..3 come from EEPROM words 0x08/0x0a/0x0c (`load_force_table` 0x178c4, defaults
  164/177/203), used by the `*` command's force levels. `force_hold_level` (0x1820a) returns entry 3.
* `select_pen(n)` (0x17fac, n = 0..8, ignored if already current): flushes and drains motion,
  saves the position, sets `cur_pen`, reloads the per-pen parameters and lifts the pen.
  `pen_save_position` (0x17ea6) / `pen_return_to_saved` (0x17f2e) keep the logical position across a
  tool change; 0xFFC394/0xFFC398 is a per-tool offset (always cleared by `clear_tool_offset`).

## 7. Coordinate transforms (double precision)

These are the stages of the logical-to-device chain called from 0x18e44 / 0x18f0e (see `plot.md`).
All take `double *x, double *y`.

| Address | Name | Operation |
|---|---|---|
| 0x18864 | `xform_factor` | x = x·p/r, y = y·q/r (`&p,q,r`) |
| 0x186e6 | `xform_rotate` | rotation about (rot_cx, rot_cy) with single-precision sin/cos (`/`) |
| 0x1863e | `xform_orientation` | mirror X unless orient bit 5, mirror Y unless bit 6 (subtract from 0xFFB578/0xFFB574), swap x/y unless flags are 0 or 0x60 |
| 0x18930 | `xform_device_offset` | stores the rounded pre-offset position (0xFFB27A/7E), then adds the cutter-offset term 0xFFC2E4 and the device offset 0xFFB4E8/0xFFB4EC |
| 0x18a12 | `xform_window_origin` | subtracts 0xFFB57C/0xFFB580 |
| 0x18a78 | `xform_user_offset` | adds `offset_x`/`offset_y` (`^`) |
| 0x18f86 | `xform_call_long` | adapter: converts long x/y to double, calls a stage, converts back |
| 0x18ae0 | `device_to_logical` | inverse direction: device position (0xFFBC3E/42) minus backlash/offset terms, landscape swap, page origin and reg-mark origin subtraction, inverse reg-mark transform; result in 0xFFBC30/34 |
| 0x19518 | `device_to_user_pos` | same idea for an arbitrary point (used by 0x10d90) |

## 8. Motion queue and stepper drive

* Stepper outputs: the X and Y motors are driven through external latches. Each step writes
  a 16-bit word from the 32-entry microstep phase table 0xFFC3A2 to **0x600008 (X)** or
  **0x600000 (Y)**. Motor current is selected through latch **0x600010** (two bits per axis,
  shadow 0xFFB5DE; 0x0A = normal, 0x0F = off).
* Each axis has a 16-bit timer channel: X uses 0xFFFF20..0xFFFF2F (compare value at 0xFFFF28),
  Y uses 0xFFFF30..0xFFFF3F (compare at 0xFFFF38). Their compare interrupts (vectors 40 and 44)
  run `timer_x_isr` (0x1bc6a) and `timer_y_isr` (0x1bcc4). Periods longer than 0xF06E counts are
  split into 60000-count chunks.
* `axis_profile_step(axis)` (0x14b5e) is the per-step state machine of one axis (state at +2):
  0 = idle, 1 = acceleration (period interpolated from the speed table 0xFFB640 with 10-bit
  fractions), 2 = constant speed with fractional accumulator, 3 = deceleration (table in reverse),
  4 = constant speed to the end. It advances the phase index by the direction (+4), writes the next
  phase word and returns the next timer period.
* Queue: an 8-entry ring of 0x42-byte segment records at 0xFFB7E2, head 0xFFB7D0, tail 0xFFB7D4.
  `motion_start_next` (0x1b162) loads the next record into both axis state blocks (0xFFBF14,
  0xFFBF58) and starts the timers. When the ring is empty it sets the motor current latch to 0x0A.
* `motion_line_to(x, y, pen_down)` (0x19b9c), in device steps:
  1. On a pen up/down change, the queue is drained first.
  2. The direction is computed in tenths of a degree (`atan_tenths`).
  3. **Corner handling:** if the direction changes by more than 25.0° and the move is longer than
     80 steps on X, the queue is drained (the tool stops at the corner); shorter moves are queued
     without stopping.
  4. **Backlash compensation on X:** when the X direction reverses, an offset 0xFFB61E is ramped
     towards the backlash value 0xFFB620 (EEPROM 0x2a, service `g4`/`c`) and added to all target
     X positions.
  5. The pen-up and pen-down speed profiles are selected (0xFFB628.. vs 0xFFB634..).
  6. The segment is planned with `motion_plan_segment` (0x19f8c) and queued.
* `motion_build_profile(speeds)` (0x19780): low word = pen-down speed, high word = pen-up speed,
  with acceleration constant 3000. It fills the period table 0xFFB640 used by the axis ISRs.
  Service command `h` (0x19764) rebuilds it with a new acceleration value.
* `motion_drain` (0x19acc) flushes the pending segment and busy-waits until the queue and both
  axes are idle.
* `jog_x_step` / `jog_y_step` (0x1bd1e / 0x1bf32): timer ISR variants used while jogging with the
  panel keys or the `ESC 00 n` remote jog. They step with a speed ramp over the same table.
* `atan_tenths(dy, dx)` (0x1bba6): integer atan2 in 0.1° using a polynomial approximation.
  `vector_angle(v)` (0x164c6) is the same idea using the ROM table 0x28de2 (128 entries, 0..45.0°).
  `isqrt32` (0x165b0) is a bit-by-bit integer square root.

## 9. Registration-mark search helpers

Used by `TB23`/`TB123` and friends (handlers at 0x1575a / 0x15fb2, see `device.md`).

* `regmark_compute_spans` (0x15a1a): the distances between marks are the page size minus 2000 steps
  (100 mm); if that is shorter than the mark length, half the mark length is used.
* `regmark_find_type1` (0x15212) / `regmark_find_type2` (0x154ca): move to each expected mark
  (0x21546) and run the edge searches 0x2114c / 0x2135c. Three marks are searched, or four unless
  bits 5..4 of `regmark_cfg` (0xFFC100) are 01 (three-mark mode). The first failure sets `regmark_result = 1`.
* `regmark_scan_first` (0x15bbc): for automatic search, steps forward by twice the mark length
  until the sensor sees the first mark, up to 20000 steps (1 m).
* `regmark_goto_origin` (0x1596c): moves to the corrected origin and resynchronises the position.
* `page_limits_update` / `page_limits_apply` / `origin_reset` (0x14ff4 / 0x150dc / 0x151be):
  round the page size to 1 mm (20 steps), swap it for landscape, and reset plot state and origin.

## 10. Service `g` sub-handlers (table at 0x28F14, dispatcher `g` 0x16bc0)

Every handler first reads `op`: 0 = query (reply `%5d` + terminator), 1 = set (more ints follow).

| g | Address | Target |
|---|---|---|
| 0 | 0x16c8e | config bit 1 |
| 1 | 0x16cfc | config bit 6 |
| 2 | 0x16d6a | axis tweak bytes EEPROM 0x13/0x14 (query replies two values; op 2 draws the calibration pattern `axis_correction_pattern` 0x173b4) |
| 3 | 0x16e22 | config bit 7 |
| 4 | 0x16e90 | X backlash, EEPROM 0x2a (also loaded into 0xFFB620) |
| 5 | 0x16f0c | config bit 4 |
| 6 | 0x16f7a | step size index (config bits 3..2), then `load_step_size` |
| 7 | 0x16ff8 | solenoid press profile 0/1, EEPROM 0x2c |
| 10 | 0x17082 | power-save idle minutes, EEPROM 0x1a |
| 11 | 0x170e8 | `g11,i`: i = 0..2 query force table word i; i = 3..5 set word i−3 and reload; i = 6..8 press the solenoid at that level; i = 9 lift |
| 12 | 0x171b2 | reg-mark parameter EEPROM 0x28 |
| 13 | 0x17220 | reg-mark nibbles EEPROM 0x29 (replies low, high) |
| 14 | 0x172dc | mark-sensor offset words EEPROM 0x20/0x22 |
| 15 | 0x17386 | `g15,0` runs the reg-mark calibration |

## 11. Trivial or unresolved functions (not given symbols)

| Address | Note |
|---|---|
| 0x12034, 0x1318a, 0x1615e, 0x1617c, 0x17f06, 0x17f12, 0x18224, 0x18256, 0x18626, 0x18918 | empty functions (hooks compiled out) |
| 0x1626a, 0x162c8, 0x162e6 | wrappers around the empty 0x1615e and the 64-bit divide; used only by the arithmetic self test (`e`, 0x6140) |
| 0x163a4, 0x163d8, 0x1640c | pass-through wrappers for the double routines 0x23d9c, 0x23c1a, 0x23948 (sin/cos family, see the runtime library notes) |
| 0x16440 | double helper: returns π/2 (0x28dba) for a zero argument, otherwise calls 0x23a26 (TODO: atan guard) |
| 0x16982 | double polynomial with constants −1.0 and 2.0 (TODO: series evaluation; no callers) |
| 0x13a64 | calls `muldiv_l(10, 20, 30)` and discards the result (dead test code) |
