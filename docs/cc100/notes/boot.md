# CC100 fw 2.30: boot, flash layout and update path

Readable C is in `../src/boot.c` and new symbols are in `../sym/boot.txt`. Addresses are ROM
addresses in `cc100ver230.bin`.

## 1. Flash image layout

| ROM range       | Content |
|-----------------|---------|
| 0x00000–0x001FF | Exception vector table: 128 × 32-bit. Unused vectors are 0xFFFFFFFF. |
| 0x00200–0x003FF | Erased (0xFF). |
| 0x00400–0x00719 | `reset` (SP = 0xFFEFB0, jsr `main`), then interrupt wrappers. Each wrapper is `push er6; stm er0-er1; jsr handler; ldm; pop er6; rte`. Unused vectors and traps are bare `rte` stubs. |
| 0x0071A–0x03FFF | **Erased** (14 822 × 0xFF). |
| 0x04000–0x0402B | **Application entry table**, 11 × 32-bit pointers (section 2). |
| 0x0402C–0x040FF | Erased. |
| 0x04100–0x298DF | Application code, constants and the initialised-data image. |

The `.mot` file in `docs/` (`cc100ver230.mot`) has these properties:

* S0 header `"main    mot"`.
* 3166 S1 records (16-bit addresses) and 6544 S2 records (24-bit addresses).
* S9 entry point **0x0400**.

It covers exactly the non-0xFF ranges of the `.bin` and is byte-identical to it:

* 0x0000–0x01A7: the vectors, written sparsely (unused vectors are omitted);
* 0x0400–0x0719;
* 0x4000–0x402B;
* 0x4100–0x298DF.

So the update file **includes the vector table and the 0x0000–0x3FFF area**. An update
programs the whole flash, block 0 included.

On H8S/2215-class parts, 256 KB flash is usually split into these erase blocks (from the
datasheet family, not verified against the exact part):

* EB0–EB7: 8 × 4 KB, 0x00000–0x07FFF;
* EB8: 32 KB, 0x08000–0x0FFFF;
* EB9–EB11: 3 × 64 KB.

Under that layout:

* the vectors and wrappers fill EB0;
* EB1–EB3 (0x1000–0x3FFF) are empty;
* the application starts at the EB4 boundary with its entry table.

0x0000–0x3FFF is therefore a reserved 16 KB **boot area**. In this release it contains only the
vectors and the wrappers; no loader is present.

## 2. Application entry table at 0x4000

| # | ROM | Pointer | Target | Same target as |
|---|-----|---------|--------|----------------|
| 0 | 0x4000 | 0x13940 | `main` | `reset` 0x400 |
| 1 | 0x4004 | 0x1BC6A | `timer_x_isr` | vec 40 wrapper 0x41A |
| 2 | 0x4008 | 0x1BCC4 | `timer_y_isr` | vec 44 wrapper 0x432 |
| 3 | 0x400C | 0x220B2 | `usb_isr` | vec 104 wrapper 0x44A |
| 4 | 0x4010 | 0x12FF2 | key (IRQ2) ISR | vec 18 wrapper 0x462 |
| 5 | 0x4014 | 0x12D7A | `timer_tick_isr` | vec 64 wrapper 0x47A |
| 6 | 0x4018 | 0x1F592 | `sci0_eri_isr` | vec 80 wrapper 0x492 |
| 7 | 0x401C | 0x1F4A4 | `sci0_rxi_isr` | vec 81 wrapper 0x4AA |
| 8 | 0x4020 | 0x1F5C6 | `sci0_txi_isr` | vec 82 wrapper 0x4C2 |
| 9 | 0x4024 | 0x1F61C | `sci0_tei_isr` | vec 83 wrapper |
| 10 | 0x4028 | 0x137EC | `boot_wait_key` | **none** |

The wrapper code at 0x400–0x719 has exactly ten `jsr` instructions (verified by disassembly).
They call entries 0–9 **directly**, in the same order as the table. Nothing in the image reads
0x4000–0x402B: there are no references in code or data, and no indirect jumps through it. The
only indirect calls in the firmware go through the USB dispatch table (0x290D8) and one
function-pointer argument.

Interpretation:

* The table is the fixed interface between a boot block (0x0–0x3FFF) and an application that
  can be relinked independently.
* A boot block built against this interface would call `@(0x4000 + 4n)` instead of absolute
  addresses. The application could then be re-flashed without touching the vector block.
* In 2.30 the boot area was linked together with the application, so the wrappers use direct
  calls and the table is vestigial.
* Entry 10 (`boot_wait_key`) is the only entry point reachable *only* through this table.

## 3. Startup

`reset` (0x400) runs:

```
SP = 0xFFEFB0; push er6; ccr.I = 1; jsr main; sleep; ...
```

The decompiler shows the pushes as stores to 0xFFEFAC (ER6) and 0xFFEFA8 (return address 0x412).

`main` (0x13940) runs these steps in order:

1. `save_return_addr` (0x1F644) stores main's return address in 0xFFEFBA. That is the top 16
   bytes of RAM, above the initial SP. **Nothing in the firmware reads it back.** It records
   which path (reset or a loader) entered the application, for a debugger or loader to inspect.
2. `hw_init_bus_ports(0)` (0x739C) programs the system and bus control registers:
   * SCKCR 0xFFFDE6, SYSCR 0xFFFDE5, SBYCR 0xFFFDE4, LPWRCR 0xFFFDEC;
   * module-stop registers 0xFFFDE8–EA (all modules stopped first);
   * PFCR 0xFFFDEB, bus width and wait states 0xFFFED0–ED5, chip-select set-up 0xFFFEC0–ECC;
   * port DDRs 0xFFFE30–E3F and pin functions 0xFFFE12–E15.

   With `keep_port == 0`, 0xFFFE14 = 4; otherwise 0.
3. `_INITSCT` (0x238EC): C runtime section init. It copies 0x29644–0x298E0 to 0xFFC230 and
   clears 0xFFB048–0xFFC230.
4. `nop_1f660` and `record_stack_ptr` (0x1F656): stores SP in 0xFFBF9C, which service command
   `y` reports.
5. `tick_timer_init`: TMR0 1 ms tick.
6. `eeprom_check_format`: checks the 93Cxx format marker and writes defaults when it is missing.
7. `workbuf_init`, `rx_init`, `usb_tx_idle`, `sci0_init`.
8. `usb_suspended` (in fact the soft-power-off flag) is set to 0, and `cur_pen` to 1.
9. `motion_init`, `settings_apply`, `solenoid_init`, `pen_reset`, `select_pen_params`,
   `sensor_init`.
10. If jumper bits 2..5 of port 0xFFFFB2 are set, it runs `factory_test_select`.
11. `power_off(0)`. The plotter comes up in the soft-off state. The power key (IRQ2) switches it
    on and enters `gpgl_loop` through the warm-restart path (0x137C4). A long press starts
    `demo_run`.

## 4. `boot_wait_key` (0x137EC)

This is an alternative application entry with the same early init as `main`. The differences
are:

* `hw_init_bus_ports(1)` keeps 0xFFFE14 = 0 during the wait. It also clears port 0xFFFF09
  (LEDs off).
* It does no `workbuf_init`, `rx_init`, USB or SCI setup before the wait.
* **It waits for the panel key to be released.** It polls port 0xFFFFBE bit 0 (1 = key
  pressed, the same polarity `power_key_isr` uses) once per 1 ms tick, until the key has read
  released for 100 consecutive ticks. A press restarts the count.
* It then watches the key for another 1000 ticks (1 s) and records in 0xFFB39A whether it was
  pressed at any time.
* Then it does the full init:
  * 0xFFFE14 = 4;
  * `workbuf_init`, `rx_init`;
  * USB set-up done inline (instead of `usb_tx_idle`), identical to the power-up sequence in
    `power_key_isr`: USB interrupt-enable bit 0 (0xC000C4) cleared, shadow
    enables |= 0x0A, configuration pending, MSTPCRB bit 0 cleared (module enabled), USB
    register 0xC00086 = 0x36;
  * `sci0_init`, power flag cleared, `motion_init`, `settings_apply`, `solenoid_init`,
    `pen_reset`, `select_pen_params`, `sensor_init`.
* It **homes the carriage** (`home_carriage`).
* If the key was pressed during that second, it runs `demo_run` (self test plus the ROM GP-GL
  job). Otherwise it runs the endless `test_motion_loop`.
* It never reaches `gpgl_loop` or `power_off`.

This is a production or test start: something outside the application (a boot loader or a
programming fixture) calls it right after flashing to exercise the mechanics without a host.

## 5. Flash programming and update

* **No flash-programming code exists in the image.**
  * The flash control registers are never accessed: FLMCR1/2 and EBR1/2 at 0xFFFFA8–AB, or at
    0xFFFFC8–CB on related parts.
  * No code is copied to RAM and executed: no `jsr @ern` into RAM, and the only RAM copy is the
    `_INITSCT` data image.
* **The USB interface has no update hook.**
  * The only vendor request is request 0, which sets `jog_keys` (`usb_vendor_request` 0xFDF6).
  * The class requests are GET_DEVICE_ID, GET_PORT_STATUS and SOFT_RESET.
  * SET_DESCRIPTOR data is buffered and ignored.
* **The GP-GL interface has no update hook.** `FP` is only the service unlock (`FPGRFCC1`).
* **`ESC 0x01` (USB)** executes `trapa #1`. Vector 9 points to an `rte` stub (0x522), so it does
  nothing unless an emulator or debugger owns the vector. It is a debug breakpoint, not an
  update entry.
* **Mode pins:** MDCR (0xFFFDE7) is never read. The firmware does not care which MCU mode it
  booted in.

Conclusion: a firmware update rewrites the whole flash from outside the application, starting
at address 0. The most likely mechanism is the H8S on-chip **boot mode**:

1. The mode pins select boot mode at reset.
2. The mask-ROM boot program receives a writer program over SCI.
3. The writer erases the flash and programs the `.mot` image.

Graphtec's updater would drive this, either through the RS-232 port (SCI0, the same pins the
boot program uses) or through a service jig. The `.mot` entry point 0x400 equals the reset
vector. The application entry table and the empty boot area show that a field-update loader
was planned, but 2.30 does not contain one. Updating over USB alone is not possible with this
image.

## 6. Orphan code

These two routines have no callers. The Hitachi linker keeps unreferenced functions, so they
are leftovers.

**`motion_clear_counters`** (0x1AF28, 60 bytes) zeroes the motor position and counter longs:

* `motor_pos_x` (0xFFBF26), `motor_pos_y` (0xFFBF6A);
* 0xFFB5E0/E4 and 0xFFB600/04/08.

`motion_init` (0x1AF64, which immediately follows it) clears the same variables inline, so this
function was superseded.

**`hypot_approx`** (0x1BB22, 132 bytes) is a fast integer approximation of √(a² + b²):

* Let M = max(a, b) and m = min(a, b). If M == 0 it returns 0.
* t = (m << 8) / M, using `__divmodsi4`; t ranges 0..256.
* q = (t × (365·t + 0x3B00)) >> 16, then result = (M × (q + 1024)) >> 10.
* That is, M·(1 + 0.0576·(m/M) + 0.3564·(m/M)²). It gives exactly M at m = 0 and 1.414·M at
  m = M.
* It assumes non-negative inputs; there is no `abs`.
* It sits just before `atan_tenths` (0x1BBA6). The firmware uses `hypot_l` (0x10180) and
  `isqrt32`/`__xd_sqrt` instead.

## 7. Side note: IEEE 1284 ID length

`usb_class_request` sends GET_DEVICE_ID from the pointer at 0xFFC2DE (0x28C98) with the length
at 0xFFC2DC, which is **0x53 = 83** in the initialised-data image. That covers the 2-byte prefix
plus all 81 text bytes, so the whole string is transmitted. Only the embedded big-endian length
prefix (0x004D = 77) is wrong. A host that trusts the prefix truncates the description. This
corrects the note in `../assets/README.md`.
