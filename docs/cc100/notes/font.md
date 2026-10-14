# CC100 fw 2.30: stroke font interpreter

The font engine lives at 0xadae–0xc994. Its data is in ROM 0x26794–0x28c04
(`assets/font.bin`) and in the pointer tables of `assets/init_data.bin`. Its only users in
2.30 are the version banner and service mode `o` (data dump), because `P`/`K` are no-ops.

The firmware behaviour is re-implemented in `assets/render_font.py`. The rendered sheets are
in `assets/font/`.

## Font selection

**Defaults.** `font_defaults` (0xadae) sets up:

| Setting | Value |
|---------|-------|
| Width | 600 |
| Height | 600 |
| Line pitch | 900 |
| Char pitch | 600 |
| Rotation (`0xffb198`/`0xffb19a`) | 0x4000 / 0 (2.14 cosine/sine) |
| Slant (`0xffb19e`) | 0 |
| Arc step (`0xffb19c`) | 8 |
| Font id (`font_select`) | 0x3fc = 1020 |

**Mode word.** `font_map_code` (0xc5fc) turns the font id into the mode word `0xffb1a2`:
* An id of 1000 or more gives `0x80 | BCD(id-1000)`. For 1020 that is `0x8020`.
* A smaller id gives a family/variant code with bit 15 clear.

**Glyph tables.** Every table is 8 group pointers. The group is chosen by code bits 7..5, and
the glyph is the `(code & 31)`-th 0xFF-terminated record in that group (`font_lookup` 0xb890).

| Table | Used when | Lookup |
|-------|-----------|--------|
| RAM 0xffc278 | mode bit 15 set (id ≥ 1000) and id ≠ 1120 | `font_find_glyph` 0xb7ec |
| RAM 0xffc298 | id 1120 (0x460). Groups 1 and 2 are replaced: rounder digits, plain `0`. | `font_find_glyph` |
| RAM 0xffc2b8 | mode bit 15 clear (id < 1000) | `font_find_glyph`, `font_find_glyph2` (code bit 8) |
| ROM 0x28954 | mode bit 15 clear: symbols and fractions | `font_find_glyph2` |

**Scale** (`font_begin` 0xb14a):
* y unit = height/14.
* x unit = width/14 when id ≥ 1000, width/8 when id < 1000.

So the c2b8 font and the ROM symbols are 14/8 wider than the c278/c298 fonts.

## Character to glyph (`font_glyph` 0xb352), mode bit 15 set (default font 1020)

**Initial pen state:**
* Codes below 0x20 start with the pen **down** and the origin moved to (0x400, 0x400).
  These glyphs are marks centred on the current point.
* All other codes start with the pen up and origin 0.

**Codes 0x00–0x7f:**
* `font_step(c)` sets the arc step. It is halved for `'0'`; otherwise it comes from
  `0xffb1d0`.
* `'0'`: the glyph is taken from `0x28b25[2·k]`, where k = mode bits 7..5. For 1020, k = 1,
  giving glyph 0x0d, a slashed zero.
* `'O'`: the glyph is taken from `0x28b26[2·k]`. For 1020 this is plain `O` (0x4f).
* National variants come from the 12-character set `#$@[\]^_{|}~` at 0x28b54. Row `mode & 15`
  gives the replacement. A replacement below 0x0f is a composite: base glyph
  `0x28bcc[2r]` plus an accent glyph `0x28bcd[2r]`, which is pushed on the call stack and
  drawn after the base. For 1020 the row is 0, the identity.

**Codes 0x80–0xdf:**
* Drawn as is: Latin-1 extras, Greek capitals and katakana.
* With mode bit 4 set, codes 0x80–0xbf become a space, and 0xc0–0xd8 are remapped through
  `0x28b2a[c]`.

**Codes 0xe0–0xf8:**
* Drawn with an extra slant of 0x2b/256, about 9.5° (`0xffb224 += 0x2b`). This is the italic
  Greek lowercase.

When mode bit 15 is clear (c2b8 fonts), the code goes through the metrics tables
0x267a8/0x26ba8/0x26c58/0x26de0 and the composite tables at 0x2865b. Glyphs come from c2b8
or from the ROM table. This path is only summarised here.

## Interpreter (`font_interp` 0xb96a)

### State

All values are int16.

| Field | Meaning |
|-------|---------|
| `x`, `y` | Current point, in 1/256 grid cell |
| `step` | Minimum arc step |
| `r` | Radius. Non-zero while an arc is in progress. |
| `a`, `end` | Current and final angle. 128 = full turn. |
| `cx`, `cy` | Arc centre |
| `da` | Signed angle step |
| `mag` | Step magnitude |

Other interpreter state:
* **Call stack:** `0xffb1d6[]`, with the depth in `0xffb1d4`.
* **Pen flag** `0xffb1be`: 0 means the next point is a move; 0xffff means it draws.
* **`font_draw_to`** (0xc4d8) maps a point through the 2×2 matrix (`font_scale_point` 0xaef2:
  scale, rotation, slant), calls `plot_solid`, then sets the pen to draw. As a result,
  consecutive points form one polyline.

### Main loop

```
while depth >= 0:
    if r != 0: arc_continue()        # one chord per iteration
    else: execute opcodes until one plots a point or starts an arc
```

### Opcodes

`unpack(b)` gives `x = (b >> 4) · 256` and `y = (b & 15) · 256`: the high nibble is the
advance direction, the low nibble is up. `s8` is a signed byte.

| Byte | Operands | Effect |
|------|----------|--------|
| 00–ED | — | Point `unpack(b)`. Draws, or moves if the pen flag says so. |
| EE | — | The next point draws. |
| EF | g | Continue with glyph g of the current table (a jump, not a call). x scale × 5/8, y scale × 19/32. The scale is restored by the next FF. |
| F0–F4 | c, v | Arc with radius ((b & 7) + 1)·256 and centre `unpack(c)`. Start = (v >> 1) & 0x70, a multiple of 45°. Span = `s8((v & 0x1f) << 3)`, a multiple of 22.5°; 0 means −128, a full turn. |
| F5–F9 | — | Call sub-glyph `*(0x285a0 + 4·b)`. The pointers live at 0x28974–0x28984. |
| FA p | — | Relative point: `x += s4(p >> 4)·256`, `y += s4(p & 15)·256`. |
| FA FA p n | — | Fine point: `unpack(p)` + (`(n >> 3) & 7`, `n & 7`) · (256 >> (n >> 6)). |
| FA EF p n v w s | — | Fine centre (as FA FA), radius ((v & 0xe0) << 3) \| w, start (v & 0x1f)·4, span `s8(s)`. |
| FB | v | Arc from the current point. Radius ((v >> 6) + 1)·256. Start q = (v & 0x30) << 1 (0°, 90°, 180°, 270°); the centre is placed so the current point lies at angle q. Span `s8(v << 4)` (multiples of 45°). |
| FC | c, v, s | Centre `unpack(c)`, radius ((v >> 5) + 1)·256, start (v & 0x1f)·4, span `s8(s)`. |
| FD | b1 b2 b3 b4 | cx = `s8(b1 & 0xfc)`·64, cy = `s8((b1 & 3) << 6 \| (b2 >> 2) & 0x3c)`·64, radius ((b2 & 15) << 9 \| (b3 & 0x80) << 1), start b3 & 0x7f, span `s8(b4)`. |
| FE | — | The next point is a move (pen up). |
| FF | — | Return. Pops the call stack. At depth < 0 the glyph ends. Otherwise the scale is restored and, for mode bit 15 clear, the next point becomes a move. |

### Arcs (`font_arc_step` 0xc160, `font_continue` 0xc24c, `font_arc` 0xc2b0)

**Starting an arc:**
* `a = start & 0x7f`.
* `q = 0x6400 / r`.
* Step `mag` is 16, or 8 if q < |sy|, or 1 if 3q < |sy|. It is then raised to at least
  `step`, which defaults to 8 (4 for `'0'`).
* `sy` = `0xffb220` = height/14 = 42 by default.
* `da = ±mag`, following the sign of the span.
* `end = a + span`.
* The first point is plotted immediately, with the current pen state.

**Continuing an arc:**
* `a = (a + da) & -mag`, so steps align to multiples of `mag`.
* Once `a` reaches or passes `end`, the arc plots `end` and clears `r`.

**Point position:** `int(r·cos(a·π/64) + cx)`, `int(r·sin(a·π/64) + cy)`, truncated toward zero.

At the default size most arcs are octagonal (45° chords) and larger ones use 22.5°. This is
why the rendered sheets look faceted. The real plot does the same.
