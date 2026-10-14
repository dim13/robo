# GP-GL as implemented by Graphtec CC100 (Craft ROBO) firmware 2.30

This specification is derived from the decompiled firmware `cc100ver230.bin`. It describes
what the firmware actually does, not what the Graphtec GP-GL manual promises. Where the two
differ, the firmware wins, and the difference is called out. Handler addresses are given so
every statement can be checked against `decompiled.c` and `src/*.c`.

Notation:

* `[x]` is optional, `...` is repetition.
* `ETX` is byte 0x03 and `ESC` is byte 0x1B.
* *int* is an integer parameter (rounded, see 3.3).
* *coord* is a coordinate in user units (see 4).
* *angle* is degrees with up to two decimals.

## 1. Overview

* The CPU is a Renesas H8S running in advanced mode.
* The host interfaces are USB (an external USB controller at 0xC00000) and RS-232 (SCI0).
* Both interfaces feed one receive ring buffer, which a single interpreter loop
  (`gpgl_loop`, 0xCBF2) consumes byte by byte.
* There is no separate HP-GL mode.
* Only the plotting subset of GP-GL is implemented: lines, curves, circles, line types,
  windows, factor, offset and rotation.
* Device-specific `F*` and `TB*` commands are added on top.
* Text, hatching, marks, polar and status-word commands are accepted but do nothing.

## 2. Transport

### 2.1 USB

* Device: VID 0x0B4D, PID 0x110A, USB 1.0, self-powered, printer class (7/1/2, bidirectional), no string descriptors.
* Endpoints: bulk OUT EP1 and bulk IN EP 0x82, 64 bytes each.

* Host-to-plotter data arrives on the bulk OUT endpoint (handler 0x11556).
* Flow control: when fewer than 100 bytes are free in the receive buffer, the endpoint
  interrupt is disabled and the host is NAKed. It is re-enabled once more than 110 bytes
  are free.
* Replies are sent on bulk IN.

### 2.2 RS-232

* 9600 baud, 8 data bits, no parity, 2 stop bits. This is derived from the SCI0 setup and assumes a 16 MHz system clock.

* Bytes are received by the SCI0 RXI interrupt (0x1F4A4).
* There is no software flow control and no overflow check on the serial path.
* A reply goes back on the interface from which the most recent data byte arrived.

### 2.3 Real-time escape sequences

These are interpreted in the receive interrupt and never reach the command interpreter.
On USB, the `ESC` and the following byte must arrive in the same USB packet.

| Sequence | USB | RS-232 | Effect |
|----------|-----|--------|--------|
| `ESC 0x00 n` | yes | yes | Remote jog keys. `n` is an active-low bit mask: bit0 = +X, bit1 = −X, bit2 = +Y, bit3 = −Y. While the interpreter is idle and `n` ≠ 0, the head jogs in the selected directions. `n` = 0 ends jog mode. |
| `ESC 0x01` | yes | no | Debug trap (empty handler). Afterwards the byte 0x01 is stored as data. |
| `ESC 0x02` | yes | yes | USB: immediate reply `"1"` + ETX if the plotter is powered on, `"0"` + ETX if it is in the soft power-off state. RS-232: cancels the current job, like USB `ESC 0x03`. |
| `ESC 0x03` | yes | no | Cancels the running job. Long plotting loops poll a cancel flag and abort. |
| `ESC 0x04` | yes | no | Accepted and ignored. Hosts send it as "initialize device". |
| `ESC 0x05` | yes | no | Immediate status reply: `"0"` + ETX when the plotter is idle and waiting for data, `"1"` + ETX when it is busy. |
| `ESC x` (other) | yes | stuck | USB: ESC is dropped and `x` is stored as data. RS-232: the escape state machine waits until a 0x00 or 0x02 arrives and swallows everything until then (firmware bug). |

Inside the command stream, the interpreter additionally drops any `ESC` or `0x1C` byte
together with the byte that follows it.

## 3. Syntax

### 3.1 Commands

* A command is one character, or two characters for the forms listed in section 7.
* For a two-character command, the interpreter consumes the first character and peeks at
  the second. If the pair is a known two-character command, the second character is
  consumed too. Otherwise the single-character command runs, and the peeked character
  becomes the start of its parameters (or of the next command).
* Between commands the interpreter silently skips:
  * digits, `+`, `,`, `-`, `.`;
  * space and all control characters, including the terminator;
  * the filler byte `0x80`;
  * `#`.
* A byte above 0x5F (lowercase letters, `{|}~`, DEL, 8-bit bytes) sets error bit 1 and is
  ignored, unless service mode is on (see 8).
* An unknown command character also sets error bit 1 and is ignored.
* Because numbers and separators are skipped, the parameters of an unsupported or no-op
  command are discarded harmlessly. If a parameter list contains a command character,
  that character starts the next command.

### 3.2 Parameters, separators and terminator

* Parameters follow the command directly and are separated by `,`. Spaces and control
  characters between parameters are ignored too.
* A parameter list ends at:
  * the terminator;
  * the next command character (any printable character other than `+ - . 0-9 ,`);
  * `/`;
  * `ESC`, `0x1C` or `0x80`.
* The terminator is ETX (0x03). It is fixed: the `=` command is not implemented, so the
  terminator cannot be changed. A complete `=t1,t2` handler exists at 0x7636, but nothing
  calls it.
* The terminator is optional and never consumed by a command. `M100,100D200,200` is valid.
* Commands with a variable number of points (`D`, `E`, `Y`, `_`, `>`) keep reading points
  while another number follows.
* Commands with a fixed arity read exactly their parameters. Surplus numbers are skipped by
  the interpreter.
* A missing parameter reads as 0 and consumes nothing. For example `M,5` moves to (0,5).

### 3.3 Numbers (`read_number`, 0x11A36)

```
number  = { skip } [ "-" ] digits [ "." digits ]
```

* Before the digits, the parser consumes any characters that are not digits, `.`, the
  terminator or command characters. A `-` among them makes the number negative.
* There is no exponent, and no sign may follow a digit.
* The integer magnitude is limited to 8 388 607 (positive) and 8 388 608 (negative).
  Further integer digits are consumed but ignored, and error bit 8 is set.
* Only the first two fraction digits are kept. The rest are consumed and ignored.
* An *int* parameter is rounded half away from zero on the first fraction digit
  (`2.5` → 3, `-2.5` → −3).
* Quirk: a negative number whose integer part is 0 loses its sign. `-0.4` reads as 0, and
  `-0.5` rounds to +1. Coordinates such as `-0.25` become positive.
* An *angle* is read as value × 100 plus two fraction digits, i.e. hundredths of a degree.

## 4. Units and coordinate system

* **Internal unit:** 0.005 mm (200 internal units = 1 mm).
* **User unit:** each *coord* is multiplied by `steps_per_unit` (0xFFB262), taken from the
  table {20, 10, 5, 2} at 0x28D28. That gives a user unit of 0.1, 0.05, 0.025 or 0.01 mm.
  * The index is stored in the configuration word (EEPROM 0x04, bits 3..2). It is a factory
    setting and is not restored by the factory reset.
  * Everything in the firmware and in host software is consistent with **0.05 mm**
    (20 units per mm), index 1. For example, the built-in demo sends `FU5440,4000` for a
    272 × 200 mm page, and `TB51,400` means 20 mm.
  * Fractions of a user unit are honoured down to the internal resolution: `D10.25,0`
    is 10.25 units.
* **Origin:** the origin and axis direction depend on orientation (`FN`) and media mode
  (`FM`).
  * Portrait (`FN0`, the default): X runs along the media feed and Y across the carriage.
  * Landscape swaps the axes.
* **Cutting area:**
  * Width is 200 mm in media mode 1, or 190 mm with a 5 mm offset in media mode 0.
  * Length is the page length set by `FO`, default 1000 mm.
* **Readback:** `?`, `[`, `U` and `G` divide internal positions by `steps_per_unit` and
  truncate toward zero.

Every point passes through this pipeline (`plot_to`, 0x1DB80):

1. **Line type.** Dashing per `L`/`B`, applied in user space.
2. **Clip polygons.** Defined by `>`. Segments are split at polygon crossings, and drawing
   alternates even-odd.
3. **Factor.** Set by `&`.
4. **Rotation.** Set by `/`.
5. **Orientation swap/mirror.**
6. **Window clipping.** Set by `\` and `Z`, using Cohen-Sutherland against the window.
7. **Device layer.** Applies:
   * the distance correction (`FB`);
   * the registration-mark transform (`TB`);
   * the cutter-offset compensation (`FC`);
   * the motion planner.

The offset (`^`) is applied at the user/device boundary. Changing the offset does not
move the tool.

## 5. Replies

* Every reply is one or more decimal fields separated by `,` and followed by a single ETX.
* Each field is right-aligned and space-padded to a fixed width: 5 or 6 characters (8 or
  10 for some service replies).
* A positive non-zero number always carries one leading space where the sign would be.
  It therefore occupies at least `digits + 1` characters, and the field grows when needed.
* A negative number has `-` directly before the digits.
* Zero is just `0`.

Examples, width 6: `0` → `"     0"`, `100` → `"   100"`, `-5` → `"    -5"`.
A point reply looks like `"   100,    -5" ETX`.

## 6. Errors and status

* `set_error(bits)` (0x10308) ORs `bits & mask` into an error byte. The mask is set by `"`
  and defaults to 0.
* Error bit 1 means an unknown command character or a byte above 0x5F. Bit 8 means a
  number overflow.
* **Nothing ever reports the error byte.** The status-word commands `V` and `#`, and the
  clear commands `:` and `;`, are not implemented. In practice this firmware has no
  readable error status.
* `@` reports the clip outcode of the current position (section 7.4).

## 7. Command reference

The following tables list every command the dispatcher recognises (0xCBF2). Commands
listed in the original Graphtec GP-GL manual but absent here set error bit 1.

### 7.1 Drawing

| Command | Syntax | Description |
|---------|--------|-------------|
| `D` | `Dx1,y1[,x2,y2...]` | Draw (pen down) to each absolute point in turn. Any number of points. |
| `M` | `Mx,y` | Move (pen up) to an absolute point. Only one point is used. |
| `E` | `Edx1,dy1[,dx2,dy2...]` | Draw relative. Each point is relative to the previous one. |
| `O` | `Odx,dy` | Move relative. One point. |
| `H` | `H` | Home. Flushes the motion queue, moves pen-up back by the accumulated chart feed, waits for motion to finish and resynchronises the position. |
| `W` | `Wx0,y0,r1,r2,θ1,θ2[,d]` | Circle, arc or spiral around centre (x0,y0). The radius changes linearly from r1 to r2 while the angle goes from θ1 to θ2 (*angle*). The start point is reached pen-up. `d` is the chord angle per segment: 0 or omitted means automatic (≈ 3150°/√rmax, at most 50°), and negative means \|d\| segments. If max(\|r1\|,\|r2\|) is below 10 internal units, only a dot is plotted at the centre. |
| `]` | `]r1,r2,θ1,θ2[,d]` | Relative circle. Like `W`, but the centre is placed so that the current point lies on radius r1 at angle θ1. The first segment is drawn pen-down. |
| `WP` | `WPa,x1,y1,x2,y2,x3,y3[,d]` | Circle through three points. **Differs from Graphtec:** the firmware reads a leading *int* `a` (clamped to ≥ 0) before the points. |
| `)` | `)a,x0,y0,r1,r2,θ1,θ2,θ3[,d]` | Ellipse with centre (x0,y0) and semi-axes r1/r2, drawn from θ1 to θ2, rotated by θ3. `a` is clamped to ≥ 0. |
| `Y` | `Ya,x1,y1,x2,y2...` | Cubic curve through the points (4-point sliding window, extrapolated end slopes). `a` ≠ 0 closes the curve. If the first point is not the current point, it is reached with a pen-up move. |
| `_` | `_a,dx1,dy1,...` | Like `Y`, with each point relative to the previous one. |
| `BZ` | `BZa,x1,y1,x2,y2,x3,y3,x4,y4[,d]` | Cubic Bézier from P1 to P4. P1 is reached pen-down if `a` = 1, pen-up if `a` = 0. `d` is the segment count: 0 or omitted means automatic, from the control polygon length (4..40 segments). The curve ends exactly at P4. |
| `X` | `Xp,q,r[,t1,t2]` | Axis with tick marks from the current point. `p` = 0 gives a Y axis and 1 an X axis. `p` = 2 or 3 is the same, but positions are computed without accumulated rounding. `q` is the tick interval (*coord*), and `r` the number of intervals (*int*, 0 draws nothing). `t1`/`t2` are the tick lengths on each side, default 200 internal units (1 mm). |

### 7.2 Line type

| Command | Syntax | Description |
|---------|--------|-------------|
| `L` | `Lp` | Line type, `p` = −9..9. Other values are ignored. The types are listed below. |
| `L100` | `L100,n,a,b[,c]` | User-defined dash pattern. `n` = 1 gives [a down, b up]. `n` = 2 gives [a, b, c, b]. `n` = 3 gives [a, b, c, b, c, b]. Lengths are *coord* and must be > 0, otherwise the command is ignored. The pattern period becomes the sum of the lengths. |
| `B` | `Bl` | Line pattern pitch (period) in *coord*. 0 keeps the previous pitch. |

Line types for `L`:

* 0: solid.
* 1..8: ROM patterns. Lengths alternate down and up. The period of 64 is scaled to the
  pitch set by `B`. A length of 0 is a dot.
* 9: dots at the vertices only.
* −1..−9: the same patterns, with the phase shifted by half of the first dash.

| type | pattern (period 64) |
|------|---------------------|
| 1 | 1,63 |
| 2 | 8,56 |
| 3 | 32,32 |
| 4 | 50,14 |
| 5 | 50,7,0,7 |
| 6 | 50,5,4,5 |
| 7 | 50,5,0,4,0,5 |
| 8 | 41,5,4,5,4,5 |

### 7.3 Coordinate setup

| Command | Syntax | Description |
|---------|--------|-------------|
| `\` | `\x,y` | Lower-left corner of the clipping window. |
| `Z` | `Zx,y` | Upper-right corner of the clipping window. Clipping against the window's maximum along the feed axis is only active after `Z` has been sent. |
| `^` | `^x,y` | Offset of the user origin. The tool does not move. |
| `&` | `&p,q,r` | Factor: x × p/r, y × q/r. All three must be non-zero, with \|p/r\| and \|q/r\| < 9, and \|r/p\| and \|r/q\| ≤ 4096. Otherwise the command is silently ignored. |
| `/` | `/x,y,θ` | Rotate by θ (*angle*) around (x,y). θ = 0 turns rotation off. |
| `>` | `>x1,y1,...,xn,yn` | Adds a clipping polygon. Several polygons accumulate, and drawing is even-odd inside them. `>` without parameters clears all polygons. |
| `SO`, `NP` | `SO[n]` | Set the origin at the current carriage position. Along the feed axis it is set to 0, or to the registration-mark origin when one is set. Any parameters are ignored. |
| `FJ` | `FJ` | Re-reference the origin. Waits for motion, clears the registration transform and offsets, and recomputes the soft origin from the current position. |
| `FT` | `FT` | Move pen-up to user (0,0). |

### 7.4 Queries

| Command | Reply | Description |
|---------|-------|-------------|
| `G`, `C` | `x,y,p` (width 6) | Waits for motion to finish, then reports the current position. `p` = tool × 10 + (1 if pen up). `C` (call GIN) is identical to `G`. There is no wait for a button press. |
| `?` | `x,y` (width 6) | Current offset (`^`). |
| `[` | `x,y` (width 6) | Window lower-left (`\`). |
| `U` | `x,y` (width 6) | Window upper-right (`Z`). |
| `@` | `n` (width 5) | Clip outcode of the current position against the hard limits, masked with 0x2F. Bit 0 means x > max, bit 1 x < min, bit 2 y > max, bit 3 y < min. |
| `FG` | `V2.30    ` | Firmware version string: `V2.30` plus four spaces, then ETX. |
| `FA` | `a,b` (width 5) | Distance correction values (EEPROM 0x1C, 0x1E). |
| `FL` | `n` (width 6) | Media mode (0 or 1). |
| `FV` | `n` (width 5) | Media type 1..13. An unset value (0 or > 13) is reported as 13. |
| `FQn` | `v` (width 5) | Stored parameter: `FQ0` speed (1..10, unset → 10), `FQ1` force level (1..3, unset → 3), `FQ2` cutter offset (0..45), `FQ3` overcut on/off (`FE`), `FQ4` orientation, `FQ5` registration-marks-found flag. Other `n` give no reply. |
| `TB71` | `x,y` (width 5) | Registration calibration values. |
| `TB98` | `1` (width 6) | Constant 1. |

### 7.5 Tool and media settings

* Commands that take a pen number `n` parse it and ignore it. Settings always apply to all 8
  pen slots.
* Values marked "saved" are written to EEPROM, and only when they change.

| Command | Syntax | Description |
|---------|--------|-------------|
| `!` | `!l[,n]` | Speed level 1..10. 101..110 is accepted as 1..10. Out of range is ignored. Saved (EEPROM 0x34). |
| `*` | `*a,f[,n]` | Force level. `a` is ignored, and `f` must be 1..3, otherwise the command is ignored. Saved (EEPROM 0x35). Also clears the `FX` force (EEPROM 0x3B). |
| `FX` | `FXn[,m]` | Cutting force 1..30, clamped. Saved (EEPROM 0x3B). `m` is ignored. It overrides the `*` level, which becomes 1 below 7, 2 below 19, otherwise 3. |
| `FC` | `FCp,q[,n]` | Cutter (drag-knife) offset. `p` = 0..45 is the blade tip offset in units of 0.025 mm (internally `p` × 5, capped at 1.2 mm), saved (EEPROM 0x36). `p` = 0 disables the compensation. `q` = 1..3 (default 1; its effect was not found). Out-of-range values are ignored individually. |
| `FD` | `FDθ[,n]` | Blade rotation control: the minimum direction change that triggers a blade swivel, 0..600.00 *angle* (default 30°). `n` must be 0..8, otherwise the command is ignored. At a corner larger than this, the knife is lifted and the tip pivoted around the knife axis in 15° steps. |
| `FE` | `FEp[,n]` | Overcut on (1) or off (0). Saved (EEPROM 0x38). When on, cuts are extended past their end point by the `FF` length. |
| `FF` | `FFa,b[,n]` | Overcut lengths, 0..9 each (default 2), in units of 0.1 mm (internally × 20). |
| `FK` | `FK[n]` | When `n` ≠ 0, each knife touchdown after a swivel is made at full force first. Default off. |
| `FN` | `FN[n]` | Orientation: 0 = portrait, 1 = landscape. Saved (EEPROM 0x37). Moves pen-up to (0,0). |
| `FM` | `FMn` | Media mode. `n` ≠ 0 selects mode 1 (200 mm width, no offset), and 0 selects mode 0 (5 mm offset, 190 mm width). The position is shifted so the tool stays on the same spot. Ignored while the registration transform is active. Not saved. |
| `FO` | `FO[len]` | Page length (*coord*). 0 or missing means 1000 mm. With registration marks active, it is only accepted for manual mark search in landscape. |
| `FW` | `FWn` | Media type 1..13. Out of range becomes 13. Saved (EEPROM 0x3A). |
| `FB` | `FBa,b` | Distance correction. Two values in −200..200 (±2 %), clamped and saved (EEPROM 0x1C, 0x1E). The pen keeps its logical position. |
| `FY` | `FYn` | Track enhancing: `FY1` disables it and `FY0` enables it. Saved (EEPROM 0x3C). |
| `FU` | `FUlen,w` | Track-enhancing roll. Rolls the media back and forth over `len`, as many times as the pass count in EEPROM 0x39. It only acts when the force is ≥ 19 (or the media type implies heavy media), track enhancing is on, and the pass count is non-zero. `w` is ignored. Hosts send the page size here. |
| `F`, `FS` | `Fl` | Chart feed length (*coord*). The value is stored but no movement happens: the feed routine is empty. |
| `TT` | `TT` | Soft reset of the plotting state:<br>• waits for motion and lifts the tool<br>• selects media mode<br>• resets orientation, pen parameters and cutter offsets<br>• seeks the home sensor<br>• reloads the cutter offset from EEPROM |

### 7.6 Registration marks (`TBn[,...]`)

`n` is an *int*. Values with a non-zero high byte are ignored.

| Command | Description |
|---------|-------------|
| `TB0` | Registration transform off. |
| `TB99` | Registration transform on. It is applied after a successful mark search. |
| `TB50[,n]` | Orientation for registration: 0 = portrait, ≠ 0 = landscape. The position is kept. Ignored while the transform is active. |
| `TB51,len` | Mark length (*coord*), accepted 5..20 mm. |
| `TB52[,n]` | Mark type: `n` = 1 (or omitted) selects type 1, anything else type 2 (the default). |
| `TB53,w` | Mark line width, 0.3..1.0 mm. Stored but unused by firmware 2.30. |
| `TB54[,x[,y]]` | Mark offset. |
| `TB55,n` | Sets a flag when `n` ≠ 0. Effect not determined. |
| `TB22[,x,y]`, `TB23[,x,y]` | Manual mark search: the tool has been positioned at the first mark. 22 is search type 1 and 23 is type 2. The optional `x,y` is the mark origin. Reply: `0` (width 5) + ETX if the marks were found, non-zero otherwise. |
| `TB122[,x,y]`, `TB123[,x,y]` | Automatic mark search, same as above. On failure the tool returns to the start, and the page reverts to 1000 × 200 mm. |
| `TB70` | Calibration. Finds the mark centre by searching its edges in four directions and cuts a cross through it. There is no reply. |
| `TB72,x,y` | Set the calibration correction, −40..40 each. Saved (EEPROM 0x24, 0x26). Read it back with `TB71`. |

### 7.7 Test and maintenance

| Command | Description |
|---------|-------------|
| `FH` | Test cut. Cuts a 10 × 10 mm square with a triangle inside at the current position, then advances 10.5 mm. |
| `FI` | Endless test plot (radial lines, zig-zags, circles). **It never returns.** Only a power cycle stops it. |
| `FP` | `FPGRFCC1` enables service mode (section 8). Anything else is ignored. **This is not a firmware update.** |

### 7.8 Accepted but not implemented

These commands are recognised, and their parameters are discarded:

* `DP`, `MP`, `EP`, `OP`, `RP`: polar commands.
* `P`, `K`, `$`, `S`, `Q`, `R`, `I`, `A`, `LP`, `RC`: text and character settings.
* `N`, `SP`, `SA`, `(`, `(P`: marks and user patterns.
* `%`: hatching.
* `T`: buzzer.
* `J`: pen select. The parameter is parsed and ignored.
* `BS`: buffer size.
* `"m[,x,y]`: error mask. It is stored, but the error status is never readable.
* `ZP`, `FR`: skip to the next command.
* `V`, `#`, `:`, `;`, `=`, `<`, `'`: status words, clear, interface clear, terminator.
  **Caution:** the parameter bytes of `=t1,t2` are then interpreted as commands.

## 8. Service mode

After `FPGRFCC1`, lowercase letters a..z are commands (handler table in `gpgl_loop`).
A restart of the interpreter clears service mode. They are listed for completeness; they
change factory calibration and should not be used by host software.

| Command | Effect |
|---------|--------|
| `a0`..`a3` | Sensor/LED tests |
| `b` | Reply home-sensor state |
| `cN` | Set a motion parameter |
| `dN` | Step size index 0..3, i.e. user unit 0.1 / 0.05 / 0.025 / 0.01 mm |
| `ea,b,c,d` | 64-bit arithmetic self test |
| `fADDR` | Reply the RAM byte at ADDR |
| `gN,0` | Query service parameter N |
| `gN,1,v` | Set service parameter N. `g17` is a factory reset of the user EEPROM settings. |
| `hN` | Recompute the acceleration tables |
| `i` | Feed media 40 mm and back |
| `j`, `k`, `y` | Reply internal counters |
| `lN` | Set an internal flag |
| `m`, `q` | Endless burn-in loops |
| `nP` | Set the mark sensor threshold |
| `o` | Undetermined |
| `p` | Toggle the panel LEDs |
| `t0`..`t4` | Mark sensor tests |
| `zN,v` | Set output port bits |

## 9. Non-volatile settings (serial EEPROM, offset from 0x300000)

The factory-reset default is the value written by the factory reset.

| Offset | Content | Set by | Factory-reset default |
|--------|---------|--------|-----------------------|
| 0x04 | Configuration word, including the step-size index (bits 3..2) | service `g` | not reset |
| 0x1C, 0x1E | Distance correction | `FB` | 0, 0 |
| 0x24, 0x26 | Registration calibration | `TB72` | 0, 0 |
| 0x34 | Speed | `!` | 10 |
| 0x35 | Force level | `*` | 3 |
| 0x36 | Cutter offset | `FC` | 0 |
| 0x37 | Orientation | `FN` | 0 |
| 0x38 | Overcut on/off | `FE` | 0 |
| 0x39 | Track-enhancing pass count | service `g18` | not reset |
| 0x3A | Media type | `FW` | 1 |
| 0x3B | Cutting force | `FX` | 0 (unset) |
| 0x3C | Track enhancing disabled | `FY` | 0 |

## 10. Example: the built-in demo program

ROM 0xDB9A..0xF095 (5372 bytes) holds a GP-GL job. Holding the panel key starts it.
It draws a box net of about 245 × 157 mm and is a good example of what Graphtec's own
software sends:

```
FU5440,4000FM1TB50,0FO5440&100,100,100,\0,0,Z5940,4200,L100,1,40,80,^0,100
M44.006,1143.797,D1045.895,1143.797M1045.895,1992.228,D44.006,1992.228 ...
L0,M29.895,1143.797,D59.881,1073.242D1015.909,1073.242 ...
&1,1,1,TB50,0FO0
```

Things to note in it:

* There are no terminators between commands.
* Trailing commas are used.
* Coordinates have three decimals, of which only two are significant.
* `L100,1,40,80` gives dashed fold lines, and `L0` solid cut lines.

## 11. Differences from the published GP-GL reference

* `=` (terminator), `V`, `#`, `:`, `;` and all text, hatching and polar commands are
  no-ops. The terminator is always ETX.
* `WP` and `)` take an extra leading integer parameter.
* `C` does not wait for a button press.
* `*` ignores its first parameter, and its second is a force level 1..3.
* `!` accepts levels 1..10 only (101..110 is mapped to 1..10).
* The pen number in `!`, `*`, `FC`, `FD`, `FE`, `FF` and `J` is ignored.
* `F` (chart feed) does not move the media.
* Numbers keep at most two decimals, and `-0.x` loses its sign.
* Corrections to earlier community notes (`Commands.md`, `cc100.txt`):
  * `FP` is the service unlock, not a firmware update.
  * `TB99` enables the registration transform; it does not reset it.
  * `FI` is an endless test plot.
  * `FU` is the track-enhancing roll.
  * `TT` is a soft reset.
