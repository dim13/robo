# CC100 fw 2.30: plotting and geometry commands

Findings from the Ghidra decompilation plus disassembly of `cc100ver230.bin`. Addresses are ROM
offsets. "Units" means GP-GL user coordinate units as received. Each one is multiplied by
`coord_mult` (see below) to give internal device units.

## Shared parameter readers

| addr | name | behaviour |
|------|------|-----------|
| 0x11a36 | `read_number` | Skips non-numeric bytes before the number. A `-` in that run makes the result negative. Reads the integer digits. After `.`, keeps the first two fraction digits in `num_frac1` (0xffb360) and `num_frac2` (0xffb362) and discards the rest. Magnitude is clamped to 0x7fffff (positive) / 0x800000 (negative) and raises error bit 1 (`set_error`). |
| 0x11b72 | `read_int` | `read_number`, rounded half away from zero on the first fraction digit. |
| 0x11bc6 | `read_coord` | `read_number` × `coord_mult` (0xffb262), plus the fraction scaled and rounded. `coord_mult` comes from table 0x28d28 = {20,10,5,2}, indexed by setting bits 2..3 of 0xffb368 (or by service command `d`). For k=20 and k=10 both fraction digits are used. k=5 uses the first fraction digit, k=2 rounds the first digit to 0/1/2. TODO: confirm the internal unit. The firmware's own test pattern sends `FU5440,4000`, which is consistent with a 0.05 mm user unit. |
| 0x11d50 | `read_angle` | Number × 100 + first two fraction digits, giving hundredths (degrees ×100). |
| 0x13ab2 | `read_point(&x,&y)` | Two `read_coord` calls. |
| 0x0cb3a | `end_of_params` | Returns 0 if another number follows (skips `,` and control/space bytes). Returns 1 on a terminator (`=` command chars, 0xffb359/0xffb35a), ESC/FS, 0x80, `/` or any printable non-number character. |
| 0x0806a | `read_opt_angle` | 0 if `end_of_params`, else `read_angle`. Used for optional `d`. |

A missing number reads as 0 because `read_number` returns 0 when no digit follows. Extra numbers
after a fixed-arity command are discarded by the main loop: characters whose `cmd_class` < 0
(digits, `,`, `.`, `+`, `-`, whitespace) are ignored there.

## Coordinate pipeline

`plot_to(x,y,pen)` 0x1db80 runs the following stages in order:

1. `plot_linetype` 0x1df5e. Stores the user position `cur_x/cur_y` (0xffb584/8). For pen down with
   a non-solid line type, it splits the segment into dash pieces (see L/B). For type 9, it puts a
   point at the start, moves pen-up to the end, then applies the pen there (dots at vertices only).
2. `plot_clip_poly` 0x1eb88. If clipping polygons are defined (`>`), the segment is intersected
   with them (0x6140). Pieces alternate pen state at each crossing, which gives even-odd
   inside drawing.
3. `plot_transform` 0x1e23e. Device position = user position with these steps applied:
   - factor: `x = x*p/r`, `y = y*q/r` (muldiv 0x16130), skipped when that factor is unity
     (`factor_unity` bits).
   - rotation about (`rot_cx`,`rot_cy`) using float `rot_sin/rot_cos` (0x1e86a), only if `rot_enabled`.
   - orientation: if `orient_flags` (0xffb5c6) is non-zero and not 0x60, x/y are swapped. If bit 6
     is clear, `x = 0xffb574 - x`. If bit 5 is clear, `y = 0xffb578 - y`. `update_window` sets
     bit 6 when LL.x > UR.x or LL.y > UR.y (both branches set bit 6, so bit 5 is never set
     there). TODO: may be a firmware bug.
   - window clipping (Cohen–Sutherland) against `clip_xmin..xmax / ymin..ymax` (0xffb554..0xffb560),
     outcodes 1:x>xmax 2:x<xmin 4:y>ymax 8:y<ymin. When 0xffb400 == 0, the x>xmax test is only
     active after `Z` has been given (`ur_set` 0xffb3b9); otherwise the y>ymax test is the one
     masked (roll-feed axis is unbounded).
   - `plot_device` 0x1c572: calibration and machine transform. Out of scope here.
4. `sync_user_pos` 0x18e44 recomputes `cur_x/cur_y` from the device position with the inverse
   transforms (orientation, rotation, factor). This runs after state-changing commands (`/`, `&`,
   `^`, `\`, `Z`).

`plot_state_reset` 0x190bc is the defaults:
- line type 0, all 8 ROM patterns reloaded
- factor 1/1/1
- offset 0
- rotation off (sin 0, cos 1.0f)
- rotation centre 0
- window = full hardware range (`window_reset` 0x1900c)
- position 0,0

## Commands

| cmd | handler | syntax | behaviour |
|-----|---------|--------|-----------|
| `D` | 0x13bca | `Dx1,y1,...,xn,yn` | While another number follows: `read_point`, `plot_to(x,y,1)`. Any number of points. |
| `M` | 0x13d3e | `Mx,y` | One point, `plot_to(x,y,0)`. Further numbers are ignored. |
| `E` | 0x13be8 | `Edx1,dy1,...` | Like D, but each point is relative: x += cur_x, y += cur_y (0x13b16). |
| `O` | 0x13d50 | `Odx,dy` | Relative move, one point, pen up. |
| `DP`,`MP` | 0x0ffd4 | | No-op. Parameters ignored. |
| `EP`,`OP` | 0x0ffe6 | | No-op. |
| `RP` | 0x10004 | | No-op. |
| `B` | 0x13b7e | `Bl` | Line scale/pitch `l` (coord). Stored in `line_pitch_arg`, then `set_line_type(type,l)`. `BS` (0x0ffc8) is a no-op. `BZ` is a Bézier, see below. |
| `L` | 0x13c68 | `Lp` | `p` = `read_int`. `set_line_type(p, pitch)`: accepted range −9..9. Out of range leaves the type unchanged. A non-zero pitch is stored as \|pitch\| in `line_pitch` (0xffbc72). Pitch 0 keeps the previous pitch. 0 = solid. 1..8 = ROM patterns. Negative = same pattern, phase shifted by half the first dash. 9 = dots at vertices. |
| `L100` | 0x13c68 | `L100,n,a,b[,c]` | User pattern, `n` = 1..3 (other values are ignored), `a`,`b`,`c` lengths (coord) > 0. Any non-positive length aborts silently. `n=1`: [a down, b up]. `n=2`: [a, b, c, b]. `n=3`: [a, b, c, b, c, b]. Pitch is set to the sum of lengths. The type becomes 100 (0x2032e). |
| `BZ` | 0x071d6 | `BZa,x1,y1,x2,y2,x3,y3,x4,y4[,d]` | Cubic Bézier. `a` (int, clamped ≥0) is the pen state used for the first `plot_to` to P1 (0 = move, 1 = draw). `d` = number of segments. If 0/omitted, automatic: control polygon length L = \|P1P2\|+\|P2P3\|+\|P3P4\| (in units after a divide). L<21 gives 4 segments, L≥1561 gives 40, otherwise a sqrt-based value. Ends exactly at P4. |
| `W` | 0x08c36 | `Wx0,y0,r1,r2,θ1,θ2[,d]` | Arc or spiral about (x0,y0). The radius goes linearly from r1 to r2 as the angle goes θ1→θ2. Angles are degrees with two decimals. `d` = chord step in degrees: 0/omitted = auto (`315000/sqrt(rmax)` hundredths, capped at 50°). A negative `d` means \|d\| segments. If max(\|r1\|,\|r2\|) < 10 internal units, it plots only a dot at the centre. The first point is reached pen-up. If setting bit 6 of 0xffb368 is set and 0xffb288≠0, the pen state comes from 0xffb2b6 instead. |
| `WP` | 0x0828e | `WPa,x1,y1,x2,y2,x3,y3[,d]` | Three-point circle (0x834c). Note: the firmware reads a leading int `a` (clamped ≥0) before the three points. |
| `]` | 0x08d14 | `]r1,r2,θ1,θ2[,d]` | Relative arc. The centre is placed so the current point lies at angle θ1 on radius r1: `c = cur − r1·(cos θ1, sin θ1)`. The first point is drawn pen-down. |
| `)` | 0x081f8 | `)a,x0,y0,r1,r2,θ1,θ2,θ3[,d]` | Ellipse (0x92e8). `a` int ≥0. r1,r2 axes. θ1,θ2 start/end. θ3 axis rotation. |
| `Y` | 0x08cf8 | `Ya,x1,y1,...` | Cubic spline through the points (0x9992, 4-point sliding window, end slopes extrapolated). `a`≠0 gives a closed curve. Pen-up move to the first point unless it equals the current point. |
| `_` | 0x08df8 | `_a,dx1,dy1,...` | Same as Y, but each point is relative to the previous one (the first is relative to the current position): the running origin accumulates instead of being reset to 0. |
| `%` | 0x1006a | | Hatching: no-op. |
| `X` | 0x13eb4 | `Xp,q,r[,t1,t2]` | Axis with ticks, pen down, starting from the current point. `p` (int): 0 = Y axis, 1 = X axis, 2/3 = same but each tick position is computed from the start (q·k), avoiding accumulated rounding. `q` interval (coord). `r` interval count (int). 0 draws nothing. `t1`/`t2` tick length up/down (coord, default 200 internal units each). It draws r intervals and r+1 ticks. |
| `>` | 0x1112c | `>x1,y1,...,xn,yn` | Adds a clipping polygon (points are pushed on a downward stack at 0xffb2c4, bounded by 0xffbc78, with a polygon counter). Several `>` accumulate. `>` with no parameters clears all clip polygons (0x110ba). |
| `/` | 0x075e2 | `/x,y,θ` | Rotation centre (`rot_cx`,`rot_cy`) and angle θ (degrees, two decimals). θ=0 disables rotation. sin/cos are stored as float in 0xffb5b2/0xffb5b6. Then `sync_user_pos`. |
| `&` | 0x07522 | `&p,q,r` | Factor: x×p/r, y×q/r. All three `read_int`, all must be non-zero. Accepted only if \|p/r\|,\|q/r\| < 9 and \|r/p\|,\|r/q\| ≤ 4096. Otherwise ignored, with no error. Unity factors are flagged in `factor_unity`. |
| `^` | 0x17838 | `^x,y` | Offset. The device position is kept: `dev += old_off − new_off`, then `sync_user_pos`. |
| `\` | 0x177da | `\x,y` | Window lower-left (0xffb564/8), then `update_window` (min/max into clip box). |
| `Z` | 0x1777c | `Zx,y` | Window upper-right (0xffb56c/70) and `ur_set`=1 (enables the x>xmax clip). |
| `ZP` | 0x0d412 | | Skips everything up to the next command letter (same as `FR`). |
| `H` | 0x13c22 | `H` | Flushes the queue (0x1108a), moves pen-up by −`feed_total` (0xffb3b0) in X, which undoes accumulated chart feed, waits for motion (0x1db52), then resyncs the position (0x18f0e). |
| `J` | 0x13c58 | `Jn` | Reads one int and ignores it. Pen select is a no-op on this single-tool device. |
| `(` | 0x100be | | No-op (user pattern). |
| `(P` | 0x1002e | | No-op. |
| `N` | 0x10046 | | No-op (mark). |

## Line patterns

ROM table 0x2901a (used when 0xffb26c==0, which is the default). An alternative set is at 0x29064.
Lengths alternate down/up, starting with down. Each pattern's period is 64 (or 100 for set B), and
the lengths are scaled so one period equals `line_pitch`. At load time they are copied to RAM
0xffbf76 + n·0x2a (n = 1..8), with the period stored at 0xffc0ee + 2n.

| type | set A (period 64) | set B (period 100) |
|------|-------------------|--------------------|
| 1 | 1,63 | 1,99 |
| 2 | 8,56 | 50,50 |
| 3 | 32,32 | 70,30 |
| 4 | 50,14 | 80,10,0,10 |
| 5 | 50,7,0,7 | 70,10,10,10 |
| 6 | 50,5,4,5 | 50,10,10,10,10,10 |
| 7 | 50,5,0,4,0,5 | 70,10,0,10,0,10 |
| 8 | 41,5,4,5,4,5 | 50,10,0,10,10,10,0,10 |

A length of 0 means a dot. `L` only accepts −9..9, and there is no ROM pattern for 9: type 9 is
"dots at vertices", handled in `plot_linetype`.

Not covered: the device-level motion (`plot_device` and below), pen force/speed, and `FS`/`F` chart
feed (`FS` reads one coord into 0xffb3aa and calls the no-op 0x1005e).
