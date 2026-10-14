# CC100 fw 2.30: range 0x1c000-0x26fff (second wave, fork G)

This range holds the device/motion layer below `plot_device`, the registration-mark sensor
driver, USB helpers and the Renesas H8S C runtime library (soft float, math). Symbols are in
`sym/wave2_G.txt` and readable C is in `src/subsys_G.c`. Addresses are ROM addresses.

Units: GP-GL coordinates are multiplied by `steps_per_unit` (default 20) to give the
*internal* unit of 0.005 mm (1 mm = 200; see notes/input.md, notes/device.md). The motion
queue and the motor position counters (`motor_pos_x/y`, 0xffbf26/0xffbf6a) use internal/10,
i.e. **0.05 mm per motor count** (`motion_move_steps` 0x1d582 divides by 10,
`motion_get_pos` 0x1d5fe multiplies by 10). Whether one count is one physical motor
(micro)step is decided in the stepper ISRs at 0x1bc6a/0x1bcc4 (outside this range; TODO there).

## 1. Device coordinate pipeline

`plot_to` (0x1db80) -> ... -> `plot_device(x, y, pen)` (0x1c572). From there:

1. `apply_origin_offsets` 0x1c824: `p = (x + offset_x + origin_x + 0xffc2e4,
   y + offset_y + origin_y + 0xffc2e8)`. The 4th argument skips the GP-GL offset (`^`).
   0xffc2e4/8 is a second origin term (TODO: set by the registration code).
2. If `regmark_active` (0xffb0aa): `regmark_clip_plot` 0x1c146 clips the segment against the
   registration area rectangle `rm_area_*` (0xffbc20..0xffbc2c) with Cohen-Sutherland
   (`regmark_area_outcode` 0x1c242, `regmark_clip_segment` 0x1c2d6, edge intersections
   0x1c39a/0x1c414 using `muldiv`), then continues at step 3. Otherwise the point is stored in
   `rm_last_x/y` and goes straight to step 3.
3. `device_plot` 0x1c608:
   - `apply_regmark_and_cal` 0x1c918: if `regmark_active`, `regmark_transform` 0x1c6c4 applies
     the affine registration correction (doubles):
     `x' = x*rm_xx + y*rm_xy`, `y' = y*rm_yy + x*rm_yx`, optionally pre-scaled by
     `rm_sx/rm_sy` when `rm_scale_on` (0xffb0ab). Then the calibration offset
     `regmark_off_x/y` (0xffb4f0/4) is added.
   - `apply_orientation` 0x1c9c0: adds `orient_off_x/y` (0xffb4f8/c), records the result in
     0xffb54c/0xffb550 (user-visible device position) and, in landscape, swaps the axes
     using `page_width` (TODO: exact formula).
   - `machine_clip_plot` 0x1caca: Cohen-Sutherland clip against the machine travel limits
     `mach_xmin/ymin/xmax/ymax` (0xffb524..0xffb530). `machine_outcode` 0x1cbc6 drops the
     x>max bit when 0xffb3ba is 0, so the media can feed beyond the soft limit.
4. `track_feed_min_and_queue` 0x1ce28: for pen-down segments it tracks the minimum X reached
   (`feed_min_x`, 0xffbc3a). Then it calls `queue_point` 0x1d0bc.
5. `queue_point`: `distance_correct` (0xf4b2, outside this range) scales the point by the
   distance-correction factors (double at 0xffb25a; skew and scale, rounded with `dround`
   0x1f3ea). Then `queue_segment` 0x1d10c.

### Short pen-up move merging

`queue_segment` (0x1d10c) does not send every pen-up move directly:

- A pen-**up** move is remembered as a *pending* target (`pending_up`, `pending_up_x/y`) and
  the tool is lifted (`tool_set(0)`). Nothing moves yet.
- A pen-**down** move that starts within `seg_merge_dist` (0xffb2b2) of the last position
  (pending-up target if one exists, else the last queued point) in both axes is *deferred*
  (`deferred_valid` 0xffc422, `deferred_x/y/pen`). It is sent when the next move arrives or
  when `tool_set(0)` or `flush_deferred_segment` (0x1da92) runs.
- `queue_segment_now` 0x1d31a: first flushes a pending pen-up move (`motion_emit(..., 0)`),
  then subtracts `home_off_x/y` (0xffc394/8, `sub_home_offset` 0x1d3c2) and calls
  `motion_emit`. Bit 0xffbc52 records a negative pen value.

`motion_emit` 0x1d3fe: normally `motion_line` 0x1d4aa: `tool_set(pen)` then
`motion_move_steps` 0x1d582, which stores the target in `motor_target_x/y` (0xffbc3e/42) and
calls the motion planner at 0x19b9c with `(x/10, (y + motor_y_bias)/10, pen)`. When
registration correction with a non-zero origin is active (0xffb288 != 0, outside scan mode)
it goes through 0x10f4e/0x10d90/0x10686/0x10ce0 instead (outside this range; TODO).

### Tool (pen/blade) up-down

`tool_set(down)` 0x1d6f8:
- `down == 0`: flushes a deferred segment, and if the tool is down (bit 5 of `tool_state`
  0xffbc46) clears the bit and `tool_force_applied` (0xffbc66).
- `down != 0`: if the tool is already down with the same force (`tool_force` 0xffb4ba) it
  returns; otherwise it sets bit 5 and remembers the force.
- Both paths then call 0x19acc (flush the planner queue) and 0x17c1c(force) which drives the
  solenoid / force PWM (outside this range).

`wait_motion_done` 0x1db52: flushes (0x17f2e, 0x19acc) and spins `delay_ticks(1)` until both
axis busy flags (`motor_busy_x/y`, 0xffbf14/0xffbf58) are 0.

`motion_get_pos` 0x1d5fe: reads the motor counters `motor_pos_x/y`, copies them to
0xffb5e0/4, 0xffb600/4/8 (planner start position) and returns
`(x*10, y*10 - motor_y_bias)` in internal units.

### Continuous axis drive (jog, sensor scan, homing)

`motor_run_continuous` 0x1d7ac(dx, dy) runs the X and/or Y motor at a constant rate without
the planner. It is used by homing (0x14ef0), jogging and the sensor scans. For each axis:
the value 0 stops it (waits for `jog_dir_x/y` 0xffbbd6/7 to clear); otherwise its sign gives
the direction, and its magnitude becomes `jog_speed_x/y` (0xffb5f0/2). On start it programs a
TPU channel: X at 0xffff24..0xffff28 (TCR=0x41, TCNT=0, TGR=min(160000, 60000)), Y at
0xffff34..0xffff38, enabling it in TSTR (0xfffeb0 bit 1 / bit 2). It also updates the
port 0x600010 shadow (`port600010_shadow` 0xffb5de; bits 2..3 for X, 0..1 for Y; probably
motor enable/current). A direction reversal while running stops the axis first. The routine
ends with `delay_ticks(speed_x)` (TODO: the argument packing of this function is not
fully clear in the decompile).

## 2. Line-type helpers (used by `plot_linetype`)

- `dash_select_pattern` 0x1dc9e: for |`line_type`| 1..8 it selects pattern `n` from RAM
  (0xffbf76 + 42*n, loaded from ROM by `load_line_patterns`) with its total length from the
  int table at 0xffc0ee; type 10 selects the user pattern at 0xffbcae. `line_pitch` becomes
  the remaining length.
- `dash_next_element` 0x1dd5a: takes the next element (bit 15 = pen up), wraps at the 0xffff
  end marker, and scales it to the pitch: `len = elem * pitch_left / total_left`
  (`muldiv`), so rounding error never accumulates.
- `user_dash_restart` 0x1de20 / `user_dash_next` 0x1de48: the same for `L100` custom
  patterns (`user_lt_n` elements of 4 bytes at 0xffbc96).
- `clip_isect_xmax/xmin/ymax/ymin` 0x1e9a6..0x1eaa8: intersection of a segment with the
  window clip lines (`clip_xmax` etc.); used by `plot_transform`.

## 3. Jog keys

`jog_read_keys` 0x1ed42 turns `jog_keys` (0xffc424, set by `ESC 00 n`; active-low bits:
0 +X, 1 -X, 2 +Y, 3 -Y) into a direction vector. `jog_with_keys` 0x1edd6 (called while
waiting for data) does a single 10-unit (0.05 mm) step for a short press; if the key is
still held after 20 x 10 ticks it drives the axes continuously with `jog_axis_ramp` 0x1f0ec
(speed ramps in steps of 5 up to 15, or 2 near the end, with a braking-distance check
against the travel limits `travel_max_x/y` 0xffbefa/e). Finally it re-syncs the position.

## 4. Registration-mark sensor

Hardware: a reflective sensor with an LED on port 0xffff06 bit 4, read by the A/D converter
(ADCSR 0xffff98, ADCR 0xffff99, result ADDRA 0xffff90/91, 10-bit = value >> 6). The LED/gain
is set with a serial digital potentiometer clocked on port 0xffff00 bits 1..3
(`sensor_pot_shift` 0x219a0, `sensor_pot_step` 0x219fa).

- `sensor_init` 0x2195c: loads the gain step count from EEPROM 0x19 and sets the detection
  threshold `sensor_threshold` (0xffc22c) = 0x200.
- `sensor_autocal_gain(blink)` 0x21a26: steps the pot until the A/D value rises above the
  threshold, stores the step count in EEPROM 0x19; with `blink != 0` it then blinks the LED
  forever (service use).
- `sensor_on` 0x2078a / `sensor_off` 0x207bc: LED on + A/D start, then wait 30 ticks / LED off.
- `sensor_filter_init` 0x207d0 reads a baseline, fills a 110-entry ring buffer (0xffc12e) and
  sets up three ring pointers. `sensor_filter_step` 0x208d0 pushes a sample and keeps a
  windowed sum; the result is scaled by `sensor_gain_x/y` (0xffc124/6) /10 /16. An edge is
  reported when the filtered value exceeds `baseline * gain / 10`.
- `sensor_scan_axis` 0x20d6a: drives one axis with `motor_run_continuous` while sampling,
  records the position where the edge is detected (`scan_record_edge` 0x20a08), stops at
  `cancel_request`, at the travel limit (`scan_check_limits` 0x20a90; ±10000 counts or the
  4000-count search window) or on edge; then restores the logical position.
- `sensor_scan_diag` 0x21620: the same with both axes moving together (first-mark auto
  search, used by 0x15bbc).
- `regmark_locate_x` 0x2114c / `regmark_locate_y` 0x2135c: scan across a mark bar along X
  (resp. Y), back off by 400, scan again in both directions, and return the centre
  (average of the two edges). `regmark_result` (0xffc225) = 0 when found.
- `regmark_find_and_mark` 0x1ff2a (called from `regmark_calibrate`, TB70): from the current
  position it locates edges in four directions 1600 counts (8 mm) apart (X, +Y, -X, -Y),
  computes the mark centre (0x5ecc), cuts a cross through it with `move_clamped_wait`
  (0x21546), and moves back 6000 counts. Returns 0 on success, 1 on failure.
- `move_clamped_wait` 0x21546: orientation transform, clamp Y to the machine limits,
  distance correction, `queue_segment`, wait. Returns non-zero if it had to clamp.
- Service `t` helpers: `sensor_measure_reply` 0x21cb6 replies the raw A/D value (`%5d`);
  `sensor_trace_start` 0x21cd6 (`tN`, N = 0/1 axis) runs `sensor_trace_scan` 0x21dfa which
  records (position, value) pairs at 0xffcf00; `sensor_trace_read` 0x21d42 returns one
  pair per call (`%6d,%6d`), or `-1,-1` at the end.

## 5. Serial and USB helpers

- `sci0_init` 0x1f458: SCI0 SMR (0xffff78) = 0x08 (8 data bits, no parity, 2 stop bits),
  BRR (0xffff79) = 0x33 (9600 baud if the CPU clock is 16 MHz; TODO: confirm the clock),
  SCMR = 0xf2, then RE/TE/RIE enabled (SCR = 0x70). Clears the SCI escape state.
- `sci0_tei_isr` 0x1f61c: clears TEIE.
- `usb_bus_reset_handler` 0x223f4 (USB status bit 0x80): programs the external USB
  controller endpoint table (0xc00000..) and re-enables interrupts.
- `usb_suspend_handler` 0x2288c (bit 0x80000000), `usb_resume_handler` 0x228c0 (bit 0x01, also
  clears `jog_keys`). `usb_tx_idle` 0x1f66c disables the bulk-IN interrupt.
- `usb_ep_read_to_ring` 0x221e0 / `usb_ep_write_from_ring` 0x222b0: FIFO <-> ring buffer
  copies. `ring_copy` 0x22386. `be_bytes_to_ulong` 0x2292e / `le_bytes_to_ulong` 0x22986
  assemble setup-packet fields.
- `record_stack_ptr` 0x1f656 stores SP at 0xffbf9c (read back by service `y`: stack usage).

## 6. Runtime library (0x22a66-0x2667f)

Renesas H8S C runtime. Doubles are IEEE-754 64-bit passed **by pointer**
(`er0` = result, `er1` = left operand, `er2` = right operand; Ghidra cannot express `er2`, so
decompiled calls look like they take two arguments). Floats are passed in registers
(`er0`, `er1`). Errors set `libc_errno` (0xffc22e): 0x44c = ERANGE, 0x44d = EDOM (TODO
values from the Renesas library manual).

| addr | name | meaning |
|---|---|---|
| 0x22a66 | `__dsub` | double a - b (negates b, falls into `__dadd`) |
| 0x22a8a | `__dadd` | double a + b |
| 0x22c6a | `__fsub` | float a - b |
| 0x22c74 | `__fadd` | float a + b |
| 0x22eee | `__ddiv` | double a / b (shift-subtract) |
| 0x234a0 | `__dmul` | double a * b |
| 0x236ac | `__fmul` | float a * b |
| 0x240a0 | `__dcmp` | returns 0 a<b, 1 a==b, 2 a>b, -1 unordered |
| 0x231f6 / 0x238bc / 0x23348 / 0x23338 / 0x232d8 | `__deq` / `__dne` / `__dlt` / `__dle` / `__dgt` | boolean wrappers of `__dcmp` |
| 0x238ca | `__dneg` | copy with sign flipped |
| 0x23358 / 0x232e4 | `__ltod` / `__itod` | long / int -> double |
| 0x2339e / 0x23316 | `__ltof` / `__itof` | long / int -> float |
| 0x23174 | `__dtol` | double -> long (truncate) |
| 0x23276 | `__ftol` | float -> long |
| 0x23202 / 0x230b4 | `__ftod` / `__dtof` | float <-> double |
| 0x2381e | `__mulsi3` | 32-bit multiply |
| 0x23052 / 0x23078 | `__divmodsi4` / `__udivmodsi4` | 32-bit divide, quotient er0, remainder er1 |
| 0x23836 | `__dcopy` | copy 8 bytes |
| 0x23854 | `__movblk` | memcpy(src er0, dst er1, n er2) |
| 0x238ec | `_INITSCT` | section init: copy ROM data to RAM (tables at 0x2963c/0x29644) |
| 0x22e16 | `__bitfield_helper` | builds `(1<<width)-1` masks for bit-field access (TODO: exact op) |
| 0x23c1a / 0x23d9c | `cos` / `sin` | double |
| 0x23cf8 / 0x23e7a | `cosf` / `sinf` | float (used by `angle_to_sincos`) |
| 0x23f1e / 0x23ffc | `sqrt` / `sqrtf` | |
| 0x23948 / 0x23a26 | `atan` / `atan2` | likely (one and two double arguments, same core 0x246ce) |
| 0x23b3c | `ceil` | trunc, then +1.0 if positive and not integral |

The math wrappers unpack into a 12-byte extended format (`__xd_unpack` 0x2414c,
`__xf_unpack` 0x24208), call a core (`__xd_sin` 0x25228, `__xd_cos` 0x24f02, `__xf_sin`
0x253de, `__xf_cos` 0x2509e, `__xd_sqrt` 0x24434, `__xf_sqrt` 0x245d6, `__xd_atan`
0x246ce, `__xd_atan2` 0x25582, `__xd_ceil` 0x24e46) built from extended add/mul/compare
helpers (0x25656 add, 0x2564a sub, 0x25dbe mul, 0x2633e cmp, 0x25f4c trunc, polynomial
evaluators 0x2604c/0x2613e/0x26290 and float versions 0x260c6/0x261ec), and pack the
result (`__xd_pack` 0x2427a, `__xf_pack` 0x2436c), setting `libc_errno` on overflow/NaN.

## Not determined here

- Speed table ('!' levels to mm/s): the speed is applied in the planner (0x19b9c/0x19f8c)
  and the stepper ISRs, outside this range.
- Pen force PWM (0x17c1c) and the planner/ISR internals (0x19xxx, 0x1bxxx).
- Cutter offset and blade rotation compensation are not in this range (no references
  to `pen_cutter_offset`/`pen_blade_rot` here).
