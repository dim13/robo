# CC100 firmware 2.30: input layer and parser core

All addresses refer to `cc100ver230.bin`, loaded at 0. CPU: H8S, advanced mode. `int` is 16 bits and `long` is 32 bits.
The reconstructed C is in `../src/input.c`, and the symbols are in `../sym/input.txt`.

## 1. Byte sources

All bytes go into one receive ring buffer:

| var        | meaning                                              |
|------------|------------------------------------------------------|
| `0xffb2d6` | `rx_buf_base` = `*0xffb0e8` = `0xffc800`             |
| `0xffb2da` | `rx_buf_size` = `*0xffb104 - 4` (the 6 KB pool at 0xffc800 is shared with other buffers, see `0x6aac`/`0x6ba8`) |
| `0xffb2de` | `rx_wr_ptr` (written by ISRs)                        |
| `0xffb2e2` | `rx_rd_ptr` (parser)                                 |
| `0xffb2e6` | `rx_count`                                           |

`rx_init` (0x11882) resets the buffer and the TX state. `rx_flush` (0x1181c) and `rx_flush_masked` (0x11852, which raises the interrupt mask first) empty it. They are called on USB bus reset/resume (0x6c3c) and by 0x21cd6/0x21d42.

### 1.1 USB (external USB controller at 0xC00000)

`usb_isr` (0x220b2) is vector 104. It reads the interrupt status `0xC000C0`, ANDs it with the enable shadow `0xffc426`, and dispatches through a table at **0x290d2**. Each entry is 10 bytes: mask(4), endpoint(1), ?(1), handler(4).

| mask       | handler  | role                                   |
|------------|----------|----------------------------------------|
| 0x80000000 | 0x2288c  | ?                                      |
| 0x01000000 | 0x0f942  | EP0 setup                              |
| 0x0e000000 | 0x0f9f8/0x0fa9c | EP0 in/out                      |
| 0x00040000 | **0x11556** | **bulk OUT (host -> plotter data)** |
| 0x00010000 | 0x0f78a  | bulk IN (replies)                      |
| 0x00020000 | 0x0f8ee  | ?                                      |
| 0x00000080 | 0x223f4  | ?                                      |
| 0x00000001 | 0x228c0  | ?                                      |

`usb_bulk_out_rx` (0x11556) works as follows:
- If `rx_count > rx_buf_size - 100`, it disables the bulk-OUT interrupt (bit 18 of `0xC000C4`) and returns. The host is NAKed. `usb_rx_flow_check` (0x11796, called every timer tick) sets the enable bit again when `rx_count < rx_buf_size - 110` and `0xffb2f2` (motor/port busy) is clear. TODO: it also requires bit 18 of the shadow `0xffc426` to be set; how the shadow differs from the register here is not fully understood.
- Otherwise it reads `count = *(u8*)0xC000BD` bytes from FIFO `0xC000A4`.
- **ESC sequences are handled here, in the USB path.** ESC and the byte after it must be in the same packet:

| bytes         | action |
|---------------|--------|
| `ESC 00 n`    | Calls `idle_timer_reset2`, then `jog_keys` (`0xffc424`) = n. These are "remote arrow keys". The bits are active-low: bit0=0 → +X, bit1=0 → −X, bit2=0 → +Y, bit3=0 → −Y. While the plotter waits for data and `jog_keys != 0`, it jogs the head (`0x1edd6` via `0x1ed42`). n=0 stops jogging (it is checked as "nonzero"). |
| `ESC 01`      | Acks the FIFO, then `trapa #1`. The vector 9 handler is empty (`rte`). Execution then falls through and stores **0x01** into the data stream. This is effectively a debugger hook. |
| `ESC 02`      | Immediate reply `"1"` if USB is not suspended (`0xffb397 == 0`), otherwise `"0"`, followed by the terminator(s). |
| `ESC 03`      | `cancel_request` (`0xffc226`) = 1. Plotting loops (0x1575a, 0x15fb2, 0x20d6a, 0x2114c, 0x2135c, 0x21620, 0x21b96) poll this flag and abort the job. |
| `ESC 04`      | Consumed, no effect. Hosts send it as "initialize". |
| `ESC 05`      | Immediate status reply: `"1"` if the parser is **not** starved for input (`rx_starved == 0`, i.e. busy or holding data), and `"0"` if the parser is idle waiting for data. The terminator(s) follow. |
| `ESC x` (other) | ESC is dropped and `x` is stored as data. Bug: only one byte is subtracted from the packet count while two were read. |

- While `usb_suspended` (`0xffb397`) is set, all other bytes are dropped.
- At the end it sets bit 5 of `0xC00084`, which acks/re-arms the OUT FIFO.

Immediate replies (`ESC 02`/`05`) bypass the parser. They write `tx_buf[0]`, append the terminator(s), and call `usb_reply_send` (0x1145a).

### 1.2 RS-232 (SCI0: SMR..RDR at 0xFFFF78..0xFFFF7D)

The SCI0 vectors are 80 ERI, 81 RXI, 82 TXI and 83 TEI.

`sci0_rxi_isr` (0x1f4a4). If `0xffbf0e` is set, it goes to an empty stub (0x1f62c): an alternative input path that is unused. Otherwise it reads RDR and clears RDRF. It then runs a small ESC state machine in `sci_esc_state` (`0xffbf11`):
- state 0: if the byte is ESC, go to state 1. Otherwise set `reply_via_sci = 1` and store the byte in the ring buffer.
- state 1: byte 00 goes to state 2. Byte 02 sets `cancel_request = 1` and returns to state 0. **Any other byte is swallowed and the state stays at 1** (firmware bug: the machine stays stuck until 00 or 02 arrives).
- state 2: `jog_keys` = byte, back to state 0.

So the serial line supports only `ESC 00 n` (jog) and `ESC 02` (cancel). There is no status query on serial. Unlike the USB path, the serial path has no flow control on the ring buffer: no RTS handling and no overflow check (TODO: verify hardware handshake).

`sci0_eri_isr` (0x1f592) stores ORER/FER/PER in `0xffbf10` and clears them. `sci0_txi_isr` (0x1f5c6) sends `tx_buf[0xffbf0c++]` until `tx_len`, then clears `tx_busy` and TIE.

## 2. Reading bytes

- `rx_peek` (0x11228) returns `*rx_rd_ptr` without consuming. If the buffer is empty:
  1. It sets `rx_starved = 1` and polls 100 × `delay_ticks(10)`.
  2. Still empty: it finishes the pending motion (`0x1108a`, `0x17f2e`, `0x19acc`, `0x1db52`), sets `reply_via_sci = 0`, and re-runs the ROM demo plot if `0xffc2da` is set (0x0d512).
  3. It polls 200 × `delay_ticks_jog(10)`. This waits and also services `jog_keys`.
  4. If config bit `0xffb368.5` is 0, it calls `0x1d6f8(0)`. TODO: probably tool-up/park.
  5. It sets the LED/port shadow `0xffb5de |= 0x0c` and writes it to `0x600010` (unless jogging), then waits forever in `delay_ticks_jog(1)` steps.
  When a byte becomes available, it clears `rx_starved` and resets the idle (power-save) timer.
- `rx_getc` (0x11368) is `peek`, then `rx_count--`, `rx_rd_ptr++`, with wrap.

The ROM demo is `0x0d512`. It draws a calibration pattern, then points the ring buffer at **0xdb9a** with length **0x14fc** (a GP-GL program stored in ROM, starting `FU5440,4000FM1TB50,0FO5440&100,100,100,\0,0,Z5940,4200,L100,1,40,80,^0,100M...`) and jumps into `gpgl_loop`. This makes it a good real-world sample of the protocol.

## 3. Main loop `gpgl_loop` (0xcbf2)

```
if (0x127d8()) 0x20268();
0xffb358 = 0; 0xffb3ae = 0;
for (;;) {
    c = getc();
    if (c == ESC || c == FS) { getc(); continue; }   // 2-byte escapes in stream are skipped
    if (service_mode && 'a' <= c <= 'z') { lowercase service command; continue; }
    last_cmd = c;
    while (c == 0x80) c = getc();                    // 0x80 = filler, skipped
    if (c > 0x5f) set_error(1);                      // incl. lowercase when not in service mode, DEL, 8-bit
    if (cmd_class(c) < 0 || c == '#') continue;      // silently ignore
    dispatch(c)  // see notes/dispatch.txt
}
```

- `cmd_class` (0x11936) returns -1 for everything except `!`..`*` (0x21–0x2a), `/` and `:`..`_` (0x3a–0x5f). So **digits, `+ , - .`, space, all control characters (including the terminator) and bytes ≥ 0x60 are ignored between commands**. `#` is a command character but is ignored explicitly.
- Two-letter commands: the first letter is consumed and the second is only peeked. If `(c<<8 | peek)` matches the list, the second letter is consumed. Otherwise the single-letter fallback runs with the peeked byte still unread.
- Unknown command letters fall through to `set_error(1)` and are ignored.
- **Service mode**: `FP` (0xd46c) reads exactly the 6 bytes `GRFCC1` (the full command is `FPGRFCC1`). On a match it sets `service_mode` (`0xffb22e`) = 1, which enables the lowercase a..z commands. It is cleared by 0x137c4 (re-entry into `gpgl_loop` after restart). A mismatch consumes bytes up to the first wrong one, and that wrong byte is lost too.

## 4. Parameters and numbers

### 4.1 `end_of_params` (0xcb3a): is there another number?

It loops on `peek()`:
- terminator (`is_terminator`) or 0x80 → **1** (end)
- c &= 0x7f
- `,` → consume, continue
- `/` → 1 (it is a command)
- `+ - . 0-9` → **0** (a number follows)
- ESC or FS → 1
- any other printable (> 0x20) → 1 (next command letter)
- control characters and space → consume, continue

The terminator is **not** consumed. The main loop later drops it as an ignored control character. Hence a list such as `D100,100,200,200` stops at the next command letter **or** at the terminator, and the terminator is optional between commands.

### 4.2 `read_digit` (0x1199e)

It skips NUL bytes. It returns 0..9 (consumed) for a digit, -3 for `.` (consumed), and -2 (not consumed) for a terminator or command character. For anything else it returns -1, not consumed.

### 4.3 `read_number` (0x11a36): the number syntax

1. Skip any characters for which `read_digit` returns -1, consuming each one. If one of them is `-`, the number becomes negative. So `+`, spaces, commas and control characters before the digits are skipped, and a `-` anywhere in that run negates the value.
2. If the next character is a terminator or command letter, the value is **0**. Empty parameters are allowed: `M,5` gives x=0.
3. Integer digits: accumulate in decimal while `acc <= limit`. Later digits are consumed but ignored. The limit is **8 388 607** (0x7fffff) for positive values and **8 388 608** for negative ones.
4. If a `.` followed: the first fractional digit goes to `frac_d1` (`0xffb360`), the second to `frac_d2` (`0xffb362`), and the rest are consumed and ignored.
5. If `acc > limit`, it calls `set_error(8)` and clamps.
6. It applies the sign. The fraction digits are **unsigned** (they hold the magnitude).

There is no exponent and no sign after digits. The number ends at the first non-digit, which is left unread.

### 4.4 `read_int` (0x11b72)

It calls `read_number` and rounds half away from zero using `frac_d1 >= 5`. Quirk: `-0.5` parses as integer 0 with fraction .5. Since 0 is not negative, it rounds to **+1**.

### 4.5 `read_coord` (0x11bc6): coordinates in internal steps

```
n = read_number();  s = steps_per_unit (0xffb262)
s==10: v = n*10 + d1 + (d2>=5)
s==5 : v = n*5  + (d1+1)/2
s==2 : v = n*2  + (d1>7 ? 2 : d1>2 ? 1 : 0)       (d2 overwritten with d1)
else : v = n*20 + 2*d1 + (d2>7 ? 2 : d2>2 ? 1 : 0) (s==20)
v += (n<0) ? -frac : +frac
```
(The last line restates the code: the fraction term is subtracted when n < 0 and added otherwise. It is folded into the cases above.)

The same `-0.x` quirk applies: a negative value with a zero integer part loses its sign.

`steps_per_unit` comes from the table at **0x28d28 = {20, 10, 5, 2}**, indexed by:
- `load_step_size` (0x12264): bits 3..2 of the config byte `0xffb368`, which is persisted in settings.
- the service command `d<idx>`, idx 0..3.

The four entries match the classic Graphtec "step size" setting of 0.1 / 0.05 / 0.025 / 0.01 mm per GP-GL unit, with an internal resolution of 0.005 mm. This is an inference from the ratios. The default setting gives 20 steps/unit, so 1 GP-GL unit = 0.1 mm. Check against the step-size handling in the motion code (other notes).

`read_point` (0x13ab2) is `x = read_coord(); y = read_coord();`.

## 5. Terminators

`term1` (`0xffb359`) and `term2` (`0xffb35a`) are both set to **ETX (0x03)** by `init_terminators` (0x131c0). That function is called from 0x102ae when its argument is 0, as part of an init/reset path. **Nothing else writes them: the `=` command is a no-op in this firmware**, so the terminator is always ETX. `is_terminator(c)` (0x11902) compares against both.

## 6. Replies

- `reply_begin` (0x113e4) waits until the previous transmission is done (`tx_busy`), then resets `tx_len` and `tx_buf[0]`.
- `reply_putc` (0x1141e) appends to `tx_buf` (`0xffb2f3`) and keeps it NUL-terminated.
- `reply_number(width, v)` (0x13252) formats decimal, right-aligned to at least `width` characters and space-padded. Non-zero values always carry a sign character: `' '` for positive, `'-'` for negative. Zero is just `0`. At most 10 digits. Examples with width 5: `0` → `"    0"`, `123` → `"  123"`, `-7` → `"   -7"`, `123456` → `" 123456"`. Widths used in the firmware: 5 (most), 6, 8 and 10.
- `reply_number_comma` (0x133be) formats the number and appends `,`.
- `reply_number_end` (0x133ee) formats the number, appends `term1` (plus `term2` if different, which never happens), then sends.
- `reply_send` (0x114ba) sends on the interface the last data came from. If `reply_via_sci == 0`, it uses USB bulk IN (sets enable bit 16 of `0xC000C4`). Otherwise it enables SCI0 TXI (SCR |= 0x84).

So a typical reply looks like `"  100,  200,    0" ETX`.

## 7. Errors

`set_error(bits)` (0x10308) sets `error_status |= bits & error_mask`.

| bit | value | raised by |
|-----|-------|-----------|
| 0   | 1     | unknown command character, or character > 0x5f outside service mode (0xce94, 0xd404) |
| 3   | 8     | numeric overflow in `read_number` (0x11b3c) |

There are no other callers. `error_mask` (`0xffb284`, 16-bit) is the low word of the 32-bit store at `0xffb282` done by the `"` command (`cmd_error_mask`, 0x102d4). That command reads three integers, keeps the first as the mask and discards the other two.

**`error_status` is never read anywhere.** The `@` status reply (0x1769e) reports `0xffb5ca & 0x2f` instead. The GP-GL error machinery is therefore vestigial. It defaults to mask 0, so nothing is recorded.

## 8. Timing helpers

- `timer_tick_isr` (0x12d7a, vector 64, 8-bit timer compare). It decrements `delay_counter` (`0xffb37c`) and `0xffb37e`. It counts idle ticks (`0xffb380`/`0xffb382`; when the idle minutes reach `0xffb388` it writes 0xf to port `0x600010`, a power-save LED/port). It calls 0x12b18 every 100 ticks, blinks the busy LED (P? bit 3 at `0xffffff09`) every 500 ticks while data is buffered, and calls `usb_rx_flow_check`. Tick length not determined (TODO, likely 1 ms).
- `delay_ticks(n)` (0x12e66) is a busy-wait. `delay_ticks_jog(n)` (0x12e88) waits the same way but runs a jog step every 5 ticks while `jog_keys != 0`.
- `idle_timer_reset` (0x12a88) and `idle_timer_reset2` (0x12ab4) clear the idle counters.

## 9. Runtime library used by the parser

| addr    | name            | semantics |
|---------|-----------------|-----------|
| 0x2381e | `__mulsi3`      | er0 = er0 * er1 (32-bit) |
| 0x23052 | `__divmodsi4`   | signed: er0 = quotient, er1 = remainder (remainder takes the dividend's sign) |
| 0x23078 | `__udivmodsi4`  | unsigned version |
| 0x23836 | `__dcopy`       | copy 8 bytes (double assignment) |
| 0x23854 | `__movblk`      | memcpy(dst=er1, src=er0, n=er2) |
