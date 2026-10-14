# CC100 firmware 2.30: device-specific commands (F*, T*, TB, service mode)

Source: Ghidra decompilation of `cc100ver230.bin` (H8S, advanced mode). Addresses are ROM
addresses of the handler functions. Names match `sym/device.txt`, and readable C is in `src/device.c`.

## Conventions used below

* **Numbers.** `read_int()` (0x11b72) reads a decimal number with an optional `-`
  and decimals, and rounds it to an integer. `read_coord()` (0x11bc6) reads a coordinate
  in user units and converts it to internal units. The internal unit is 1/20 of the
  default user step, i.e. **0.005 mm** when the step is 0.05 mm. 1 mm = 200 internal units.
* **Optional parameters.** `more_params()` (0x0cb3a, returns 0 when another number follows) decides
  whether an optional parameter is present.
* **Replies** go through a small output buffer at 0xFFB2F3:
  * `reply_begin()` waits until the previous reply has been sent, then clears the buffer.
  * `reply_putnum(width, v)` writes `v` right-justified in a `width`-character field,
    padded with spaces. A negative number gets a `-` directly before the digits. A positive
    non-zero number always has at least one leading space, so a 5-digit value in width 5 is
    printed as 6 characters.
  * `reply_num_comma(w, v)` writes the number followed by `,`.
  * `reply_num_end(w, v)` writes the number, then terminator 1 (0xFFB359), then terminator 2
    (0xFFB35A) if it differs, then sends.
  * The terminators are set by the GP-GL `=` command. Their power-on defaults are handled
    elsewhere; they are normally ETX (0x03).
  * Example: `FA` with stored values 0 and -5 answers `    0,   -5` + terminator.
* **EEPROM.** Persistent settings live in a serial EEPROM, accessed through
  `eeprom_read_byte/word(0x300000+off)` and `eeprom_write_byte/word(v, 0x300000+off)`.
  Byte writes are skipped when the value is unchanged.

### EEPROM map (offsets)

| off       | size | meaning | written by | factory default (svc `g17`) |
|-----------|------|---------|------------|-----------------------------|
| 0x02+2n   | word | service table (n=0..9), unknown meaning | svc `g11` | – |
| 0x04      | word | config bit word (copied to 0xFFB368), bits 1,6,7,4 and field 2..3 | svc `g0,g1,g3,g5,g6` | – |
| 0x13,0x14 | byte | unknown pair | svc `g2` | – |
| 0x19      | byte | sensor/ADC calibration result | svc `g16` | – |
| 0x1a      | word | unknown | svc `g10` | – |
| 0x1c      | word | distance correction A, -200..200 | `FB` | 0 |
| 0x1e      | word | distance correction B, -200..200 | `FB` | 0 |
| 0x20,0x22 | word | mark-sensor offset X/Y, ×20 internal units | svc `g14` | – |
| 0x24,0x26 | word | registration calibration X/Y, -40..40, ×10 internal units | `TB72` | 0 |
| 0x28      | byte | reg-mark parameter (×2 into 0xFFC11A) | svc `g12` | – |
| 0x29      | byte | two nibbles (0xFFC124 / 0xFFC126) | svc `g13` | – |
| 0x2a      | byte | unknown | svc `g4` | – |
| 0x2b      | byte | unknown | – | 1 |
| 0x2c      | byte | 0/1 flag | svc `g7` | – |
| 0x34      | byte | speed (1..10), reported by `FQ0` | – | 10 |
| 0x35      | byte | acceleration (1..3), reported by `FQ1` | – | 3 |
| 0x36      | byte | value 0..45, reported by `FQ2` | – | 0 |
| 0x37      | byte | orientation (0 portrait, 1 landscape), `FQ4` | `FN` | 0 |
| 0x38      | byte | reported by `FQ3` | – | 0 |
| 0x39      | byte | track-enhancing pass count | svc `g18` | – |
| 0x3a      | byte | media type 1..13 | `FW` | 1 |
| 0x3b      | byte | cutting force 1..30 | `FX` | 0 |
| 0x3c      | byte | track enhancing disabled flag | `FY` | 0 |

## F* commands

The dispatcher consumes both command letters. Unless noted, parameters follow directly,
separated by `,`, and the command is ended by the terminator or by the next command letter.

### FG — firmware version (0x0ff4e)
* Syntax: `FG`
* Reply: the ROM string at 0x28CEC, `V2.30` followed by four spaces (`"V2.30    "`),
  then terminator 1 (and terminator 2 if different).
* Confirms `Commands.md` / `cc100.txt` ("Version").

### FA — read distance correction (0x0767a)
* Syntax: `FA`
* Reply: `%5d,%5d` + terminator, built from EEPROM words 0x1c and 0x1e.

### FB — set distance correction (0x076be)
* Syntax: `FBa,b`
* Both values are integers, clamped to -200..200 (cc100.txt says -200..200 = ±2 %),
  and written to EEPROM 0x1c / 0x1e.
* Effect: reloads the correction (0x12040) and recomputes the scaling (0x18f0e). The current
  logical position (0xFFB4F8/FC) is converted to device coordinates and back, so the pen
  keeps its logical place.
* `cc100.txt` writes `FBn,0`, so the second value is a separate correction (probably the other axis).

### FJ — re-reference origin (0x07832)
* Syntax: `FJ`
* Effect: waits for motion to finish, clears the registration transform, and zeroes the
  offsets 0xFFB4F0..FC. It then computes the soft origin 0xFFB51C/0xFFB520 from the
  device extents 0xFFB514/18 and the converted current position (0xFFB544/48).
  The position is re-applied via 0x151be.
* This matches "Set Current Origin" in cc100.txt. TODO: exact semantics of 0x151be.

### FL — query media mode (0x07a3a)
* Syntax: `FL`
* Reply: `%6d` + terminator of 0xFFB3B8 (`     0` or `     1`).
* `cc100.txt` "returns 1" matches FM1 being the normal setting.

### FM — set media mode (0x07a64 / set_media_mode 0x07aa4)
* Syntax: `FMn`. n ≠ 0 selects mode 1. Ignored while the registration transform is active.
* Effects:
  * Mode 0: Y origin offset 1000 (5 mm), usable width 38000 (190 mm).
  * Mode 1: offset 0, width 40000 (200 mm).
* The current position is shifted by ±1000 so it stays on the same paper location.
  Motion limits are recomputed (0x14ff4, 0x17e8e, 0x150dc).
* Interpretation: mode 1 = cutting mat / carrier sheet, mode 0 = paper without mat. Not stored in EEPROM.

### FN — orientation (0x201f8)
* Syntax: `FN[n]`. A missing n means 0. 0 = portrait, 1 = landscape.
* Effect: the value is stored in EEPROM 0x37 and applied by `set_orientation()` (0x20268).
  That call:
  * removes any registration offset;
  * swaps axes (0xFFB400) and resets offsets;
  * moves pen-up to 0,0 at speed 3000 (0x19780 / 0x1c678).
* Confirms cc100.txt "Orientation".

### FO — page length (0x07c82)
* Syntax: `FO[len]`. `len` is a coordinate (user units). 0 or missing means 200000 internal (1000 mm).
* Without registration marks: sets the page length 0xFFB50C.
* With an active registration transform: only accepted for manual mark search in landscape.
  The length then gets the mark origin 0xFFB3FC added.
* Copies limits to 0xFFB52C and sets the "page set" flag 0xFFB3BA.
* The test pattern uses `FO5440` (272 mm) and resets with `FO0`.

### FP — service unlock (0x0d46c)
* Syntax: `FPGRFCC1` (six literal characters after `FP`).
* Effect: if all characters match, it sets 0xFFB22E = 1, which enables the lowercase
  service commands (see below). Any mismatch stops at the first wrong character, and that
  character is consumed.
* `TT`/reset (0x137c4) clears the flag again.
* Contradiction: cc100.txt guesses "Program/Update Firmware". It is a service-mode unlock instead.

### FQ — query stored parameters (0x07d46)
* Syntax: `FQn`. Reply `%5d` + terminator.

| n | value | clamping |
|---|-------|----------|
| 0 | EEPROM 0x34 (speed) | 0 or >10 → 10 |
| 1 | EEPROM 0x35 (acceleration) | 0 or >3 → 3 |
| 2 | EEPROM 0x36 | >45 → 0 |
| 3 | EEPROM 0x38 | none |
| 4 | EEPROM 0x37 (orientation) | none |
| 5 | 0xFFB0AA, 1 if registration marks were found and the transform is active | none |

* Other values of n give no reply.
* cc100.txt: "FQ5 emitted after auto calibration". This matches: the host asks whether
  mark detection succeeded.

### FT — move to origin (0x148ce)
* Syntax: `FT`. Effect: `plot_to(0, 0, pen up)` (0x1db80).

### FU — track-enhancing feed (0x148e8)
* Syntax: `FUlen,w`. Two coordinates. Only the first, the length along feed, is used.
* Acts only when **all** of these hold:
  * The force is ≥ 19, or (force unset and media table entry == 3).
  * Track enhancing is not disabled (`FY`, 0xFFB3AF ≠ 1).
  * The pass count in EEPROM 0x39 is non-zero.
* It then feeds the media back and forth over `len`, pass-count times, and returns to the start.
* Registration-aware variant: uses the stored corners 0xFFBC20..0xFFBC34.
* cc100.txt guesses "Cutting Area". Hosts send the page size here (`FU5440,4000`), so it
  doubles as the area hint, but the firmware only uses it for the track-enhancing roll.

### FV — query media type (0x07ec8)
* Syntax: `FV`. Reply `%5d` + terminator of EEPROM 0x3a. 0 or >13 is reported as 13.
* cc100.txt says "returns 13" (the unset default).

### FW — set media type (0x07f1a)
* Syntax: `FWn`. n outside 1..13 becomes 13. Stored in EEPROM 0x3a and in 0xFFB4C8.

### FX — cutting force (0x07f7a)
* Syntax: `FXn[,m]`. n is clamped to 1..30 and stored in 0xFFB122 and EEPROM 0x3b.
  The second number is read and ignored.
* Effect: sets the "force changed" flag (0xFFB2BE), selects pen 1 (0x17992), and
  re-applies the force (0x18020 / 0x180b2).

### FY — track enhancing (0x14b12)
* Syntax: `FYn`. 0xFFB3AF = (n ≠ 0), stored in EEPROM 0x3c.
* `FY1` **disables** track enhancing (`FU` skips), and `FY0` enables it. This matches
  the Silhouette driver usage ("Track Enhancement").

### FH — test cut (0x141d2)
* Syntax: `FH`
* Draws a 10 mm × 10 mm square (2000 × 2000 internal) at the current position, with a
  triangle inside: (200,400) → (1800,400) → (1000,1800) → (200,400). This is
  `draw_test_cut_mark` (0x14068).
* Then moves pen-up 2100 units (10.5 mm) further along feed and lifts the tool.

### FI — endless test plot (0x14246)
* Syntax: `FI`
* **Never returns.** It plots a burn-in pattern in an infinite loop until power-off:
  * 91 radial line pairs computed with soft-float sin/cos;
  * two zig-zag polylines from table 0x28D6C / 0x28D74;
  * a 12-segment polyline from table 0x28D82;
  * circles (0x146ee ×2, 0x1476a);
  * six horizontal 30 mm lines.
* Contradiction: cc100.txt "Test Cut 2". It is an endless test plot, not a single test cut.

### TT — reset (0x08002)
* Syntax: `TT`
* Effect, in order:
  * waits for motion to finish and lifts the tool (0x1d6f8);
  * selects media mode (0x07aa4);
  * runs 0x13436;
  * clears the 8 per-pen "pen down" flags (0xFFB446..);
  * re-initialises the plot state (0x1054c), the interface (0x14ef0) and the defaults (0x104e6).
* Commonly called "Home/CR" in cc100.txt. It is a soft reset of plotting state.

## TB — registration marks (dispatcher 0x1f6cc)

* Syntax: `TBn[,args]`. n is read as an integer. Values with a non-zero high byte are ignored.
* Configuration word 0xFFC100 (bit 15 = MSB):

| bits  | meaning |
|-------|---------|
| 15    | apply registration transform after search (`TB99` sets, `TB0` clears) |
| 14    | `TB55` flag |
| 13..12| search type, 1 = TB22/TB122, 2 = TB23/TB123 (bits 5..4 of byte 0xFFC100) |
| 11..10| mark type, 1 = `TB52,1`, 2 = otherwise (selects detector 0x15212 vs 0x154ca) |

* Defaults set by `regmark_init` (0x20616):
  * type field = 2;
  * search type 0; flags 0;
  * mark length 10 mm;
  * offsets 0;
  * calibration loaded from EEPROM.

| n | handler | syntax | effect |
|---|---------|--------|--------|
| 0 | 0x1f930 | `TB0` | clear bit 15 (registration off) |
| 22, 23 | 0x1fb80 | `TB22[,x,y]`, `TB23[,x,y]` | set search type (1/2); optional mark origin x,y (coords) → 0xFFC112/16; then **manual** search (0x1575a): the user has positioned the tool at the first mark |
| 122, 123 | 0x1fbe6 | `TB122[,x,y]`, `TB123[,x,y]` | same, but **automatic** search (0x15fb2) |
| 50 | 0x1f8aa | `TB50[,n]` | orientation for registration: n≠0 = landscape; swaps axes keeping the position (0x1f79c). Ignored while the transform is active |
| 51 | 0x1f956 | `TB51,len` | mark length; coordinate/200 (mm); accepted 5..20 mm → 0xFFC223. Silhouette `TB51,400` (400 × 0.05 mm = 20 mm) is accepted |
| 52 | 0x1f9a0 | `TB52[,n]` | mark type; n == 1 (or missing) → type 1, else type 2 |
| 53 | 0x1f9ee | `TB53,w` | mark line width; low word of the coordinate, accepted 60..200 internal (0.3..1.0 mm) → 0xFFC224. The value is **never read** elsewhere in 2.30 |
| 54 | 0x1fa30 | `TB54[,x[,y]]` | mark offset → 0xFFC102/0xFFC106 (coordinates) |
| 55 | 0x1fa6c | `TB55,n` | bit 14 = (n≠0). cc100.txt "type 2 reg marks" |
| 70 | 0x1fa9c | `TB70` | calibration run (0x1fc4c, below) |
| 71 | 0x1fab4 | `TB71` | reply `%5d,%5d` + terminator: EEPROM 0x24 / 0x26 (calibration X/Y) |
| 72 | 0x1faf8 | `TB72,x,y` | clamp ±40, write EEPROM 0x24 / 0x26, reload offsets (0x206d6). cc100.txt "40,40..-40,-40" ✓ |
| 98 | 0x1f90a | `TB98` | reply `     1` + terminator (constant 1, width 6). Probably a "registration supported" probe |
| 99 | 0x1f930 | `TB99` | set bit 15 (apply registration transform) |

**Search result reply.** Both manual and automatic search finish with `reply_num_end(5, result)`:
* `    0` + terminator: marks found. 0xFFB0AA is set (see `FQ5`), and the affine transform is
  built by 0x54e4 when bit 15 is set.
* Non-zero: marks not found. The tool returns to the start, and the page reverts to
  1000 mm × 200 mm.

`TB23` in cc100.txt "Search Marks Manual" and `TB123` "Search Marks Automatic" are both confirmed.

**Calibration (`TB70`).** In portrait (landscape uses 0x1ff2a):
1. Moves to the stored mark position (0xFFB54C).
2. Searches the mark edges in four directions, 1600 units (8 mm) apart (0x2114c / 0x2135c).
3. Computes the centre (0x5ecc).
4. **Cuts a cross** through the found centre lines.
5. Moves back 30 mm.

Return: 0 on success, 1 if a search failed. There is no reply bytes: the host
reads the result with `TB71` after adjusting with `TB72`.

## Built-in demo / self test (0x0d512) and the ROM GP-GL program

* **Trigger.** The key interrupt (vector 18, handler 0x12ff2) calls `demo_run()` when the
  button is held long (more than about 51 polling rounds). The demo flag 0xFFC2DA is then
  set. From then on, `peek()` (0x11228) re-runs the demo every time the input stays idle
  for about 100 polls.
* **Sequence:**
  1. Runs a mechanical self test: jogs ±10/±3 steps on both axes at different speeds
     (0x19780 / 0x0d4c8), then large moves.
  2. Points the GP-GL input pointer (0xFFB2E2) at ROM 0xDB9A with length
     0xF096 - 0xDB9A = 5372 bytes (0xFFB2E6).
  3. Calls the normal interpreter `gpgl_loop()`.
* **The ROM program** (plain GP-GL, no terminators between commands):
  * starts with `FU5440,4000` `FM1` `TB50,0` `FO5440` `&100,100,100,` `\0,0,`
    `Z5940,4200,` `L100,1,40,80,` `^0,100`;
  * then `M`/`D` polylines alternating between `L100,1,40,80,` (dashed fold lines) and `L0,`
    (solid cut lines);
  * ends with `&1,1,1,TB50,0FO0`.
* **What it draws.** A papercraft box / package net about 245 × 157 mm. It has a central
  body with fold flaps, rounded tuck tabs, and a hexagonal side flap strip, with dashed
  perforation lines on the folds. A rendering is in the scratchpad as `tp.png`.
* The program also shows real usage of the custom line type `L100,1,a,b` (dash a, gap b)
  and of `FU`, `FM`, `TB50`, `FO`.

## Lowercase service commands (only when 0xFFB22E ≠ 0, i.e. after `FPGRFCC1`)

These are single lowercase letters handled first in `gpgl_loop`. Parameters are integers
via `read_int()`. "Query/set" commands share a pattern: `xN,0` replies the current value
as `%5d`+terminator, and `xN,1,v` writes it.

| cmd | handler | syntax / effect |
|-----|---------|-----------------|
| `a` | 0x21b22 | `a0`: set port bit 0xFFFF06.4. `a1,n`: ADC/sensor routine 0x219fa(n). `a2`: sensor read 0x21c40 once. `a3`: 99 times |
| `b` | inline  | reply `%5d` of input port 0xFFFFB6 & 8 (switch/sensor state) |
| `c` | 0x19758 | `cN` → 0xFFB620 (motion parameter) |
| `d` | inline  | `dN`, N=0..3 → step size table {20,10,5,2} → 0xFFB262 (user unit 0.1/0.05/0.025/0.01 mm), recompute scaling |
| `e` | 0x101fe | `ea,b,c,d`: arithmetic self test of 64-bit compare 0x1631a; replies three `%10d` numbers |
| `f` | inline  | `fADDR`: reply `%5d` of the memory byte at ADDR (peek) |
| `g` | 0x16bc0 | `gN,0` query / `gN,1,v` set parameter N (table below) |
| `h` | 0x19764 | `hN`: recompute the acceleration/speed profile tables (0xFFB640..) with soft-float math |
| `i` | 0x174fc | feed media +8000 (40 mm) and back to 0, set origin |
| `j` | 0x12ad0 | reply `%6d,%6d,%6d` of 0xFFB38A/8E/92 (sensor or counters) |
| `k` | inline  | wait for motion; reply `%6d,%8d` of 0xFFBF26 and 0xFFBF6A |
| `l` | inline  | `lN` → 0xFFB396 |
| `m` | 0x17556 | endless burn-in: 10 × diagonal moves (52000,38000)↔(0,0) pen down, then 100 pen up/down cycles, forever |
| `n` | 0x21af8 | `nP`: sensor threshold = P·1024/100 → 0xFFC22C |
| `o` | 0x0aab2 | starts subsystem 0xFFB178 (TODO) |
| `p` | inline  | toggle 0xFFC2D8 and write 0x0F / 0x0A to latch 0x600010 (front-panel LEDs) |
| `q` | inline  | endless pen/solenoid up/down every 500 ticks |
| `t` | 0x21b96 | `t0..t4`: registration-mark sensor tests (`t1,a,b` edge search 0x20d6a, `t2,n` 0x2114c) |
| `y` | inline  | reply `%8d` of 0xFFBF9C |
| `z` | 0x1799e | `z1,v` / `z2,v` / …: set or clear individual output port bits (0xFFFF0E) |

`g` parameters (handler table at 0x28F14):

| N | target |
|---|--------|
| 0, 1, 3, 5 | config word bits 1, 6, 7, 4 of 0xFFB368, written to EEPROM 0x04 via 0x12302 |
| 2 | EEPROM 0x13 / 0x14 pair |
| 4 | EEPROM 0x2a |
| 6 | config word field bits 3..2 |
| 7 | EEPROM 0x2c |
| 10 | EEPROM word 0x1a |
| 11 | `g11,i,…`, word table i=0..9 (read 0x08+2i, write 0x02+2i) |
| 12 | EEPROM 0x28, then re-init reg-marks |
| 13 | EEPROM 0x29 (two nibbles) |
| 14 | EEPROM words 0x20 / 0x22 |
| 15 | run `TB70` calibration |
| 16 | sensor auto-calibration (0x21a26), result in EEPROM 0x19; reply `%5d` |
| 17 | **factory reset** of user EEPROM settings (0x11f7e, defaults in the EEPROM table) |
| 18 | EEPROM 0x39 (track-enhancing pass count) |

## Cross-check against Commands.md / cc100.txt

| Item | Status |
|------|--------|
| FG Version | confirmed |
| FA/FB distance correction ±200 | confirmed (two values) |
| FH test cut | confirmed |
| FI "test cut 2" | it is an endless test plot |
| FJ set current origin | roughly confirmed |
| FL "returns 1" | confirmed (media mode, default 1) |
| FM1 | media mode 1 (200 mm, no Y offset) |
| FN orientation | confirmed, persisted |
| FP "firmware update?" | **wrong**: service unlock `FPGRFCC1` |
| FQ5 after calibration | confirmed (registration success flag); FQ0..4 also exist |
| FU cutting area? | used only for the track-enhancing roll |
| FV "returns 13" | default media type |
| FW media type | confirmed, persisted |
| FX force 1..30 | confirmed, persisted |
| FY track enhancement | confirmed, FY1 = off |
| TT Home/CR | is a soft reset |
| TB50,0/1 | landscape/portrait for registration (0 = portrait) |
| TB51 mark length | confirmed |
| TB55,1 | flag bit only |
| TB70/71/72 calibrate/get/set | confirmed (±40) |
| TB99 | **sets** registration transform enable; `TB0` clears it |
| TB23 / TB123 manual / automatic search | confirmed; TB22 / TB122 are the type-1 variants |

Not covered here, handled by other notes: FC, FD, FE, FF, FK, FR, FS, SO/NP, ZP.
