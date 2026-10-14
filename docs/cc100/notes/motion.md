# CC100 fw 2.30: motion layer, speeds and force

This note covers the motion layer below `motion_move_steps` (0x1d582): the speed profile, the
planner, the segment queue, the stepper interrupts, jogging and the solenoid (tool force).
Readable C is in `../src/motion.c`. RAM names are in `../sym/motion.txt`.

## 1. Units and clocks

| Quantity | Value | Evidence |
|---|---|---|
| CPU clock φ | **16 MHz** | The profile builder hard-codes 16 000 000 timer counts per second (ROM double at 0x28fba). The 1 ms tick (TMR0, φ/64, compare 250) is consistent with it. |
| Stepper timers | TPU channels at 0xFFFF20 (X) and 0xFFFF30 (Y) | `motion_init` writes TCR = 0x20: counter clears on TGRA compare, clock φ/1. Each count is therefore 62.5 ns. The compare value goes to TGRA (0xFFFF28 / 0xFFFF38), the interrupt enable to TIER (0xFFFF24 / 0xFFFF34; 0x41 = run, 0x40 = stop), and the status to TSR (0xFFFF25 / 0xFFFF35). |
| Motor count | **0.05 mm**, one half-step | The planner gets internal units / 10. Each step advances the phase index by one in the 8-state table `step_phase_table` (RAM 0xffc3a2, 32 entries = 4 × {1, 0x15, 4, 0x16, 2, 0x1a, 8, 0x19}). The value is written to latch 0x600008 (X) or 0x600000 (Y). Bit 4 in the odd entries marks the two-coil half-steps. The profile maths uses 20 counts per mm. |
| Speed unit | **mm/s** | `apply_pen_params` passes `pen_speed × 10`, i.e. `!` level × 10, to `motion_build_profile`. The builder treats it as mm/s together with 0.05 mm/count and 16 MHz. |
| Acceleration | **3000 mm/s²** (about 0.3 g) | Constant 3000 in `motion_build_profile`, stored in `accel` (0xffb7dc). |
| Start speed | 10 mm/s | `start_speed` (0xffb626), set by `motion_init`. Service `h n` changes it and rebuilds the table. |

Periods longer than 0xF06E counts (about 3.85 ms) do not fit the 16-bit compare register. They
are split into 60000-count chunks, and the axis steps only when the remainder runs out.

## 2. Speed profile table (`motion_build_profile`, 0x19780)

The argument packs the pen-down speed in its low word and the pen-up speed in its high word, in
mm/s.

* `vmax` is the larger of the two speeds, and `v0` is the smaller of the start speed and both speeds.
* `N = (vmax² − v0²)·10 / 3000` is the length of the ramp, in counts.
* If `N ≥ 40` the call is **rejected** (returns −1) and the old profile stays in force.
  The highest speed accepted is therefore about 109 mm/s.
* For every count `i = 0..N+8` the table `speed_ramp` (0xffb640) holds two longs:
  * `T[i] = round(16e6 · t(i))`. This is the time to reach the middle of count `i`, starting
    from `v0` at constant acceleration `A`:
    `t = sqrt((v0/A)² + (i+0.5)/(10·A)) − v0/A`. Since `(i+0.5)/(10A) = 2·s/A` with
    `s = (i+0.5)/20` mm, this is the standard constant-acceleration formula.
  * `P[i] = round(2.048e8 / (v0 + A·t))`. This is the step period at that speed in counts, ×256.
    One count at v mm/s lasts `16e6/(20·v) = 8e5/v` timer counts.
* The builder also keeps these values:
  * `ramp_steps_down` / `ramp_steps_up`: ramp length up to the pen-down / pen-up speed.
  * `ramp_time_down` / `ramp_time_up`: the matching `T[]` values.
  * `max_subseg_*`: speed × 1600, capped at 100 000 counts.
  * `corner_ramp_idx[k] = corner_speed_tab[k]²·10/3000`: the ramp index for each corner speed (§4).

The axis interrupts do not use `P[]`. They interpolate `T[]` at a fractional ramp position, using
10-bit fractions, and step with the difference between successive times. This keeps the
acceleration exact even when one axis moves fewer counts than the path.

## 3. Physical speeds

`!l` sets `pen_speed[] = l` (1..10). `apply_pen_params` (0x18020) then sets **both** the pen-down
and the pen-up speed to `10·l` mm/s. At power-on `motion_init` uses 100/100 mm/s.

| `!` | Speed (mm/s) | Step rate (half-steps/s) | Ramp length (counts / mm) | Ramp time (ms) |
|---|---|---|---|---|
| 1 | 10 | 200 | 0 / 0.00 | 0 |
| 2 | 20 | 400 | 1 / 0.05 | 3.3 |
| 3 | 30 | 600 | 2 / 0.13 | 6.7 |
| 4 | 40 | 800 | 5 / 0.25 | 10.0 |
| 5 | 50 | 1000 | 8 / 0.40 | 13.3 |
| 6 | 60 | 1200 | 11 / 0.58 | 16.7 |
| 7 | 70 | 1400 | 16 / 0.80 | 20.0 |
| 8 | 80 | 1600 | 21 / 1.05 | 23.3 |
| 9 | 90 | 1800 | 26 / 1.33 | 26.7 |
| 10 | 100 | 2000 | 33 / 1.65 | 30.0 |

Notes on the table:

* These are path speeds; each axis gets its share.
* The planner never spends more than a quarter of a segment on acceleration and a quarter on
  deceleration (§4).
* Short segments therefore do not reach the nominal speed.
* Jogging (§7) uses its own speed scale.

## 4. Planner

### 4.1 `motion_line_to` (0x19b9c)

Inputs are absolute counts and a pen state.

* **Pen change.** The queue is drained first: the pending point is planned and the function
  waits until the ring is empty.
* **Turn angle.**
  * The direction of the new move is computed with `atan_tenths` (in 0.1°).
  * The turn is |Δθ| folded into 0..180°.
  * An empty queue counts as a 180° turn.
* **Corner stop.** If the turn is above 25.0°:
  * When |dx| > 80 counts (4 mm), the planner drains, so the tool stops at the corner.
  * Shorter moves only kick the queue.
* **X backlash.**
  * `backlash_x` (0xffb620) comes from EEPROM 0x2a, service `g4`/`c`.
  * On an X direction reversal, the running offset `backlash_cur` moves toward 0 or
    `backlash_x`. A sharp turn changes it in one step; a shallow one changes it in steps of 1–2
    counts per move.
  * The offset is added to every X target.
  * A sharp reversal with |dx| < 2·|dy| and the pen down first queues a pure take-up move.
* **Profile selection.** Pen-down moves use `ramp_steps_down`, `ramp_time_down`,
  `max_subseg_down` and `speed_down`. Pen-up moves use the `_up` variants.
* **Planning and start.** It calls `motion_plan_segment(x + backlash, y, turn)`. For pen-up
  moves it starts the queue immediately if the motors are idle.

### 4.2 `motion_plan_segment` (0x19f8c)

It works on one segment from the plan position (`plan_pos_x/y`) to (x, y).

1. **Length.** `a ≈ max + min²·k`, computed without a square root:
   `f = r(r·365 + 15104)/65536 + 1024` with `r = 256·min/max`, then `a = max·f/1024`.
   The error is under 1 %.
2. **Splitting.** If `a > max_subseg_len`, the segment is cut into `a/max_subseg_len + 1` equal
   parts. This only happens for very long pen-up moves.
3. **Ramp.**
   * Ramp length `n = min(ramp_steps, a/4)`.
   * Cruise length `a − 2n`.
   * Duration `2·T[n] + (a − 2n)·P[n]/256` timer counts, computed in doubles at 0x1a27c (the ROM constant 256.0 is at 0x28fca).
4. **Per-axis records** (X and Y separately, in the 0x42-byte ring record):
   * direction;
   * axis ramp steps `n·axis/a`;
   * ramp rate `axis·1024/a` (path counts per axis count, 10-bit fraction);
   * cruise count;
   * cruise period with a 16-bit fractional accumulator;
   * an axis that does not move gets state 0.
5. **Force trim by direction.** This applies only when the cutter offset is active
   (`cutoff_radius` ≠ 0) and the tool is down. The D/A value for the segment is:
   ```
   level + (dir_force_trim[force_level][bucket] · level · 2) >> 8
   ```
   * `bucket = floor(direction / 15°) + 12`, clamped to 0..23, for direction −180°..180°.
   * The table is copied from ROM 0x28fd2 by `motion_init`. Row 1 is
     `0,0,0,0,0,0,0,0,−7,−13,−13,−23,−23,−13,−13,−7,0,...`; rows 2 and 3 have −26 in the middle.
   * Effect: when cutting toward +X (feed direction) within ±60°, the force is reduced by up to
     about 18 % (level 1) or 20 % (levels 2–3).
6. **X resonance avoidance.** For the first sub-segment that has an X cruise phase, the average
   X period is checked.
   * If it lies between 22 858 and 39 999 timer counts (X speed about 20–35 mm/s), the path speed
     is scaled down to `v·period/40000` (not below `start_speed`).
   * That brings X to 20 mm/s, and the segment is re-planned.
   * Y is not checked.
7. **Junction blending.** This applies when the queue is not empty and the turn is ≤ 25.0°.
   * The speed at the junction is
     `min(previous ramp index, this ramp index, corner_ramp_idx[turn/1.6°])`.
   * A 0x34-byte blend record (0xffb9f2 + i·0x34) shortens the previous segment's deceleration
     and this segment's acceleration, so the tool passes the corner without stopping.
   * Corner speeds (`corner_speed_tab`, RAM 0xffc3e2, in mm/s, per 1.6° of turn):

     | Turn | Corner speed (mm/s) |
     |---|---|
     | 0° | 200 (capped by the cruise speed) |
     | 1.6° | 120 |
     | 3.2° | 90 |
     | 4.8° | 70 |
     | 6.4° | 60 |
     | 8° | 50 |
     | 9.6° | 45 |
     | 11.2° | 40 |
     | 12.8° | 36 |
     | 14.4° | 33 |
     | 16° | 31 |
     | 17.6° | 29 |
     | 19.2° | 27 |
     | 20.8° | 25 |
     | 22.4° | 24 |
     | 24° | 23 |

8. **Queueing.** The record goes into the 8-entry ring (0xffb7e2, head `seg_head`, tail
   `seg_tail`). When the ring is full, the planner starts the next record or waits 5 ticks. With
   more than two records queued and idle motors, it starts the queue.

## 5. Segment start (`motion_start_next`, 0x1b162) and current control

* **Empty ring.** The motor current latch goes to 0x0A (both axes hold) and the function returns.
* **Otherwise:**
  * It copies the record into the two axis blocks (`axis_x` 0xffbf12, `axis_y` 0xffbf56) and
    applies the blend record of this junction and of the next one.
  * The axis with the shorter total time gets its first period shortened by half a cruise
    period, minimum 0x212 counts.
* **Motor current** (latch 0x600010, shadow 0xffb5de; bits 3..2 = X, 1..0 = Y), set per segment
  from the cruise period:

  | Axis state | Bits | Current |
  |---|---|---|
  | moving, period < 16 001 counts (> 50 mm/s) | 00 | full |
  | moving, period 16 001..80 000 (10–50 mm/s) | 01 | |
  | moving, period > 80 000 (< 10 mm/s) | 10 | |
  | idle | 10 | hold, unless flag 0xffb2c1 is set, then 00 |

  0x0F (both axes 11) is "off", used by power-off and the idle power save.
* **Start.** The first compare values are written, TIER is set to 0x41, both counters are
  zeroed, and the segment's D/A force value is written to 0xFFFDAC when it is non-zero.

## 6. Axis interrupts (`timer_x_isr` 0x1bc6a, `timer_y_isr` 0x1bcc4, `axis_profile_step` 0x14b5e)

* **Dispatch.** If the jog direction is set, the interrupt runs `jog_x_step`. Otherwise it clears
  TGFA and, unless a segment is being loaded, calls `axis_profile_step(&axis_x)`.
* **Y in the X interrupt.** If the Y interrupt flag is not pending, it also services
  `axis_profile_step(&axis_y)` directly, to keep the axes in lockstep.
* **Next segment.** When both axes are idle, it starts the next segment.

`axis_profile_step` is a state machine (`state` at +2 = `motor_busy_x/y`):

| State | Action per count |
|---|---|
| 0 | idle: TIER = 0x40 (stop) |
| 1 | accelerate: step, `ramp_pos += ramp_rate`, period = `T(ramp_pos) − T(previous)` with 10-bit interpolation. When the ramp count runs out, go to 2 (or 3 if there is no cruise). |
| 2 | cruise: step, constant period plus a 16-bit fractional accumulator. When the cruise count runs out, go to 3 (or 0 if there is no ramp). |
| 3 | decelerate: as state 1 in reverse (`ramp_pos −= ramp_rate`). When the ramp count runs out, go to 0. |
| 4 | constant speed to the end (used when blending removed the ramp) |

"Step" here means `phase += dir`, followed by writing `step_phase_table[phase & 31]` to the
latch. The phase counter is also the motor position (`motor_pos_x/y`, 0xffbf26/0xffbf6a).

## 7. Jogging (`jog_x_step` 0x1bd1e / `jog_y_step` 0x1bf32)

* **Target period.** `160000 / jog_speed` timer counts, or 160 000 when `jog_speed` is 0.
  That is 100·`jog_speed` half-steps/s, i.e. `jog_speed × 5` mm/s; the minimum is 5 mm/s.
* **Ramp.** It steps up and down the same `T[]` table one index per step (`jog_ramp_idx`,
  0xffb622), at most `ramp_steps_max + 4`.
* **Stop.** It stops when the speed is 0 and the ramp is back at index 1.
* **Current.** The motor current is chosen from the period, as in §5.

## 8. Tool force (solenoid)

The tool is pressed by a solenoid on the 8-bit D/A output 0xFFFDAC.

### 8.1 Force levels

`set_force_level(level)` (0x180ec) works as follows.

* **Without `FX`** (`cut_force` < 1):
  * the D/A value is `force_table[level]`, `level` = 1..3 from `*`;
  * entries 1..3 come from EEPROM words 0x08/0x0a/0x0c, factory 164/177/203;
  * the RAM defaults 180/190/200 are overwritten at boot;
  * entry 0 is 200.
* **With `FX F`:**
  * F < 15: `v = f1 + (f2 − f1)(F − 2)/8`;
  * F ≥ 15: `v = f2 + (f3 − f2)(F − 10)/17`;
  * division truncates toward zero;
  * the press-profile level is 1 for F < 7, 2 for F < 19, otherwise 3.

With the factory table (164/177/203):

| FX | D/A | FX | D/A | FX | D/A |
|---|---|---|---|---|---|
| 1 | 163 | 11 | 178 | 21 | 193 |
| 2 | 164 | 12 | 180 | 22 | 195 |
| 3 | 165 | 13 | 181 | 23 | 196 |
| 4 | 167 | 14 | 183 | 24 | 198 |
| 5 | 168 | 15 | 184 | 25 | 199 |
| 6 | 170 | 16 | 186 | 26 | 201 |
| 7 | 172 | 17 | 187 | 27 | 203 |
| 8 | 173 | 18 | 189 | 28 | 204 |
| 9 | 175 | 19 | 190 | 29 | 206 |
| 10 | 177 | 20 | 192 | 30 | 207 |

`*1`, `*2` and `*3` correspond to `FX2`, `FX10` and `FX27`.

### 8.2 `solenoid_set(level)` (0x17c1c)

**Media trim.** This applies only without `FX` and outside the blade alignment stroke:
`level += level·media_force_pct[media]/100`. The table at 0xffc385 is all 0 except media
types 7 and 8 (−2 %).

**Already down.** Only the D/A value changes.

**Press** (tool was up):
* USB receive is blocked and the interrupt mask set.
* A kick profile is played: five (ticks, D/A) steps from `press_profiles[sel][force_level−1]`,
  where `sel` comes from EEPROM 0x2c (service `g7`). Factory profiles:

  | `sel` | Steps |
  |---|---|
  | 0 | (5 ms, 220) (60 ms, 155) (10 ms, 170 / 190 / 220 for levels 1 / 2 / 3) |
  | 1 | same, with 160 in the middle step |

* Then the hold level is set and the press settle delay runs (EEPROM 0x2e, default 20 ms).

**Lift** (`level` = 0):
* An optional pre-lift delay runs (EEPROM 0x30).
* The lift profile is played (`lift_profile`: 1 ms at D/A 1).
* The D/A goes to 0, then the lift settle delay runs (EEPROM 0x32, default 70 ms).

**Per-segment trim.** The direction-dependent trim of §4.2 step 5 replaces the D/A value at each
segment start.

### 8.3 Thermal model

The firmware also keeps a thermal estimate: the solenoid integrator grows by (D/A − 180)/10 per
interval (see `subsys_F.md`, power).

## 9. Open points

* **Junction blend arithmetic.** The arithmetic in the 0x34-byte blend records is reproduced
  structurally in `motion.c`; exact rounding should be checked against `decompiled.c`.
* **Clock frequency.** 16 MHz is what the firmware assumes in its arithmetic. The crystal was not
  checked on hardware.
