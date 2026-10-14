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

This is the stroke font used by `font_interp` (0xB96A). Offsets are relative to the start of
the file, which is ROM 0x26794.

| ROM | Offset | Content |
|-----|--------|---------|
| 0x26794 | 0x0000 | Character code remap (16 bytes) |
| 0x267A4 | 0x0010 | Glyph metrics tables (4 signed bytes per glyph). `font_metrics` (0xC994) also reads 0x267A8, 0x26BA8, 0x26C58 and 0x26DE0. |
| 0x26E00 | 0x066C | Glyph programs for the RAM font sets at 0xFFC278, 0xFFC298 and 0xFFC2B8 (see the pointer tables below). |
| 0x285A0 | 0x1E0C | Sub-glyph table, called by opcodes 0xF5–0xF9. |
| 0x28954 | 0x21C0 | Index of 8 glyph-group pointers for ROM font 0, followed by its glyphs. |
| 0x28B54 | 0x23C0 | National character substitution maps (`#$@[\]^_{|}~` variants). |
| 0x28BCC | 0x2438 | Accented-character composition pairs. |

How glyphs are found and drawn:

* A glyph group is selected by code bits 7..5.
* Inside the group, the glyph is reached by skipping `code & 31` records that end in 0xFF.

| Opcode | Meaning |
|--------|---------|
| 0x00–0xED | Point: x is the high nibble, y the low nibble |
| 0xEE | Pen up for the next point |
| 0xEF | Special scaling |
| 0xF0–0xF4 | Arcs |
| 0xF5–0xF9 | Call a sub-glyph |
| 0xFA–0xFD | Curves |
| 0xFE | Pen down |
| 0xFF | End or return |

The RAM font pointer sets (in `init_data.bin`) each hold 8 group pointers:

* 0xFFC278: 0x26E00 0x26EDD 0x27005 0x2710C 0x271F2 0x2732D 0x2741D 0x2758B
* 0xFFC298: 0x26E00 0x276EF 0x27860 0x2710C 0x271F2 0x2732D 0x2741D 0x2758B
* 0xFFC2B8: 0x27968 0x27AB4 0x27C2A 0x27D5A 0x27EC7 0x27FF2 0x280ED 0x284F8

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

The length field says 0x004D (77 bytes including itself), but the text is 81 bytes long. A host
reading exactly the stated length gets `...DESCRIPTION:Graphtec CC`. This was not checked on
hardware.

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
