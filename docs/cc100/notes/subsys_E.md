# CC100 fw 2.30 — subsystems in 0x00400–0x10fff (wave 2, fork E)

Function names are in `sym/wave2_E.txt`, and readable C for the larger routines is in `src/subsys_E.c`.

Calling conventions:
- Application code passes arguments in ER0 and ER1, then on the stack.
- The soft-float runtime (0x22xxx–0x24xxx) passes pointers instead: ER0 = result, ER1 = a, ER2 = b.

Soft-float helpers seen from this range. Their names belong to the 0x22000+ range and are given here only as a reading aid:

| addr | role |
|---|---|
| 0x23358 | long → double (`*r = (double)l`) |
| 0x22a8a | double add |
| 0x22a66 | double sub (negates b, then calls add) |
| 0x234a0 | double mul |
| 0x22eee | double div |
| 0x23174 | double → long (truncate) |
| 0x1f3ea | double → long, rounded (adds ±0.5 from 0x290c2/0x290ca) |
| 0x231f6 | double == 0 test (via compare 0x240a0) |
| 0x23f1e | sqrt (used on dx²+dy²) |
| 0x2339e / 0x23316 / 0x236ac / 0x22c6a / 0x22c74 / 0x23276 | single-precision float: from long, from int, mul, add, sub, to long |

## Interrupt vectors

Each used vector points to a small wrapper (`vecN`, 0x41a–0x4da). The wrapper does `push er6; mov sp,er6; stm er0-er1; jsr handler; ldm; pop er6; rte`. Every unused vector, and most reserved ones, point to an empty `rte` stub (0x4f2–0x70e).

The reset entry at 0x400 does: `mov.l #0xffefb0,sp; orc #0x80,ccr (mask interrupts); jsr main; sleep`.

Vector numbers follow the H8S/2200 advanced-mode layout. Peripheral names are inferred from the H8S/2215 vector map and from what each handler does.

| vec | wrapper | handler | source | function |
|---|---|---|---|---|
| 0 | 0x400 | main 0x13940 | reset | |
| 18 | 0x462 | 0x12ff2 | IRQ2 | front-panel key. A long press starts the demo (`demo_run`), per device.md |
| 40 | 0x41a | 0x1bc6a | TPU1 TGI1A | stepper pulse timer, motor axis 1 (motion range, not analysed here) |
| 44 | 0x432 | 0x1bcc4 | TPU2 TGI2A | stepper pulse timer, motor axis 2 |
| 64 | 0x47a | timer_tick_isr 0x12d7a | TMR0 CMIA0 | system tick (delays, idle timer, key scan) |
| 80 | 0x492 | sci0_eri_isr 0x1f592 | SCI0 ERI0 | RS-232 receive error |
| 81 | 0x4aa | sci0_rxi_isr 0x1f4a4 | SCI0 RXI0 | RS-232 receive (ESC 00/02 handling) |
| 82 | 0x4c2 | sci0_txi_isr 0x1f5c6 | SCI0 TXI0 | RS-232 transmit |
| 83 | 0x4da | 0x1f61c | SCI0 TEI0 | RS-232 transmit end |
| 104 | 0x44a | usb_isr 0x220b2 | USB | USB function controller interrupt |
| 5–11, 16–17, 19–36, 41–43, 45–47, 65–75, 84–91, 105 | 0x4f2..0x70e | — | trace, NMI, TRAPA, other IRQ/TPU/TMR/SCI1/SCI2 | empty `rte` |

## Runtime data initialisation

The table at ROM 0x29630 is the C startup section table:
- **Initialised data:** copy ROM 0x29644..0x298e0 to RAM 0xffc230.
- **BSS:** clear RAM 0xffb048..0xffc230.

The initialised data includes:
- `dump_reply_tab` (0xffc230, 18 string pointers, see below);
- the font table pointers 0xffc278/0xffc298/0xffc2b8;
- the USB endpoint descriptor block around 0xffc42a. Its stride is 0x18 bytes per endpoint, and it holds pointers to 0xffb000 and to the USB controller registers 0xc000xx.

## Hardware init: `hw_init_bus_ports` (0x739c)

Called from `main` (0x13940) and 0x137ec. It writes fixed values to:
- the bus-controller and port registers at 0xfffde4–0xfffdec, 0xfffe12–0xfffe47, 0xfffec0–0xfffecc and 0xfffed0–0xfffed5;
- ABWCR/ASTCR/WTCR-class registers, DDRs and pull-ups.

The argument selects bit 2 of 0xfffe14: arg == 0 writes 4, otherwise 0. The function ends by clearing bit 6 of 0xffff02.

Two external peripherals sit on the bus:
- **0x600010** is a write-only latch. It drives LEDs and motor enables; 10 and 15 are written by the `p` service command.
- **0xc00000** is a memory-mapped USB function controller (registers 0xc00084/86/c0/c4/cc).

## Work buffer: `workbuf_init` (0x6c3c)

Called from `main`, 0x137ec and the IRQ2 key handler.
- Copies 6 region sizes from ROM 0x2660a = {3072, 1024, 0, 0, 0, 0} to `workbuf_sizes` (0xffb104).
- `workbuf_fix_sizes` (0x6ba8) grows region 0 so the six sizes total 0x1800 (6 KiB). Region 0 becomes 5120 bytes and region 1 stays 1024.
- `workbuf_layout` (0x6aac) carves the 6 KiB RAM area at 0xffc800 into consecutive regions. The pointers go to 0xffb0e8..0xffb100, plus an end pointer at 0xffb11c.
- Finally calls `clip_reset` and `rx_flush_masked`.

This is the GP-GL `BS` (buffer size) concept with fixed sizes. `BS` itself is not implemented, so the layout never changes.

## Clip-polygon support (used by `plot_clip_poly` 0x1eb88)

- **`clip_segment_hits`** (0x6140) intersects one segment (x0,y0)→(x1,y1) with every edge of every clip polygon on the polygon stack.
  - The polygon stack grows downward. Each polygon is stored as a vertex count followed by (x,y) pairs.
  - The segment is first normalised by translating to (x0,y0). If |dy| > |dx|, x and y are swapped; flags bit0 = swapped, bit1 = direction negative.
  - `clip_edge_local` (0x6950) translates each edge into that frame.
  - The side tests use 0x162c8 (sign of a cross product) and 0x162e6.
  - Proper crossings are computed exactly with doubles and rounded (0x1f3ea) into `out[]`. At most 30 points are kept.
  - Touching edges are counted in `*n_touch`. The segment end point is always appended.
- **`clip_sort_hits`** (0x4100) is an exchange sort of the (x,y) list by the major axis, ascending or descending according to the same flags (mode&1 = sort by y, mode&2 = descending). This orders the hits along the segment, so `plot_clip_poly` can toggle the pen at each crossing.

## Registration marks: geometry (0x46b8–0x5ecc)

These routines are called by the manual and automatic mark searches (`regmark_search_manual` 0x1575a, `regmark_search_auto` 0x15fb2) and by the calibration routine 0x1fc4c.

- **Mark storage.** Detected marks are stored at `regmark_marks` (0xffb0ae). `regmark_get_mark(idx, &x, &y)` (0x5452) returns mark 0/1 (both 0xffb0ae/b2), mark 2 (b0b6/ba) or mark 3 (b0be/c2).
- **`regmark_find_corners`** (0x46b8) turns the detected mark edges into corner points. The edges are stored as double line coefficients at 0xffb3bc..0xffb3e4.
  - It uses `regmark_line_intersect` (0x48ba, general line–line intersection) or `regmark_axis_intersect` (0x5082, one line axis-parallel).
  - Both set `*parallel = 1` when the lines do not intersect.
  - Type-2 marks (search type bits 0x30 == 0x10) use two intersections; type-1 marks use three.
- **`regmark_compute_transform`** (0x54e4):
  - Takes mark 1 and mark 2 and forms the vector between them. Its length comes from `sqrt(dx²+dy²)` (0x23f1e).
  - Unit direction cosines go to doubles `rm_yx`/`rm_xx` (0xffb06a/72). Rounded fixed versions go to 0xffb08a..0xffb090.
  - The 4-mark type also uses mark 3 for the second axis (0xffb07a/82).
  - `regmark_scale_x`/`_y` (0x5bc0/0x5d40) work out the measured mark distance. If a nominal distance was given with TB23/TB123 (0xffc112/16) and bit 6 of `regmark_cfg` is set, they compute a scale factor in 2.14 fixed point (0x4000 = 1.0) into 0xffb0a2/0xffb0a4 and set 0xffb0ab.
  - The result is an origin offset `regmark_off_x/y` (0xffb4f0/f4) and `regmark_active = 1`. Then `sync_device_pos`.
  - On any degenerate case it sets `regmark_result = 1` (failure), parks with 0x1c678(0,0), and returns.
- **`regmark_extend_segment`** (0x5ecc) lengthens the segment p1–p2 by 3000/|p1p2| of its length at both ends. It is used by the calibration cross (TB70) and by 0x1ff2a.

## Built-in stroke font (0xadae–0xc994)

There is a complete vector font engine, even though `P`/`K` (labels) are no-ops in this firmware. Its only users are the data-dump mode (service `o`) and the version banner.

- **`font_defaults`** (0xadae):
  - width = height = 600, line pitch 900, character pitch 600 (internal units; 600 = 3 mm);
  - slant 0, scale 0x4000 (1.0 in 2.14);
  - font id 0x3fc (`font_select`).
- **`font_draw_string`** (0xae5c) draws each character with `font_glyph` (0xb352), then advances by `font_char_pitch` along +y.
- **`font_map_code`** (0xc5fc) maps a character code to a glyph-table selector. Font ids ≥ 1000 use BCD-coded tables, and ids 99/101 are special cases.
- **`font_metrics`** (0xc994) fetches 4 signed bytes of metrics per glyph from ROM tables 0x267a4/0x267a8/0x26ba8/0x26c58/0x26de0.
- **`font_lookup`** (0xb890) finds a glyph program. It uses 8 table pointers (code bits 7..5), then skips (code & 31) 0xFF-terminated records.
- **`font_interp`** (0xb96a) is a byte-code interpreter with a call stack (0xffb1d6, depth in 0xffb1d4):

| byte | meaning |
|---|---|
| 0x00–0xED | point: high nibble = x, low nibble = y (`font_unpack_xy`), then draw to the scaled point (`font_draw_to` 0xc4d8 → `plot_solid`) |
| 0xEE | pen up for the next point |
| 0xEF | special (0xbb70, uses division and scaling) |
| 0xF0–0xF4 | arc / circle primitives (0xbbe2 → 0xc160/0xc2b0) |
| 0xF5–0xF9 | call a sub-glyph from table 0x285a0 (push) |
| 0xFA–0xFD | curve primitives (0xbcae, 0xbd38, 0xbdf0, 0xbe78) |
| 0xFE | pen down |
| 0xFF | return (pop the call stack; end of glyph at depth 0) |

## Data-dump mode: service command `o` (0xaab2)

`svc_o_dump_mode` never returns:
1. Resets the font (`font_defaults`).
2. Moves to (2000, 2000), i.e. 10 mm / 10 mm.
3. Prints the firmware version string (`fw_version_string`), feeds one line and lifts the tool (0x1d6f8).
4. Enters `dump_loop` (0xa8a6).

Every byte read through `dump_getc` (0xa5b8) is plotted with `dump_print_char` (0xacd6):
- Printable characters, and Kana when 0xffb26c is set, are drawn as themselves.
- Control characters, 0x7f, and 0xa0–0xff when Kana is off are drawn as two hex digits plus a space, at 2/3 size and half pitch (`dump_hex_byte` 0xab7a, `dump_format_char` 0xac58).
- After each character, `dump_advance` moves along y by the pitch (twice for hex). After 36000 units (180 mm) it starts a new line at x += line pitch.

The dump loop also watches the stream so that host queries still get an answer. Answers are canned strings from `dump_reply_tab` (0xffc230; the first character of each string is skipped). Each reply is terminated by `dump_term1` and also `dump_term2` if different. Both default to ETX.

| received | reply idx | reply text |
|---|---|---|
| `G` | 0 | `  3690,  2715,    10` |
| `C` | 1 | `  3690,  2715,    10` |
| `?` | 2 | `  3690,  2715` |
| `[` | 3 | `     0,     0` |
| `U` | 4 | `  3690,  2715` |
| `@` | 5 | `    0` |
| `BS` | 6 | `0` |
| `FA` | 7 | `    0,    0` |
| `FG` | — | firmware version string |
| `FL` | 11 | `     1` |
| `FQ0`..`FQ5` | 12..17 | `   10`, `    3`, `    0`, `    0`, `    0`, `    0` |
| `TB98` | 8 | `     1` |
| `TB71` | 10 | `    0,    0` |
| `TB22`/`TB23` | 9 | `    0` |
| `=ab` | — | sets the dump terminators to a, b |
| `P…`, `K…` | — | text skipped up to a terminator |
| `(` | — | `dump_skip_frame('(')`: it is meant to skip a 0x80-framed block, but it receives the `(` byte itself, so it does nothing |

Bytes 0x2b–0x2e, digits and characters above 0x5f are passed over without matching. All bytes are printed in any case.

## Pen tables and cutter-offset compensation (0x10332–0x10fd8)

**`pen_tables_init`** (0x10332) fills 8 entries per table from EEPROM:

| entry | address | source |
|---|---|---|
| speed | 0xffb416 | NV 0x34, valid 1..10, default 10 |
| force | 0xffb436 | NV 0x35, valid 1..3, default 3 |
| — | 0xffb426 | 1 |
| cutter offset | 0xffb446 | 0 |
| cutter q | 0xffb456 | 1 |
| swivel threshold angle | 0xffb466 | 30000 = 30.000° (long) |
| FE mode | 0xffb486 | NV 0x38, valid 0/1 |
| FF a | 0xffb496 | 2 |
| FF b | 0xffb4a6 | 2 |

It also loads `cut_force` (NV 0x3b), `media_type` (NV 0x3a, 1..13, default 13) and `track_enhance_off` (NV 0x3c), and clears `fk_flag`.

**`select_pen_params`** (0x1054c) copies the current pen's entries into the working variables:
- `cutoff_radius` = offset × 5, max 240 internal units (1.2 mm);
- `cutoff_min_angle`;
- `cutoff_mode` (FE);
- overcut lengths FF a×20 and b×20 (0xffb2b8/0xffb2ba);
- the speed code via 0x181da.

It then marks the parameters dirty.

The compensation itself is a tangential (drag-knife) offset. Angles are in 1/1000 degree, so 360000 = full turn.
- **`cutoff_on_move`** (0x10686) runs on each pen-down segment. It computes the segment direction with 0x164c6 (degrees ×10, then ×100). If the direction changed, it calls `cutoff_swivel`.
- **`cutoff_swivel`** (0x10718) turns the blade:
  - If the turn is smaller than `cutoff_min_angle`, nothing is done. This check only applies when the threshold is non-zero and bit 5 of 0xffbc46 is clear.
  - Otherwise it lifts the knife (unless FE mode 2 swivels in place) and pivots the blade tip.
  - The pivot follows a circle of radius `cutoff_radius` around the knife axis, in 15° steps (15000).
  - The last step goes through `cutoff_set_lead`. The knife is then lowered again; when `fk_flag` is set, a full-force (0xff) touchdown comes first.
- **`cutoff_set_lead`** (0x10c1e) sets the offset vector (r·cos, r·sin) for an angle. **`cutoff_move`** (0x10ca4) issues a device move (0x1d4aa) to the target plus that offset.
- **`cutoff_end`** (0x10ce0) and **`cutoff_overcut`** (0x10fd8) extend the cut past the end point by the FF overcut length in the current direction when FE mode is on.
- **`cutoff_blade_init`** (0x10d90) runs at the first pen-down after a parameter change. It cuts a short alignment stroke: down 200 units, then sideways 6×radius. The stroke is made at y−400, or at the media edge / registration origin depending on 0xffb368 bit 1. This orients the knife before the real job.

## Distance correction (FA/FB): 0xf096–0xf61e

`cmd_FB_set_distance_corr` stores the ±200 correction values in EEPROM 0x1c/0x1e. It then calls 0x12040, which calls `distcorr_compute` (0xf096).

- `distcorr_compute` builds double scale factors from the corrections and ROM constants 0x28c78..0x28c90:
  - forward: `distcorr_x`/`_y` (0xffb24a/52);
  - inverse: `distcorr_inv_x`/`_y` (0xffb36a/72);
  - an extra term at 0xffb25a/5e.
- `distcorr_to_device` (0xf61e) and `distcorr_to_user` (0xf4b2) apply or undo the scaling on a point. They are used when FB changes the correction while a pending position (0xffb4f8) exists, and by 0x18ae0/0x19518.
- `distcorr_apply_page` (0xf330) rescales `page_length`/`page_width`. It is unreferenced in 2.30.

## USB function controller driver (0xf78a–0xfe12)

The handlers are reached through the endpoint/event table at ROM 0x290d2, the same table that leads to `usb_bulk_out_rx` 0x11556. The controller registers are at 0xc00084/86 (FIFO data/select), 0xc000c0 (interrupt status, write-0-to-clear masks), 0xc000c4 (interrupt enable) and 0xc000cc.

- **`usb_ep0_setup`** (0xf942) handles the setup packet. The 8-byte packet is read into 0xffb000 (`usb_setup_pkt`); 0xffb008 holds data.
- **`usb_std_request`** (0xfb6c) decodes it:
  - type bits 0x60 = 0x00 → standard request; 0x20 → class request (`usb_class_request` 0xfe12); 0x40 → vendor request (`usb_vendor_request` 0xfdf6); anything else stalls (0x1f6be);
  - GET_DESCRIPTOR(6) returns the device (1), configuration (2) or string (3) descriptor, clipped to wLength;
  - SET_DESCRIPTOR(7) is accepted into a buffer.
- **Printer class requests** (`usb_class_request`):
  - GET_DEVICE_ID (0) returns the IEEE-1284 ID string pointed to by 0xffc2de, length 0xffc2dc;
  - GET_PORT_STATUS (1) returns 0x18 (selected, no error);
  - SOFT_RESET (2) clears the tx state.
- **Vendor request 0** with wValue: `jog_keys = wValue low byte`. This is how the host drives the jog keys, together with the ESC 00 sequence.
- `usb_ep0_in` (0xf9f8) and `usb_ep0_out` (0xfa9c) handle the data and status stages.
- `usb_ep_in_fill` (0xf78a) copies up to the endpoint's packet size from `tx_buf` (0xffb2f3), starting at `usb_tx_offset`, into the IN FIFO. It clears `tx_busy` when done. `usb_ep_in_start` (0xf8ee) arms the IN endpoint.

Descriptors (ROM):

| item | address | value |
|---|---|---|
| device descriptor | 0x29166 | bcdUSB 1.00, max packet 64, **VID 0x0B4D (Graphtec), PID 0x110A**, bcdDevice 1.00, 1 configuration |
| configuration | 0x29178 | one interface, class 7 (printer), subclass 1, protocol 2 (bidirectional), self-powered, 32 mA; EP1 OUT bulk 64, EP2 IN bulk 64 |
| strings | 0x29198.. | language 0x0409, then "123" and short ASCII strings |

## Spline and arc helpers

- **`spline_segment`** (0x9d24) is the inner loop of `Y`/`_` (via `spline_feed` 0x9992). It draws one Catmull-Rom segment between p1 and p2 of a 4-point window (p0..p3).
  - `spline_coeffs` (0x9ebc) computes per-axis coefficients `c0 = 4(3(p1-p0) - 3(p2-p0) + (p3-p0))`, `c1 = 2(4(p2-p0) - 5(p1-p0) - (p3-p0))`, `c2 = p2-p0`, `c3 = p1`. The parameter runs over 0..0x8000 in 16.16 fixed point, which supplies the usual ½ factors.
  - `spline_eval_fx` (0x9fc4) evaluates the cubic with Horner's rule using rounded 16.16 multiplies.
  - The step count is max(|p1−p2|x, |p1−p2|y)/64, minimum 4. All points are drawn with `plot_to(..., pen=-1)` (continue).
- **`circle3_draw`** (0x834c) is the `WP` solver: circumcentre from three points with doubles, then `arc_draw`.
- **`arc_draw_float`** (0x9824) is an alternative arc renderer with single-precision sin/cos. It is unreferenced in 2.30.
- **`fx_mul16`** (0xa12a) is a 16.16 multiply. It is unreferenced.

## Miscellaneous

- **`approx_dist`** (0x6c76) approximates hypot without sqrt. With a = |dx|, b = |dy| and b > a: if b < 4a, the result is b + a/2 − (b/4)/2, otherwise b. Symmetric for a > b. Used by `bezier_draw` to choose the segment count.
- **`set_area_extent`** (0x7984) sets 0xffb51c/0xffb520 = (0xffb514 − x, 0xffb518 − y), with a minimum of 1, and sets 0xffb4f8/fc = (x, y). Then 0x151be(1) and `sync_device_pos`. It is used by `set_media_mode` and the TB50 orientation change.
- **`move_mm`** (0xd4c8) does a device move to (x·200, y·200) with the pen up. 200 internal units = 1 mm, so the internal unit is 0.005 mm. It is used by `demo_run`.
- **`cmd_term_unused`** (0x7636) is the real GP-GL `=t1,t2` handler: it reads two raw bytes into `term1`/`term2` and calls `term_changed_hook` (0x100ca, empty). **Nothing references it**, so `=` is a no-op in 2.30 and the terminator cannot be changed. This confirms the input/control notes.
- **`plot_soft_reset`** (0x102ae) calls `cmd_nop_A`, then 0x13a8c, then `init_terminators` (unless keep_term != 0) and `clip_reset`. It is called from 0x151be.
- **`svc_e_calc`** (0x101fe) is service `e a,b,c,d`. It replies `%10d,%10d,%10d` with `muldiv`-like 0x1631a(a,b,c,d), 0, 0. It is a factory math test.
- **Dispatcher hooks:**
  - `hook_nop` (0xffbc) is empty and called before many commands (formerly a buffer sync);
  - `hook_pen_param` (0xd448) is `hook_nop()`;
  - `hook_sync` (0xd458) is `hook_nop(); wait_motion (0x1db52)`, used before `C` and `G`;
  - `hook_B` (0xffc8) is empty.
- **Empty command handlers:** 0xffd4 (`DP`/`MP`), 0xffe6 (`EP`/`OP`), 0x10004 (`RP`), 0x1002e (`(P`), 0x10046 (`N`), 0x1006a (`%`), 0x100be (`(`).
