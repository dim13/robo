# ROM data assets

These are data blocks cut out of `cc100ver230.bin`. Recreate them with:

```
python3 extract.py ../../cc100ver230.bin .
```

In `extract.py`, each file is listed with its exact ROM range. All multi-byte values are
big-endian.

| File | ROM range | Content |
|------|-----------|---------|
| `demo_job.plt` | 0x0DB9A–0x0F096 | GP-GL job (plain text) |
| `dump_replies.bin` | 0x266BA–0x26748 | Canned query replies |
| `font.bin` | 0x26794–0x28C04 | Stroke font |
| `step_table.bin` | 0x28D28–0x28D2C | Step-size table |
| `fi_tables.bin` | 0x28D6C–0x28DBA | `FI` test plot shapes |
| `atan_table.bin` | 0x28DE2–0x28EE4 | Arctangent lookup |
| `axis_pattern.bin` | 0x28F28–0x28F84 | Axis-correction test points |
| `motion_table.bin` | 0x28FD2–0x2901A | Motion initialisation table |
| `linetypes.bin` | 0x2901A–0x290BA | Line type patterns |
| `ieee1284_id.bin` | 0x28C98–0x28CEC | Printer device ID |
| `version.txt` | 0x28CEC–0x28CF5 | Version string |
| `usb_descriptors.bin` | 0x29166–0x291B8 | USB descriptors |
| `init_data.bin` | 0x29644–0x298E0 | Initialised RAM image |

## `demo_job.plt`

* This is the job that `demo_run` (0xD512) plays.
* `demo_run` points the receive buffer at this text and runs `gpgl_loop` on it.
* Holding the panel key starts it.
* It is plain GP-GL with no terminators between commands.
* It draws a box net of about 245 × 157 mm, with dashed fold lines (`L100,1,40,80`).
* It is a valid input for `robo` or any GP-GL renderer.

## `dump_replies.bin`

* These are 12 NUL-terminated strings.
* RAM table `dump_reply_tab` (0xFFC230, from `init_data.bin`) points at them.
* Service command `o` (data dump mode) answers host queries with them.
* The first character is the query letter and is skipped when the reply is sent. Example:
  `G  3690,  2715,    10` is sent as `  3690,  2715,    10` + ETX.

## `font.bin`

This is the built-in stroke font. `render_font.py` re-implements the firmware's font
interpreter and draws the glyphs into `font/`:

| File | Content |
|------|---------|
| `font_1020.svg` | Characters 0x20–0xFF as the default font 1020 draws them. Note the slashed zero and the italic Greek lowercase at 0xE0–0xF8. |
| `font_c278.svg` | Raw glyph table 0xFFC278 (font id ≥ 1000). |
| `font_c298.svg` | Raw glyph table 0xFFC298 (font id 1120). |
| `font_c2b8.svg` | Raw glyph table 0xFFC2B8 (font id < 1000), drawn 14/8 wider as the plotter does. |
| `font_rom.svg` | Raw ROM symbol table 0x28954 (symbols, fractions). |

Run `python3 render_font.py [outdir]` to regenerate them. It uses only the standard library.

Offsets below are relative to the start of the file, which is ROM 0x26794.

| ROM | Offset | Content |
|-----|--------|---------|
| 0x26794 | 0x0000 | Code remap for mode bit 9 (codes < 0x0F) |
| 0x267A4 | 0x0010 | Metrics: 4 signed bytes per glyph (`font_metrics` 0xC994). Tables at 0x267A4, 0x267A8, 0x26BA8, 0x26C58 and 0x26DE0. |
| 0x26E00 | 0x066C | Glyph programs of the RAM tables 0xFFC278, 0xFFC298 and 0xFFC2B8 (see the pointer tables below) |
| 0x2865B | 0x1EC7 | Composite pairs for codes < 0x20 in the c2b8 fonts |
| 0x28954 | 0x21C0 | ROM symbol table: 8 group pointers, then its glyphs |
| 0x28974 | 0x21E0 | Sub-glyph pointers for opcodes F5–F9 (base 0x285A0 + 4·opcode) |
| 0x28B25 | 0x2391 | `'0'` / `'O'` variant table, selected by mode bits 7..5 |
| 0x28B54 | 0x23C0 | National variants of `#$@[\]^_{|}~`: 12 bytes per row, row = mode & 15 |
| 0x28BCC | 0x2438 | Composite (base, accent) glyph pairs for national variants below 0x0F |

**Glyph lookup.** Every table holds 8 group pointers. A code selects a group with bits 7..5.
Inside the group, the glyph is reached by skipping `code & 31` records that each end in 0xFF.

**Glyph coordinates.**
* A point byte holds x in the high nibble (advance direction) and y in the low nibble (up).
* The c278/c298 fonts use an 8 × 14 cell. Both axes are scaled by size/14.
* The c2b8 font scales x by width/8 instead.
* The pen starts up, every point after the first draws, and `FE` lifts the pen.

**Byte code.** `s8` is a signed byte, and angles are in 1/128 of a turn.

| Byte | Operands | Meaning |
|------|----------|---------|
| 00–ED | — | Point (x = high nibble, y = low nibble) |
| EE | — | The next point draws |
| EF | g | Jump to glyph g, scaled 5/8 (x) × 19/32 (y) |
| F0–F4 | c v | Arc of radius (b & 7) + 1 around centre c. Start = ((v >> 5) & 7) · 45°. Span = `s8((v & 0x1F) << 3)`, and 0 means a full turn. |
| F5–F9 | — | Call a sub-glyph |
| FA | p | Relative point (signed nibbles) |
| FA FA | p n | Fine point: p plus a fraction from n (3 + 3 bits, scale 256 >> (n >> 6)) |
| FA EF | p n v w s | Fine centre. Radius ((v & 0xE0) << 3 \| w) / 256. Start (v & 0x1F) · 4. Span `s8(s)`. |
| FB | v | Arc continuing from the current point. Radius (v >> 6) + 1. Start (v >> 4 & 3) · 90°. Span `s8(v << 4)`. |
| FC | c v s | Centre c. Radius (v >> 5) + 1. Start (v & 0x1F) · 4. Span `s8(s)`. |
| FD | 4 bytes | Signed 6-bit centre, 5-bit radius, 7-bit start, 8-bit span |
| FE | — | The next point is a move |
| FF | — | Return, or end of glyph |

**Arc tessellation.** Arcs are drawn as chords, aligned to multiples of 16, 8 or 1
(in 1/128 turn), depending on the radius and text size. At the default 3 mm size the steps are
mostly 45° or 22.5°, which is why the sheets look faceted.

**Pointer tables.** The RAM font tables (in `init_data.bin`) each hold 8 group pointers:

* 0xFFC278: 0x26E00 0x26EDD 0x27005 0x2710C 0x271F2 0x2732D 0x2741D 0x2758B
* 0xFFC298: 0x26E00 0x276EF 0x27860 0x2710C 0x271F2 0x2732D 0x2741D 0x2758B
* 0xFFC2B8: 0x27968 0x27AB4 0x27C2A 0x27D5A 0x27EC7 0x27FF2 0x280ED 0x284F8

The full interpreter semantics, with addresses, are in `../notes/font.md`.

## `linetypes.bin`

* These are the dash patterns for line types `L1`–`L8`.
* There are two sets: set A (period 64, the default) followed by set B (period 100,
  used when 0xFFB26C ≠ 0).
* Each pattern starts with 0xFFFF and is followed by u16 lengths.
* Bit 15 set means pen down; the low bits are the length.
* Example: `FFFF 8001 003F` is 1 unit down, then 63 up.
* At load time the lengths are scaled so that one period equals the `B` pitch.

## `step_table.bin`

* Four `u8` values: {20, 10, 5, 2}.
* Each is the number of internal steps (0.005 mm) per GP-GL user unit.
* The value in use is selected by EEPROM 0x04 bits 3..2, or by service command `d`.

## `atan_table.bin`

* 129 `u16` values.
* Entry i is atan(i/128) in tenths of a degree (the last entry, 450, is 45.0°).
* `vector_angle` (0x164C6) uses it to get segment directions for cutter-offset compensation.

## `fi_tables.bin`, `axis_pattern.bin`, `motion_table.bin`

**`fi_tables.bin`**

* Polylines (`s16` x,y pairs) drawn by the endless `FI` test plot (0x14246).
* Sub-tables start at ROM 0x28D6C, 0x28D74 and 0x28D82.

**`axis_pattern.bin`**

* Moves for `axis_correction_pattern` (0x173B4).
* Coordinates are `u16`/`s32` mixed, see the decompile.

**`motion_table.bin`**

* `motion_init` (0x1AF64) copies this table into RAM at 0xFFBBD8.
* It holds motor/acceleration profile seeds.

## `ieee1284_id.bin`

* The device ID returned by the USB printer-class request GET_DEVICE_ID.
* Format: a big-endian `u16` length, then text, then NUL.

```
MANUFACTURER:Graphtec;MODEL:CC100-20;CLASS:PRINTER;DESCRIPTION:Graphtec CC100-20;
```

The embedded length field says 0x004D (77 bytes), although the text is 81 bytes. The USB
request handler ignores it and sends 0x53 (83) bytes, so hosts receive the complete string.
Only the prefix is wrong.

## `version.txt`

The `FG` reply: `V2.30` followed by four spaces. The firmware appends ETX when sending it.

## `usb_descriptors.bin`

| Offset | Length | Descriptor |
|--------|--------|------------|
| 0x00 | 18 | Device: USB 1.0, class 0, max packet 64, VID 0x0B4D, PID 0x110A, bcdDevice 1.00, no string indices, 1 configuration. |
| 0x12 | 32 | Configuration (self-powered), then interface 0: class 7/1/2 (printer, bidirectional), 2 endpoints. Endpoint 0x01 is bulk OUT, 64 bytes; endpoint 0x82 is bulk IN, 64 bytes. |
| 0x32 | 4 | String 0: language 0x0409. |
| 0x36 | 8 | String: `"123"` in UTF-16. |
| 0x3E | 20 | Leftover ASCII bytes (`1234 21234 3qrst tqrst u`). They are listed in the descriptor table at 0x29136 with length 5, but no descriptor references them. |

## `init_data.bin`

The C startup (section table at 0x29630) copies this image to RAM at 0xFFC230. BSS
(0xFFB048–0xFFC230) is zeroed. The image contains:

* `dump_reply_tab`: 18 pointers into `dump_replies.bin`.
* The three font pointer sets listed above.
* Assorted initial values of RAM variables. Look up the names in `../symbols.txt` by RAM
  address, which is 0xFFC230 plus the file offset.
