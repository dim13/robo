# CC100 firmware 2.30 reverse engineering

Decompilation of `../cc100ver230.bin`, the firmware of the Graphtec CC100 (Craft ROBO) cutting
plotter.

The CPU is a Renesas H8S running in advanced mode (24-bit addresses), with 256 KB flash at
0x000000 and on-chip RAM at 0xFFB000..0xFFEFBF. External devices sit at fixed addresses:

* USB controller: 0xC00000
* stepper and LED latches: 0x600000..0x600010
* serial EEPROM: 0x300000+n (a bit-banged 93Cxx part)

| Path | Content |
|------|---------|
| `GP-GL.md` | GP-GL specification as implemented by this firmware. |
| `symbols.txt` | ~800 names and prototypes (functions and RAM globals), applied by the Ghidra scripts. |
| `decompiled.c` | Full raw Ghidra output with `symbols.txt` applied (681 functions). |
| `functions.txt` | Function list: address, size, callers, callees, name. |
| `src/*.c`, `src/cc100.h` | Hand-cleaned, readable C. Each function carries its original address. |
| `notes/*.md` | Analysis notes per subsystem. `notes/dispatch.txt` is the command dispatch table. |
| `ghidra/H8S` | Ghidra processor module for H8S (see below). |
| `ghidra/scripts` | Headless scripts: `Setup` (memory map, vectors), `Names` (apply symbols), `Dump` (decompile all), `Funcs`, `Stats`, `Dis`. |
| `ghidra/build.sh` | Regenerates `decompiled.c` and `functions.txt`. |
| `assets/` | ROM data cut into files (demo GP-GL job, stroke font, line types, USB descriptors, ...) by `assets/extract.py`. |

Readable sources by area:

| File | Area |
|------|------|
| `src/input.c` | Receive buffer, USB/serial input, escape sequences, number parsing, main command loop. |
| `src/control.c` | Query, interface, pen and tool commands; reply formatting. |
| `src/plot.c` | Drawing commands, line types, curves, factor, rotation, window, clipping, `plot_to` pipeline. |
| `src/device.c` | `F*`, `TT`, `TB` (registration marks), service commands, built-in demo. |
| `src/subsys_E.c` | Vectors, startup, cutter-offset compensation, registration geometry, stroke font, USB descriptors and requests. |
| `src/subsys_F.c` | `main`, power, thermal model, EEPROM driver, solenoid, stepper queue, panel. |
| `src/subsys_G.c` | Device transform, motion planner, mark sensor; runtime library map (soft-float, math). |

## Ghidra processor module

Ghidra has no H8 support. `ghidra/H8S` starts from the H8/300H module by Shiz
(<https://codeberg.org/shiz/ghidra-h8-300>, WTFPL), with these changes:

* **H8S instructions added** in `h8s.sinc`:
  * `STM`/`LDM`
  * `STC`/`LDC`/`ANDC`/`ORC`/`XORC` on `EXR`
  * shifts and rotates by 2
  * bit operations on `@aa:16` and `@aa:32`
* **Compiler idioms modelled** so stack tracking works: `sub.w`/`add.w #imm,r7` adjusts the
  full SP, and `add.w #-n,rN` is treated as forming `&local`.
* **Upstream bugs fixed:**
  * the inverted mask in the `setb` macro
  * `rotr.w`/`rotr.l` mislabelled as `rotl`
  * the register field in `mov.l ERs,@(d:24,ERd)`
  * the function-start pattern, which matched `pop` instead of `push.l er6`
* **Branch conditions rewritten** so the decompiler can fold signed compares.
* **New compiler spec** (`h8s.cspec`) for the Renesas H8S C compiler: arguments in ER0 and ER1,
  then on the stack; result in ER0; ER2..ER6 preserved.

To rebuild:

```
ln -s "$PWD/ghidra/H8S" ~/Library/ghidra/ghidra_<version>_PUBLIC/Extensions/H8S
ghidra/build.sh            # GHIDRA=/path/to/ghidra if not Homebrew
```

## Caveats

* Nothing here has been checked on hardware.
* Statements marked TODO in the notes or sources are inferences.
* The default user unit (0.05 mm) is a factory setting in EEPROM, inferred from consistent use.
  The firmware itself supports 0.1 / 0.05 / 0.025 / 0.01 mm.
