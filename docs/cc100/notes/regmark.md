# CC100 fw 2.30: registration marks (TB), end to end

The readable C is in `../src/regmark.c`, and new RAM names are in `../sym/regmark.txt`.

* Units are internal device units: 0.005 mm, 200 per mm.
* "X" is the logical X axis, along the media feed in portrait. "Y" runs across the carriage.
* In landscape (`TB50,1`) the scans swap motor axes. Otherwise the procedure is the same.

## 1. Sensor

### Hardware

* A reflective optical sensor sits next to the knife.
* Its LED is on port 0xFFFF06 bit 4.
* The photo signal is read on A/D channel 0, giving a 10-bit value; paper reads high.
* The LED drive is a serial digital pot: 10-bit word, MSB first, on 0xFFFF00 (bit 1 data, bit 2 clock, bit 3 /CS).
* `sensor_pot_step(level)` (0x219fa) sets level 0..255. It is misnamed: it sets a level rather than stepping.

### Gain auto-calibration (service `g16`, 0x21a26)

* The level is raised from 0 until one A/D conversion exceeds `sensor_threshold`.
* The default threshold is 0x200 (50 %); service `n<percent>` sets it as percent × 1024 / 100.
* The result is stored in EEPROM 0x19 and reloaded by `sensor_init`.

### Edge detection

An edge means the live A/D value has dropped below a threshold.

| Scan kind | Threshold |
|---|---|
| Fixed ("precise") | `baseline × gain / 10`. `baseline` is the reading at scan start. `gain` is a nibble from EEPROM 0x29 (low nibble X, high nibble Y; service `g13`), e.g. 7 → 70 % of the paper level. |
| Adaptive ("coarse"), `sensor_filter_*` 0x207d0 / 0x208d0 | `gain/10 ×` the average of 16 samples taken about N samples earlier, with N = min(1000 / scan_speed, 105), using a 110-entry ring. |

The adaptive threshold tracks slow changes in paper brightness and only reacts to a sharp drop.

### Sensor-to-knife offset

Defined by `scan_off_x`/`scan_off_y` (`regmark_load_calibration`, 0x206d6):

* nominal: (0, −15 mm);
* minus 0.1 mm × EEPROM 0x20/0x22 (factory, service `g14`);
* minus 0.05 mm × EEPROM 0x24/0x26 (user, `TB72`).

While `regmark_scan_mode` is set, every queued move is shifted by this offset (`motion_line` 0x1d4aa, `device_to_logical` 0x18ae0). Coordinates during a search therefore name the point under the sensor.

## 2. Scan primitives

### `sensor_scan_axis` (0x20d6a)

* Drives one motor directly (`motor_run_continuous`), ramping the speed 1 step per sample.
* It stops on any of:
  * an edge (success);
  * `ESC 03` / `ESC 02` cancel (fail);
  * in coarse mode, the travel limits (fail);
  * in precise mode, 20 mm travelled (fail).
* When it stops it ramps down and reports the position at the moment the edge was seen.

### `regmark_locate_x(dir)` / `regmark_locate_y(dir)` (0x2114c / 0x2135c)

Finds the centre line of one printed bar:

1. Coarse scan in `dir` until the bar is hit (E1).
2. Back off 2 mm, precise scan forward to the leading edge (E2).
3. Go 3 mm past E1, precise scan back to the trailing edge (E3).
4. The result is the midpoint of E2 and E3. Its error does not depend on the scan speed or edge sharpness.

The bar must be narrower than about 3 mm. `TB53` (0.3..1.0 mm) is stored but never used.

### `sensor_scan_diag` (0x21620)

Both motors run at once. It is used only to find mark 1 automatically.

## 3. Mark shapes

A mark is an **L** of two bars, each `regmark_length` mm long:

* `TB51`, accepted 5..20 mm, default 10;
* width 0.3..1 mm.

The firmware samples each bar once:

| Sample | Bar |
|---|---|
| P0 | bar parallel to Y of mark 1 |
| P1 | bar parallel to X of mark 1 |
| P2 | bar parallel to Y of mark 2 |
| P3 | bar parallel to X of mark 2 |
| P4 | bar parallel to X of mark 3 |
| P5 | bar parallel to Y of mark 3 |

* Mark 1 is the page origin. Mark 2 lies along +X (feed direction), mark 3 along +Y.
* **Type 2** (default, `TB52,2` or any value ≠ 1): inward corners. Each L's arms point into the page.
  * Mark 1 is found behind the sensor (−X, then −Y).
  * Mark 2 is found scanning +X; its X-bar extends towards mark 1.
  * Mark 3 is found scanning +Y; its Y-bar extends towards mark 1.
* **Type 1** (`TB52,1`): every scan runs in the + direction, so the arms point outward.
  * Mark 1's X-bar extends −X from its Y-bar.
  * Mark 2's X-bar extends +X.
  * Mark 3's Y-bar extends +Y.

The shapes are inferred from the scan directions and half-arm offsets.

### Scan path, type 2 (`regmark_find_type2`, 0x154ca)

1. From the start, `locate_x(−1)` gives P0.
2. Move +L/2 in X, then `locate_y(−1)` gives P1.
3. Move to P0.x + span, then `locate_x(+1)` gives P2.
4. Move −L/2 in X, then `locate_y(−1)` gives P3.
5. Stop here for a 2-mark search.
6. Otherwise move to (P0.x + L/2, P1.y + span), then `locate_y(+1)` gives P4.
7. Move −L/2 in Y, then `locate_x(−1)` gives P5.

Type 1 is analogous, with all scans + and the half-arm offsets mirrored.

### Search span

The span is where the search for mark 2 or 3 starts (`regmark_compute_spans`, 0x15a1a):

* If the nominal distance `d` was given (`TB23,dx,dy` / `TB123,dx,dy`), the span is d − 10 mm.
* Otherwise it is L/2 for type 2 or 0 for type 1. The coarse scan then runs on until the next bar.

The coarse X scan may run up to 1100 mm past the feed origin. The Y scan runs between 4 mm and 200 mm of the sensor range.

## 4. Corners and transform

### Corners (`regmark_find_corners`, 0x46b8)

* **2 marks:**
  * The X line runs through P1 and P3.
  * The mark 1 corner is P0 dropped perpendicularly onto it; the mark 2 corner is P2 dropped onto it (`regmark_axis_intersect`, 0x5082).
* **3 marks:**
  * The mark 1 corner is the intersection of the X line (P1, P3) and the Y line (P0, P5) (`regmark_line_intersect`, 0x48ba).
  * Mark 2 is P2 projected onto the X line.
  * Mark 3 is P4 projected onto the Y line.

### Transform (`regmark_compute_transform`, 0x54e4)

```
ux = (m2 - m1) / |m2 - m1|            must point +X, within +-5 deg (|dy/dx| <= tan 5)
2 marks: uy = ux turned by +90 deg
3 marks: uy = (m3 - m1) / |m3 - m1|   must point +Y, within +-5 deg   (keeps skew)
x' = rm_xx*x + rm_xy*y   (rm_xx = ux.x, rm_xy = uy.x)
y' = rm_yy*y + rm_yx*x   (rm_yx = ux.y, rm_yy = uy.y)
origin = m1 + transform(TB54 offset)
```

Any failed check or degenerate geometry sets result 1.

### Scaling (`TB55,1`, 0x5bc0 / 0x5d40)

* With a nominal distance given, scale = nominal / measured, accepted 0.6667..2.0.
* Points are multiplied by the scale before the rotation.
* Without `TB55`, or outside that range, the scale is 1.

### Clipping area

The area is the rectangle (0, 0)..(X span, Y span), minus the `TB54` offset.

* X span: the nominal distance (when scaling), otherwise measured + 3 mm.
* Y span: for 3 marks, as X. For 2 marks, Y is unlimited (1000 mm).

`plot_device` (0x1c572) clips every segment against this area with Cohen–Sutherland (0x1c146) before `regmark_transform` is applied (0x1c918). Nothing is cut outside the registered print.

## 5. Host protocol

### Typical Graphtec/Silhouette sequence

`FQ5` reads the result flag (`regmark_active`) at any time.

```
TB50,0          orientation (0 portrait, 1 landscape); ignored while registered
TB52,2          mark type (2 = default inward L)
TB51,400        arm length in user units (400 x 0.05 mm = 20 mm)
TB53,10         line width (stored only)
TB54,x,y        optional: page origin relative to mark 1
TB55,1          optional: scale to the nominal distances
TB99            build the transform after the search (else only marks are found)
TB123,dx,dy     automatic search, 3 marks, nominal distances (or TB23: manual)
-> "    0" ETX  found (non-zero: not found)
... job ...
TB0 / TT        (TB0 only stops building transforms; the active one stays)
```

### Search behaviour

* **Manual** (`TB22` / `TB23`, 0x1575a): the user has positioned the sensor on mark 1.
  * The feed origin is put 20 mm (+ L/2 for type 2, + any negative `TB54` offset) behind the sensor.
  * The page length becomes 1100 mm minus the travelled distance.
* **Automatic** (`TB122` / `TB123`, 0x15fb2): mark 1 is found by sweeping diagonals between the X and Y axes from the current position (0x15bbc).
  * The diagonals step outward 1.6 × L mm at a time, giving up at 100 mm.
  * Once a bar is hit, the sensor is placed on the mark and the search continues as in manual mode.
* **On success:** the knife moves to the registered (0, 0).
* **On failure:** the origin and page revert to the state before the search (1000 × 200 mm), the transform is cleared, and the reply is non-zero.

### Calibration (`TB70` / `TB71` / `TB72`)

1. The user puts the sensor on a printed "+" cross and sends `TB70`.
2. The firmware samples the vertical bar, the horizontal bar 8 mm to the side, the vertical bar 8 mm up, and the horizontal bar 8 mm to the other side. It uses a ±50 mm scan window (0x1fc4c portrait, 0x1ff2a landscape).
3. It cuts both lines, each lengthened 15 mm at both ends, so the knife cuts a "+" where the sensor saw the cross.
4. The knife then moves back 30 mm to Y = 0.
5. There is no reply. If a bar is not found, it returns to the start.
6. The user measures the offset between printed and cut cross and stores it with `TB72,x,y` (−40..40 × 0.05 mm, EEPROM 0x24/0x26). `TB71` reads it back.

### Other TB codes

* `TB98` always replies 1.
* `TB0` clears and `TB99` sets the "build transform" bit (bit 15 of `regmark_cfg`). `TB0` does not cancel an already active transform.

## 6. `regmark_cfg` (0xFFC100)

| bits | meaning | set by |
|------|---------|--------|
| 15 | build transform after search | `TB99` / `TB0` |
| 14 | scale to nominal distances | `TB55` |
| 13..12 | 1 = 2 marks, 2 = 3 marks | `TB22`/`122`, `TB23`/`123` |
| 11..10 | mark type 1 or 2 (default 2) | `TB52` |

## 7. Service commands (after `FPGRFCC1`)

| command | effect |
|---|---|
| `a0` | LED on |
| `a1,l` | set gain level |
| `a2` | reply the A/D value once |
| `a3` | reply the A/D value 99 times |
| `n<pct>` | threshold |
| `t0` | LED on, reply the A/D value |
| `t1,axis,speed` | one coarse scan |
| `t2,dir` | locate a bar along X |
| `t3,axis` | record a raw trace: A/D and distance; turns back after 800 samples or 30 mm; replies −1 first |
| `t4` | read the next trace pair (`%6d,%6d`), or `-1,-1` at the end |
| `g12` | coarse scan speed (EEPROM 0x28 × 2) |
| `g13` | edge gain nibbles |
| `g14` | factory sensor offset |
| `g16` | gain auto-calibration |

## 8. Firmware quirks

* `TB51` with no number compares an uninitialised local.
* `locate_x`/`locate_y` copy an uninitialised point on failure (the callers ignore it).
* `regmark_clear_transform` leaves rm_yy = 0 and rm_xy = 1. This only matters while the transform is inactive, and then it is not used.
* `regmark_clip_plot` plots the clipped end point twice.
* `regmark_set_orientation` swaps the stored origin offsets to keep the media position.
