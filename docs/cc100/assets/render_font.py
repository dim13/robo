#!/usr/bin/env python3
"""Render the CC100 firmware 2.30 stroke font (font.bin) to SVG specimen sheets.

Reimplements font_interp (0xb96a) and its primitives; see ../notes/font.md.
Coordinates are glyph units of 1/256 grid cell: x = high nibble (advance
direction), y = low nibble (up). Arc tessellation follows the firmware
(font_arc_step 0xc160 / font_continue 0xc24c) at the default text size.

usage: render_font.py [outdir]     (default: ./font next to this script)
"""
import math
import os
import struct
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
FONT_BASE = 0x26794   # font.bin = ROM 0x26794..0x28c04
RAM_BASE = 0xFFC230   # init_data.bin is copied here

font = open(os.path.join(HERE, "font.bin"), "rb").read()
ram = open(os.path.join(HERE, "init_data.bin"), "rb").read()


def rb(a):
    return font[a - FONT_BASE]


def rom32(a):
    return struct.unpack(">I", font[a - FONT_BASE:a - FONT_BASE + 4])[0]


def ram32(a):
    return struct.unpack(">I", ram[a - RAM_BASE:a - RAM_BASE + 4])[0]


def s8(v):
    v &= 0xFF
    return v - 0x100 if v & 0x80 else v


def s16(v):
    v &= 0xFFFF
    return v - 0x10000 if v & 0x8000 else v


# Glyph tables: 8 group pointers, group = code bits 7..5, glyph = (code & 31)-th 0xFF-terminated record.
TABLES = {
    "c278": [ram32(0xFFC278 + 4 * i) for i in range(8)],  # font id >= 1000 (default 1020)
    "c298": [ram32(0xFFC298 + 4 * i) for i in range(8)],  # font id 1120 (0x460)
    "c2b8": [ram32(0xFFC2B8 + 4 * i) for i in range(8)],  # font id < 1000
    "rom": [rom32(0x28954 + 4 * i) for i in range(8)],    # symbol table (font_find_glyph2)
}
GROUP_ENDS = sorted({a for t in TABLES.values() for a in t} | {0x28974, 0x2865B, 0x28B54})


def lookup(code, table):
    """font_lookup 0xb890. Returns None when the record lies past the group's data."""
    start = table[(code >> 5) & 7]
    end = min(a for a in GROUP_ENDS if a > start)
    p = start
    for _ in range(code & 31):
        while p < end and rb(p) != 0xFF:
            p += 1
        p += 1
    return p if p < end else None


def unpack(b):
    """font_unpack_xy 0xc450."""
    return (b & 0xF0) << 4, (b & 0x0F) << 8


def glyph_strokes(prog, table, res=42, pen_draw=False, origin=None, mode_neg=True):
    """Run font_interp on program address prog; return list of polylines.

    res is |ffb220| (glyph height scale, font_height/14 = 42 by default) and only
    affects arc tessellation. pen_draw/origin model font_glyph's start state
    (codes < 0x20 start with the pen down at (0x400, 0x400)). mode_neg is the sign
    of the font mode word ffb1a2: True for the c278/c298 fonts, False for c2b8/ROM.
    """
    ef_table = table if mode_neg else TABLES["c2b8"]
    strokes = []
    cur = list(origin) if origin else None
    sx = sy = 1.0          # EF scaling (ffb21c / ffb220 relative to base)
    base_res = res
    pen = pen_draw         # ffb1be: False = next point is a move
    stack = [prog]         # ffb1d6[], depth = len-1
    ctx = dict(x=0, y=0, step=8, r=0, a=0, end=0, cx=0, cy=0, da=0, mag=0)

    def draw(x, y):
        nonlocal pen, cur
        p = (x * sx, y * sy)
        if pen and cur is not None:
            if strokes and strokes[-1][-1] == cur:
                strokes[-1].append(p)
            else:
                strokes.append([cur, p])
        cur = p
        pen = True  # font_draw_to leaves ffb1be = 0xffff

    def arc_point():
        t = ctx["a"] * math.pi / 64          # 128 units = full turn
        ctx["x"] = int(ctx["r"] * math.cos(t) + ctx["cx"])
        ctx["y"] = int(ctx["r"] * math.sin(t) + ctx["cy"])
        draw(ctx["x"], ctx["y"])

    def arc_start(start, span):
        """font_arc_step 0xc160."""
        ctx["a"] = start & 0x7F
        q = int(0x6400 / ctx["r"])
        mag = 16
        if q < res:
            mag = 8
            if 3 * q < res:
                mag = 1
        mag = max(mag, ctx["step"])
        ctx["mag"] = mag
        span = s8(span) or -128
        ctx["da"] = -mag if span < 0 else mag
        ctx["end"] = ctx["a"] + span
        arc_point()

    def nxt():
        b = rb(stack[-1])
        stack[-1] += 1
        return b

    def fine_point():
        """font_op_helper3 0xc0a8: skip marker, point byte, fraction byte."""
        stack[-1] += 1
        x, y = unpack(nxt())
        n = rb(stack[-1])
        m = 0x100 >> ((n >> 6) & 3)
        x += ((n >> 3) & 7) * m
        n = nxt()
        y += (n & 7) * m
        return x, y

    while stack:
        if ctx["r"]:                               # font_continue 0xc24c
            ctx["a"] = (ctx["a"] + ctx["da"]) & -ctx["mag"]
            d = s16(ctx["a"] - ctx["end"])
            if d == 0 or (d ^ ctx["da"]) >= 0:
                ctx["a"] = ctx["end"]
                arc_point()
                ctx["r"] = 0
            else:
                arc_point()
            continue
        while True:
            b = nxt()
            if b == 0xEE:
                pen = True
            elif b == 0xEF:                        # jump to glyph, scale 5/8 x 19/32
                stack[-1] = lookup(nxt(), ef_table)
                sx, sy = sx * 5 / 8, sy * 19 / 32
                res = int(res * 19 / 32)
            elif 0xF0 <= b <= 0xF4:                # circle arc, radius (b&7)+1
                ctx["r"] = ((b & 7) + 1) << 8
                ctx["cx"], ctx["cy"] = unpack(nxt())
                v = nxt()
                arc_start((v >> 1) & 0x70, v << 3)
                break
            elif 0xF5 <= b <= 0xF9:                # call sub-glyph
                stack.append(rom32(0x285A0 + 4 * b))
            elif b == 0xFA:
                n = rb(stack[-1])
                if n == 0xFA:                      # fine absolute point
                    ctx["x"], ctx["y"] = fine_point()
                    draw(ctx["x"], ctx["y"])
                elif n == 0xEF:                    # fine centre, free radius
                    ctx["cx"], ctx["cy"] = fine_point()
                    v = nxt()
                    ctx["r"] = ((v & 0xE0) << 3) | nxt()
                    arc_start((v & 0x1F) << 2, nxt())
                else:                              # relative point, signed nibbles
                    dx, dy = unpack(nxt())
                    ctx["x"] += s16(dx << 4) >> 4
                    ctx["y"] += s16(dy << 4) >> 4
                    draw(ctx["x"], ctx["y"])
                break
            elif b == 0xFB:                        # arc continuing from current point
                v = nxt()
                r = (((v & 0xC0) + 0x40) << 2)
                ctx["r"] = r
                q = (v & 0x30) << 1
                rs = r if q & 0x40 else -r
                if q & 0x20:
                    ctx["cx"], ctx["cy"] = ctx["x"], ctx["y"] + rs
                else:
                    ctx["cx"], ctx["cy"] = ctx["x"] + rs, ctx["y"]
                arc_start(q, v << 4)
                break
            elif b == 0xFC:                        # arc: centre byte, radius/start, span
                ctx["cx"], ctx["cy"] = unpack(nxt())
                v = nxt()
                ctx["r"] = ((v & 0xE0) + 0x20) << 3
                arc_start((v & 0x1F) << 2, nxt())
                break
            elif b == 0xFD:                        # arc: 6-bit signed centre, 5-bit radius
                b1, b2, b3, b4 = nxt(), nxt(), nxt(), nxt()
                ctx["cx"] = s8(b1 & 0xFC) << 6
                ctx["cy"] = s8(((b1 & 3) << 6) | ((b2 >> 2) & 0x3C)) << 6
                ctx["r"] = (((b2 & 0xF) << 8) | (b3 & 0x80)) << 1
                arc_start(b3, b4)
                break
            elif b == 0xFE:
                pen = False
            elif b == 0xFF:                        # return / end of glyph
                stack.pop()
                if stack:
                    sx = sy = 1.0
                    res = base_res
                    if not mode_neg:
                        pen = False
                    continue
                break
            else:                                  # point 0x00..0xED
                ctx["x"], ctx["y"] = unpack(b)
                draw(ctx["x"], ctx["y"])
                break
    return strokes


def char_strokes(c):
    """Character c as font_glyph 0xb352 draws it in the default font 1020 (mode 0x8020)."""
    table = TABLES["c278"]
    pen, origin, slant = c < 0x20, (0x400, 0x400) if c < 0x20 else None, 0
    step = 8
    if c < 0x80:
        if c == 0x30:
            step = 4                               # font_step: finer arcs for '0'
            c = rb(0x28B25 + 2)                    # mode bits 7..5 = 1: slashed zero
        elif c == 0x4F:
            c = rb(0x28B26 + 2)
        # national map row 0 is the identity for mode 0x8020
    elif c >= 0xE0 and c < 0xF9:
        slant = 0x2B                               # ffb224 += 0x2b: italic shear
    p = lookup(c, table)
    if p is None:
        return []
    s = glyph_strokes(p, table, pen_draw=pen, origin=origin)
    if origin:
        s = [[(x - origin[0], y - origin[1]) for x, y in st] for st in s]
    if slant:
        s = [[(x + y * slant / 256, y) for x, y in st] for st in s]
    return s


def svg(cells, title, path, aspect=1.0):
    """cells: list of (code, strokes); 16 columns, cell 16x18 grid units.

    aspect stretches x: fonts with id < 1000 scale x by width/8 instead of width/14
    (font_begin 0xb14a), i.e. 14/8 relative to y.
    """
    title = title.replace("&", "&amp;").replace("<", "&lt;")
    cw, ch = 16 * 256, 18 * 256
    rows = (cells[-1][0] >> 4) - (cells[0][0] >> 4) + 1
    r0 = cells[0][0] >> 4
    out = [f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {16 * cw} {rows * ch + 2 * 256}" '
           f'width="{16 * 48}" height="{(rows * ch + 2 * 256) * 48 // cw}">',
           '<style>path{fill:none;stroke:#000;stroke-width:40;stroke-linecap:round;stroke-linejoin:round}'
           'text{font:180px monospace;fill:#888}rect{fill:none;stroke:#ddd;stroke-width:8}</style>',
           f'<text x="0" y="300" style="font-size:300px;fill:#000">{title}</text>']
    for code, strokes in cells:
        col, row = code & 15, (code >> 4) - r0
        ox, oy = col * cw + (16 - 8 * aspect) * 128, 2 * 256 + row * ch + 16 * 256
        out.append(f'<rect x="{col * cw}" y="{2 * 256 + row * ch}" width="{cw}" height="{ch}"/>')
        out.append(f'<text x="{col * cw + 40}" y="{2 * 256 + row * ch + 200}">{code:02X}</text>')
        if strokes:
            d = " ".join("M" + " L".join(f"{ox + x * aspect:.0f} {oy - y:.0f}" for x, y in st) for st in strokes)
            out.append(f'<path d="{d}"/>')
    out.append("</svg>")
    open(path, "w").write("\n".join(out) + "\n")


def check():
    a = char_strokes(ord("A"))
    assert a == [[(0, 0), (1024, 3584), (2048, 0)], [(256, 768), (1792, 768)]], a
    b = char_strokes(ord("B"))
    xs = [x for st in b for x, _ in st]
    ys = [y for st in b for _, y in st]
    assert (min(xs), max(xs), min(ys), max(ys)) == (0, 2048, 0, 3584), (xs, ys)
    o = glyph_strokes(lookup(0x30, TABLES["c278"]), TABLES["c278"])
    assert len(o) == 1 and abs(o[0][0][0] - o[0][-1][0]) <= 1 and o[0][0][1] == o[0][-1][1], o


def main():
    check()
    out = sys.argv[1] if len(sys.argv) > 1 else os.path.join(HERE, "font")
    os.makedirs(out, exist_ok=True)
    svg([(c, char_strokes(c)) for c in range(0x20, 0x100)],
        "CC100 font 1020 (default): characters 20..FF", os.path.join(out, "font_1020.svg"))
    titles = {"c278": "glyph table 0xFFC278 (font id >= 1000)", "c298": "glyph table 0xFFC298 (font id 1120)",
              "c2b8": "glyph table 0xFFC2B8 (font id < 1000)", "rom": "ROM glyph table 0x28954 (symbols)"}
    for name, table in TABLES.items():
        codes = range(0x60) if name == "rom" else range(0x100)
        cells = []
        for c in codes:
            p = lookup(c, table)
            neg = name in ("c278", "c298")
            cells.append((c, glyph_strokes(p, table, mode_neg=neg) if p is not None else []))
        svg(cells, f"CC100 {titles[name]}: glyphs 00..{codes[-1]:02X}", os.path.join(out, f"font_{name}.svg"),
            aspect=1.0 if neg else 14 / 8)


if __name__ == "__main__":
    main()
