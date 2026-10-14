# CC100 firmware 2.30: control, query and interface commands

Scope: control/query/interface handlers reached from `gpgl_loop` (0x0cbf2).
All addresses are ROM addresses in `cc100ver230.bin`.

## Common machinery

### Parameter parsing

* `peek_byte` (0x11228) / `get_byte` (0x11368): peek at and consume the next input byte.
* `cmd_class(c)` (0x11936): -1 for bytes that cannot start a command:
  digits, `+ , - .`, control characters and space. Otherwise it returns a class number:
  `!`..`*` give 0..9, `/` gives 10, `:`..`_` give c-0x2f.
  The dispatcher silently skips class -1 bytes. That is why stub commands (no-ops)
  never desynchronise: their numeric arguments are just dropped.
* `read_digit` (0x1199e): skips NUL bytes. It returns the digit value (consumed),
  -3 for `.` (consumed), -2 for a terminator or command character (not consumed),
  and -1 for any other separator.
* `read_number` (0x11a36): skips separators. A `-` seen while skipping makes the
  number negative; `+`, `,`, space and CR are ignored. It reads integer digits,
  then optionally `.` and up to two fraction digits into `frac_digit1`/`frac_digit2`
  (0xffb360/0xffb362); further fraction digits are dropped. The magnitude is clamped
  to 0x7fffff (positive) or 0x800000 (negative). On overflow it calls `set_error(8)`.
  A missing number (a terminator or command letter comes next) returns **0 without
  consuming anything**.
* `read_int` (0x11b72): `read_number`, then rounds half away from zero
  using the first fraction digit.
* `read_coord` (0x11bc6): converts to internal steps using `user_unit` (0xffb262),
  which takes values 20, 10, 5 or 2 (table at 0x28d28, chosen by the service command `d`).
  The fraction is honoured to that resolution.
* `read_fixed100` (0x11d50): value x 100 plus two fraction digits (hundredths).
* `end_of_params` (0x0cb3a): skips whitespace and `,`. It returns 1 when the next byte
  is a terminator, ESC/FS, `/`, 0x80 or a command character, and 0 when another number follows.
  Optional trailing parameters are read only when this returns 0.

### Terminator (`=` command)

* `is_terminator(c)` (0x11902) compares c with `term1` and `term2` (0xffb359/0xffb35a).
* Both are set to **0x03 (ETX)** only by `init_terminators` (0x131c0). Nothing else writes them.
* The `=` command (0x3d) is a **no-op** in the dispatcher, so terminators cannot be changed.
  Note: the two parameter bytes of `=t1,t2` then go through the dispatcher themselves.
  This is harmless if they are control characters (class -1). If they are command
  characters, they would be executed as commands.

### Replies

* `reply_begin` (0x113e4): waits until the previous reply has been sent (`0xffb357`),
  then resets the output buffer (`0xffb2ea` length, buffer at 0xffb2f3).
  Its arguments (field count, command char) are ignored.
* `reply_number(width, value)` (0x13252) formats a signed decimal number
  right-aligned in a field of `width` characters, padded with spaces:
  * zero: `"0"` padded to `width`;
  * non-zero: a sign character (`-` for negative, **space for positive**) sits directly
    before the digits. The field widens if sign plus digits exceed `width`,
    so a positive number always has at least one leading space;
  * at most 10 digits.
* `reply_field(w,v)` (0x133be) = number followed by `,`.
* `reply_last(w,v)` (0x133ee) = number, then `term1` (and `term2` if it differs),
  then `reply_send`. **Every reply therefore ends with a single 0x03.**
* `reply_send` (0x114ba) starts transmission on the active interface
  (USB at 0xc000c4, or SCI via 0xffff7a when 0xffbf0f is set).
* `reply_point(cmd, x, y)` (0x17648): `x/user_unit` (width 6), `,`, `y/user_unit`
  (width 6), ETX. Division truncates toward zero (`__divsl` 0x23052).

Example: point (100, -5) is sent as `"   100,    -5\x03"`.

### Errors

* `set_error(bits)` (0x10308): `err_status |= bits & err_mask`, where `err_mask` is the
  low word of the long at 0xffb282, i.e. 0xffb284, and `err_status` is the byte at 0xffb286.
* Callers: unknown command letter (`set_error(1)`, at 0xce94 and 0xd404) and
  number overflow (`set_error(8)`, at 0x11b3c).
* **Nothing reads `err_status`**: no command reports it. Status word 1/3 (`V`, `#`) are not implemented.

### Non-volatile parameters

`nv_read(addr)` (0x127d8) and `nv_write(val, addr)` (0x12786) access the parameter
store at 0x300000+n. Setters write only when the value changed.

| NV addr  | meaning                             |
|----------|-------------------------------------|
| 0x300034 | speed (`!`)                         |
| 0x300035 | force level (`*`)                   |
| 0x300036 | cutter offset (`FC` first param)    |
| 0x300038 | `FE` flag                           |
| 0x30003b | thickness (`FX`; `*` resets it to 0)|

### Per-pen tables

Each table has 8 entries, but every setter below writes **all 8 entries**.
The optional pen-number argument is parsed and then ignored.
After a change, `pen_params_dirty` (0xffb2be) is set to 1 and `apply_pen_params`
(0x18020) applies entry `cur_pen-1` (0xffb414).

| table   | type    | set by |
|---------|---------|--------|
| 0xffb416 | int[8]  | `!` speed |
| 0xffb436 | int[8]  | `*` force level |
| 0xffb446 | int[8]  | `FC` p (offset) |
| 0xffb456 | int[8]  | `FC` q |
| 0xffb466 | long[8] | `FD` angle x100 |
| 0xffb486 | int[8]  | `FE` |
| 0xffb496/0xffb4a6 | int[8] | `FF` a / b |

`apply_pen_params` programs velocity `speed*10` (then x10 inside 0x19780,
together with constant 5000, probably acceleration). `set_force_level` (0x180ec) sets the force.
If `thickness` (0xffb122, set by `FX`) is less than 1, the force comes from the
`*` level (1..3) via 0x181da. Otherwise it is interpolated between calibration points
0xffc37e/0xffc380/0xffc382, and the level becomes 1 (<7), 2 (<19) or 3.

## Commands

| Cmd | Handler | Syntax | Behaviour |
|-----|---------|--------|-----------|
| `!` | 0x074bc `cmd_speed` | `!l[,n]` | `l` is rounded to an integer; `l>=100` gives `l-100` (accepts the 101..110 form). Valid 1..10, otherwise ignored. `n` is optional and ignored. Saves NV 0x34 and sets all speed entries. |
| `*` | 0x07570 `cmd_force` | `*a,f[,n]` | `a` is read and ignored (the `-100` adjustment is computed but never used). `f` valid 1..3, otherwise the command is ignored. Resets `thickness` (and NV 0x3b) to 0, saves NV 0x35, sets all force entries, applies speed and force. |
| `"` | 0x102d4 `cmd_error_mask` | `"m[,x,y]` | Error mask = `m & 0xffff`. Two more ints are read and dropped (0 if absent). |
| `$` `I` `K` `P` `Q` `R` `S` `A` `T` `N` `(` | stubs | – | **No-ops** (empty functions). Arguments are discarded by the dispatcher's skip logic. `P`/`K` text, `$` font, `S` size, `Q` spacing, `R` rotation, `I` italic, `A` alpha reset, `T` buzzer. |
| `LP` `RC` `SP` `SA` | 0x0fff8, 0x10022, 0x1008e, 0x10052 | – | **No-ops**. The second letter is consumed. |
| `V` `:` `;` `<` `=` `'` | – | – | **No-ops** in the dispatcher (`V` = status word 1, `:` clear, `;` interface clear, `=` terminator). |
| `F` / `FS` | 0x13c06 `cmd_chart_feed` | `Fl` / `FSl` | `l` is read as a coordinate into `chart_feed_len` (0xffb3aa). The follow-up hook 0x1005e is empty, so there is no movement. For `FS`, the dispatcher does **not** consume `S`, but `S` is then read as a separator. |
| `?` | 0x1780e `cmd_read_offset` | `?` | Replies `reply_point(offset_x 0xffb4e0, offset_y 0xffb4e4)`. |
| `[` | 0x177b0 | `[` | Replies lower-left of the user window (0xffb564/0xffb568). The dispatcher calls hook 0x0ffbc first. |
| `U` | 0x17752 | `U` | Replies upper-right of the user window (0xffb56c/0xffb570). |
| `@` | 0x1769e `cmd_read_status2` | `@` | Replies one number (width 5), `cur_outcode & 0x2f`, then ETX. The outcode comes from `clip_outcode` (0x1e7a8): bit0 x>xmax, bit1 x<xmin, bit2 y>ymax, bit3 y<ymin against the hard limits 0xffb554..0xffb560. Flags 0xffb400/0xffb3b9 suppress bit0 or bit2. Bit 5 is never produced. |
| `G` | 0x176d6 `cmd_gin` | `G` | After the motion queue is flushed (0x0d458 → 0x1db52) and the position is updated (0x17f1e), replies `x,y,p`: current pen position 0xffb584/0xffb588 divided by `user_unit` (width 6 each), then `p = cur_tool*10 + (pen up ? 1 : 0)`. Here `cur_tool` is 0xffb4b8 and pen-up is bit5 of 0xffbc46 clear. All fields are width 6, followed by ETX. |
| `C` | 0x17740 `cmd_call_gin` | `C` | Same as `G` (calls 0x176d6 with 'G'). No wait-for-button. |
| `SO` / `NP` | 0x13ffa → 0x14008 | `SOn` | Takes no parameters (the digits are skipped). Flushes motion, then sets the carriage origin: x is 0, or 0xffb298 when 0xffb288 is non-zero (TODO: likely the registration-mark origin); y is unchanged (0xffbc42). See `set_carriage_origin` (0x1dacc). |
| `TT` | 0x08002 `cmd_TT` | `TT` | Full reset of motion: flush/stop, orientation set to 1 (0x7aa4(1)), reload pen parameters (0x13436), clear cutter offsets, 0x1054c, then `home_carriage` (0x14ef0: seeks the home sensor at port 0xffffb6 bit3 and sets origin (0,-2720)). Finally reloads the cutter offset from NV 0x36 (0..45, else 0). |
| `FC` | 0x0a268 | `FCp,q[,n]` | `p` 0..45, otherwise ignored: saves NV 0x36 and sets all `pen_cutter_offset` entries. `q` 1..3, otherwise ignored: sets all `pen_cutter_q` entries. `n` is read and ignored. |
| `FD` | 0x0a316 | `FDθ[,n]` | `θ` is read with hundredths (0..600.00, value x100 ≤ 60000), otherwise the command is ignored. `n` must be 0..8 (default 0), else the command is ignored. Sets all `pen_blade_rot` entries. |
| `FE` | 0x0a3cc | `FEp[,n]` | `p` must be 0 or 1. Saves NV 0x38 and sets all `pen_FE` entries. Meaning unknown (TODO). |
| `FF` | 0x0a46e | `FFa,b[,n]` | `a`, `b` 0..9 (both required). `n` 0..8 (default 0) is checked. Sets all `pen_FF_a`/`pen_FF_b` entries. Meaning unknown (TODO). |
| `FK` | 0x0a580 | `FK[n]` | `fk_flag` (0xffb2bc) = `n != 0` (default 0). Meaning unknown (TODO). |

Out of scope but seen: `FQn` (0x07d46) reads NV values: 0 = speed (1..10, default 10),
1 = force (1..3, default 3), 2 = cutter offset (0..45), 3 = NV 0x38, 4 = NV 0x37,
5 = byte 0xffb0aa. Each is sent as one width-5 number plus ETX.
