#!/usr/bin/env python3
"""Extract data assets from CC100 firmware 2.30 ROM image.

usage: extract.py [cc100ver230.bin] [outdir]
"""
import os
import sys

# name, start, end (exclusive), description
ASSETS = [
    ("demo_job.plt", 0x0DB9A, 0x0F096, "GP-GL job run by demo_run (0xd512): box net with fold lines"),
    ("dump_replies.bin", 0x266BA, 0x26748, "canned replies of service 'o' dump mode, NUL-separated"),
    ("font.bin", 0x26794, 0x28C04, "stroke font: code map, metrics, glyph byte-code, national char maps"),
    ("step_table.bin", 0x28D28, 0x28D2C, "steps per user unit {20,10,5,2} (u8)"),
    ("fi_tables.bin", 0x28D6C, 0x28DBA, "FI test plot polylines (s16 pairs)"),
    ("atan_table.bin", 0x28DE2, 0x28EE4, "atan lookup, 129 x u16 (vector_angle 0x164c6)"),
    ("axis_pattern.bin", 0x28F28, 0x28F84, "axis correction test pattern points (axis_correction_pattern 0x173b4)"),
    ("motion_table.bin", 0x28FD2, 0x2901A, "motion_init (0x1af64) table copied to RAM 0xffbbd8"),
    ("linetypes.bin", 0x2901A, 0x290BA, "dash patterns for L1..L8, set A (period 64) and set B (period 100)"),
    ("ieee1284_id.bin", 0x28C98, 0x28CEC, "IEEE 1284 device ID for GET_DEVICE_ID: u16 length 0x004d, then text, NUL"),
    ("version.txt", 0x28CEC, 0x28CF5, "firmware version reply of FG"),
    ("usb_descriptors.bin", 0x29166, 0x291B8, "USB device, configuration and string descriptors"),
    ("init_data.bin", 0x29644, 0x298E0, "initialised RAM image, copied to 0xffc230 at startup"),
]


def main():
    src = sys.argv[1] if len(sys.argv) > 1 else os.path.join(os.path.dirname(__file__), "..", "..", "cc100ver230.bin")
    out = sys.argv[2] if len(sys.argv) > 2 else os.path.dirname(os.path.abspath(__file__))
    rom = open(src, "rb").read()
    assert len(rom) == 0x298E0, "unexpected image size"
    for name, start, end, desc in ASSETS:
        with open(os.path.join(out, name), "wb") as f:
            f.write(rom[start:end])
        print(f"{name:20s} 0x{start:05x}..0x{end:05x} {end - start:5d}  {desc}")


if __name__ == "__main__":
    main()
