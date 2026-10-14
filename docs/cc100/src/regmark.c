/*
 * CC100 firmware 2.30 -- registration marks ("TB" commands), end to end.
 *
 *   1. mark-sensor driver (LED, digital pot, A/D, adaptive edge filter)
 *   2. scan primitives (one axis, diagonal, bar-centre locate)
 *   3. mark search (manual TB2x / automatic TB12x, mark types 1 and 2)
 *   4. geometry (corners from bar samples, rotation/skew/scale transform)
 *   5. use of the transform in the device pipeline (area clip, rotate)
 *   6. TB command handlers, TB70 calibration, service a/n/t sensor tests
 *
 * See ../notes/regmark.md for the physical procedure and host protocol.
 *
 * Units: all positions are internal device units, 1 unit = 0.005 mm
 * (200 units = 1 mm).  regmark_length is in whole mm.  The firmware does
 * every floating-point step through the Renesas soft-float runtime
 * (__ltod/__dmul/__ddiv/__dtol, doubles passed by pointer); it is written
 * here as native double arithmetic.  (long) conversions truncate.
 *
 */
#include "cc100.h"

/* ---- local declarations (not in cc100.h) ---- */

#define REG8(a) (*(volatile unsigned char *)(a))
#define P1DDR_BIT  REG8(0xfffffde8)   /* bit 1 cleared by sensor_init */
#define POT_PORT   REG8(0xffffff00)   /* bit1 data, bit2 clock, bit3 /CS of LED-gain pot */
#define LED_PORT   REG8(0xffffff06)   /* bit4: sensor LED */
#define STAT_LED   REG8(0xffffff09)   /* bit3: front-panel LED (active low) */
#define ADDRAH     REG8(0xffffff90)   /* A/D result, 10 bit left-aligned */
#define ADDRAL     REG8(0xffffff91)
#define ADCSR      REG8(0xffffff98)   /* 0x20 = start single, 0x30 = start scan, bit7 = ADF */
#define ADCR       REG8(0xffffff99)
#define ADC_VALUE() ((unsigned)((ADDRAH << 8) | ADDRAL) >> 6)   /* 0..1023 */

/* regmark_cfg (0xffc100, 16 bit; byte 0xffc100 is bits 15..8) */
#define CFG_HI            (*(unsigned char *)&regmark_cfg)
#define CFG_ENABLE        0x80   /* bit15: build transform after search (TB99 / TB0) */
#define CFG_SCALE         0x40   /* bit14: TB55, scale to nominal mark distance */
#define CFG_SEARCH_MASK   0x30   /* bits13..12: 1 = 2 marks (TB22/122), 2 = 3 marks (TB23/123) */
#define CFG_SEARCH_2MARK  0x10
#define CFG_TYPE_MASK     0x0c   /* bits11..10: mark type 1 (TB52,1) or 2 (default) */
#define CFG_TYPE1         0x04
#define CFG_TYPE2         0x08

extern long regmark_marks[6];        /* 0xffb0ae  corners: mark1 x,y; mark2 x,y; mark3 x,y */
extern int  rm_yx_fx, rm_xx_fx;      /* 0xffb08a/c  2.14 fixed copies of rm_yx/rm_xx */
extern int  rm_yy_fx, rm_xy_fx;      /* 0xffb08e/90 */
extern int  rm_sx_fx, rm_sy_fx;      /* 0xffb0a2/4  2.14 fixed scale, 0x4000 = 1.0 */
extern char rm_flag_ac;              /* 0xffb0ac  cleared with the transform (TODO: never set here) */
extern long rm_unk_a6;               /* 0xffb0a6  cleared with the transform */
extern pt_t regmark_samples[6];      /* 0xffb3bc  bar-centre samples P0..P5, see find_type1/2 */
extern long span_x1, span_y1;        /* 0xffb3ec/f0  type-1 start offsets to mark 2 / mark 3 */
extern long span_x2, span_y2;        /* 0xffb3f4/f8  type-2 start offsets */
extern long regmark_feed_origin;     /* 0xffb3fc  margin between mark 1 and the new feed origin */
extern long regmark_offset_x;        /* 0xffc102  TB54 x */
extern long regmark_offset_y;        /* 0xffc106  TB54 y */
extern long regmark_dist_x;          /* 0xffc112  nominal mark1->mark2 distance (TB2x/12x x) */
extern long regmark_dist_y;          /* 0xffc116  nominal mark1->mark3 distance (TB2x/12x y) */
extern long sensor_filter_sum;       /* 0xffc12a */
extern int  sensor_ring[110];        /* 0xffc12e..0xffc209 */
extern int *sensor_ring_tail;        /* 0xffc20a */
extern int *sensor_ring_head;        /* 0xffc20e */
extern int *sensor_ring_mid;         /* 0xffc212 */
extern long scan_limit_xmax;         /* 0xffc21a  feed origin + 1100 mm */
extern long scan_start;              /* 0xffc21e  position where the current scan started */
extern pt_t dev_pos;                 /* 0xffb54c  dev_pos_x/dev_pos_y: current position, device logical coords */
extern long move_last_x, move_last_y;/* 0xffb264/8  last position handed to queue_segment */

static const double D_ONE = 1.0, D_MINUS_ONE = -1.0, D_ZERO = 0.0;   /* ROM 0x265ba/c2/ca */
static const double TAN5  = 0.087488663;     /* ROM 0x265d2: |slope| limit between marks (5 deg) */
static const double FX14  = 16384.0;         /* ROM 0x265e2: 2.14 fixed point */
static const double SCALE_MIN = 0.6667;      /* ROM 0x265f2 */
static const double SCALE_MAX = 2.0;         /* ROM 0x265fa */
static const double AREA_PAD  = 600.0;       /* ROM 0x26602: 3 mm added to the measured distance */

extern double sqrt(double);                                      /* 0x23f1e */

void regmark_clip_plot(long x, long y, int pen);
int  regmark_clip_segment(pt_t *p, pt_t *q, int pen, int code);
int  regmark_find_and_mark(void);
void sensor_filter_init(void);
int  sensor_filter_step(int axis, int sample);
void scan_record_edge(pt_t *edge, const pt_t *pos, int *v);
void scan_check_limits(int axis, int wide, int speed, pt_t *edge, int *v, char *fail);
void sensor_scan_axis(int axis, int filtered, int speed, char *fail, pt_t *edge, int wide);
void sensor_scan_diag(int axis_swap, int speed, char *fail, pt_t *edge);
void regmark_locate_x(signed char dir, char wide, pt_t *out);
void regmark_locate_y(signed char dir, char wide, pt_t *out);
static void sensor_pot_shift(int nbits, unsigned int bits);
static int  sensor_trace_scan(char axis);


/* =====================================================================
 * 1. Sensor hardware
 *
 * Reflective sensor: LED on P0x bit 4 (0xffff06), photo signal on A/D
 * channel 0.  The LED drive (gain) is a 10-bit serial digital pot
 * clocked on 0xffff00 (bit 1 data, bit 2 clock, bit 3 /CS).  Paper reads
 * high, a printed mark reads low; an edge is "value fell below a threshold".
 * ===================================================================== */

/* 0x2078a */
void sensor_on(void)
{
	LED_PORT |= 0x10;
	ADCR = 0x37;
	ADCSR = 0x30;          /* continuous conversion */
	delay_ticks(30);       /* let the LED settle */
}

/* 0x207bc */
void sensor_off(void)
{
	ADCSR = 0;
	LED_PORT &= ~0x10;
}

/* 0x219a0: shift the top `nbits` of `bits` out MSB first */
static void sensor_pot_shift(int nbits, unsigned int bits)
{
	int i;

	POT_PORT &= ~0x08;                              /* /CS low */
	for (i = 0; i < nbits; i++) {
		if (bits & 0x8000)
			POT_PORT |= 0x02;
		else
			POT_PORT &= ~0x02;
		bits <<= 1;
		POT_PORT |= 0x04;                       /* clock pulse */
		POT_PORT &= ~0x06;
	}
	POT_PORT |= 0x08;                               /* /CS high, latch */
}

/* 0x219fa: set the LED gain, level 0..255 (symbols.txt: sensor_pot_set_level) */
void sensor_pot_set_level(int level)
{
	sensor_pot_shift(10, (level << 6) & 0x3fc0);   /* 8 data bits + 2 trailing bits */
	delay_ticks(10);
}

/* 0x2195c */
void sensor_init(void)
{
	P1DDR_BIT &= ~0x02;
	POT_PORT |= 0x08;
	POT_PORT &= ~0x06;
	sensor_pot_set_level(eeprom_read_byte(0x300019));   /* gain found by auto-calibration */
	sensor_threshold = 0x200;                       /* 50 % of full scale */
	LED_PORT &= ~0x10;
}

/*
 * 0x21a26 (service g16): raise the LED gain from 0 until a single A/D
 * conversion exceeds sensor_threshold; store the level in EEPROM 0x19.
 * Run with the sensor over blank paper.
 */
int sensor_autocal_gain(int blink_forever)
{
	unsigned v;
	int level = 0;

	LED_PORT |= 0x10;
	ADCSR = 0;
	ADCR = 0x37;
	sensor_pot_set_level(level);
	for (;;) {
		ADCSR = 0x20;                           /* single conversion */
		while (!(ADCSR & 0x80))
			;
		v = ADC_VALUE();
		ADCSR = 0;
		if (v > (unsigned)sensor_threshold)
			break;
		sensor_pot_set_level(++level);
		if (level == 0xff)
			break;
	}
	eeprom_write_byte((unsigned char)level, 0x300019);
	if (!blink_forever)
		return level;
	for (;;) {                                      /* service: show "done" */
		delay_ticks(500);
		STAT_LED &= ~0x08;
		delay_ticks(500);
		STAT_LED |= 0x08;
	}
}

/* 0x21af8 (service n): threshold in percent of full scale */
void svc_n_threshold(void)
{
	sensor_threshold = (int)((unsigned long)read_int() * 1024 / 100);
}

/* 0x21c40: reply the current A/D value (%5d) */
void sensor_reply_adc(void)
{
	unsigned v;

	ADCR = 0x37;
	ADCSR = 0;
	v = ADC_VALUE();
	reply_begin(1);
	reply_number_end(5, v);
}

/* 0x21cb6 (service t0) */
void sensor_measure_reply(void)
{
	LED_PORT |= 0x10;
	delay_ticks(50);
	sensor_reply_adc();
	LED_PORT &= ~0x10;
}

/* 0x21b22 (service a) */
void svc_a(void)
{
	unsigned i;

	switch (read_int()) {
	case 0: LED_PORT |= 0x10; break;                /* LED on */
	case 1: sensor_pot_set_level((int)read_int()); break;/* set gain */
	case 2: sensor_reply_adc(); break;
	case 3: for (i = 1; i < 100; i++) sensor_reply_adc(); break;
	}
}

/*
 * Adaptive edge filter, 0x207d0 / 0x208d0.
 *
 * A ring of 110 samples is pre-filled with the baseline (paper) reading.
 * Each step writes the new sample at `head`, which runs N = min(1000 /
 * sensor_scan_speed, 105) entries ahead of `tail`; `sum` is the sum of the
 * 16 samples between `tail` and `mid`.  So sum/16 is the average of the
 * samples taken ~N steps ago, i.e. a few mm behind the sensor.  The step
 * returns gain/10 of that average: the live sample must drop below e.g.
 * 70 % (gain nibble 7) of what the paper looked like just before.
 */
void sensor_filter_init(void)
{
	int i, n;

	wait_motion_done();
	sensor_baseline = ADC_VALUE();
	for (i = 0; i < 110; i++)
		sensor_ring[i] = sensor_baseline;
	sensor_ring_tail = &sensor_ring[0];
	n = 1000 / sensor_scan_speed;
	if (n > 105)
		n = 105;
	sensor_ring_head = &sensor_ring[n];
	sensor_ring_mid = &sensor_ring[16];
	sensor_filter_sum = (long)sensor_baseline * 16;
}

int sensor_filter_step(int axis, int sample)
{
	int *end = &sensor_ring[110];
	int gain = axis == 0 ? sensor_gain_x : sensor_gain_y;

	*sensor_ring_head++ = sample;
	if (sensor_ring_head >= end)
		sensor_ring_head = sensor_ring;
	sensor_filter_sum -= *sensor_ring_tail++;
	if (sensor_ring_tail >= end)
		sensor_ring_tail = sensor_ring;
	sensor_filter_sum += *sensor_ring_mid++;
	if (sensor_ring_mid >= end)
		sensor_ring_mid = sensor_ring;
	return (int)((unsigned long)(gain * sensor_filter_sum) / 10 >> 4);
}


/* =====================================================================
 * 2. Scan primitives
 *
 * The scans drive the motors directly (motor_run_continuous) and ramp the
 * speed by one step per sample: v = (vx, vy), inc = (ix, iy).  Stopping
 * is done by flipping the increments (scan_record_edge) and letting the
 * loop ramp the speed back to zero.  While regmark_scan_mode is set, every
 * ordinary move is offset by scan_off_x/y so that positions refer to the
 * sensor instead of the knife (motion_line 0x1d4aa).
 * ===================================================================== */

/* 0x20a08: remember where the edge was and start decelerating.
 * v points at { vx, vy, ix, iy }. */
void scan_record_edge(pt_t *edge, const pt_t *pos, int *v)
{
	*edge = *pos;
	v[2] = -v[2];
	v[3] = -v[3];
	v[0] += v[2];
	v[1] += v[3];
}

/*
 * 0x20a90: travel limits during a filtered scan.  `wide` (set by the TB70
 * calibration) limits the scan to +-10000 (50 mm) around its start;
 * otherwise X may run back to the sensor offset and forward to 1100 mm
 * past the feed origin, Y between 4 mm and 200 mm of the sensor range.
 */
void scan_check_limits(int axis, int wide, int speed, pt_t *edge, int *v, char *fail)
{
	pt_t p;

	motion_get_pos(&p.x);
	if (axis == 0) {
		if (!wide) {
			if (speed <= 0 ? p.x < scan_off_x : p.x > scan_limit_xmax)
				goto stop;
		} else if (speed <= 0 ? p.x < scan_start - 10000 : p.x > scan_start + 10000) {
			goto stop;
		}
		return;
	}
	if (!wide) {
		if (speed <= 0 ? p.y + motor_y_bias < scan_off_y + 800
		               : p.y + motor_y_bias > scan_off_y + 40000)
			goto stop;
		return;
	}
	if (speed <= 0) {
		if (p.y < scan_start - 10000 || p.y + motor_y_bias < scan_off_y + 800)
			goto stop;
	} else {
		if (p.y > scan_start + 10000 || p.y > travel_max_y)
			goto stop;
	}
	return;
stop:
	if (!*fail) {
		scan_record_edge(edge, &p, v);
		*fail = 1;
	}
}

/*
 * 0x20d6a: run one axis until the sensor sees a mark edge.
 *
 *   axis      0 = X, 1 = Y (motor axes)
 *   filtered  1: adaptive threshold + travel limits (coarse "find" scan)
 *             0: fixed threshold baseline*gain/10, at most 4000 units (20 mm)
 *   speed     signed target speed (motor_run_continuous units)
 *   *fail     set when no edge was found (limit, window, ESC-cancel)
 *   *edge     position (device logical coords) where the edge was seen
 *   wide      passed to scan_check_limits
 */
void sensor_scan_axis(int axis, int filtered, int speed, char *fail, pt_t *edge, int wide)
{
	int v[4] = { 0, 0, 0, 0 };              /* vx, vy, ix, iy */
	int dir = speed > 0 ? 1 : -1;
	int vmax = speed > 0 ? speed : -speed;
	unsigned thr, adc;
	char found = 0;
	pt_t p, start, saved;
	long window_end;

	sensor_on();
	sensor_filter_init();
	*fail = 0;
	motion_get_pos(&start.x);
	if (axis == 0) {
		thr = (unsigned)(sensor_baseline * sensor_gain_x) / 10;
		v[2] = dir;
		window_end = (start.x + dir * 4000L) * dir;
		scan_limit_xmax = regmark_feed_origin + 220000;
		scan_start = start.x;
	} else {
		thr = (unsigned)(sensor_baseline * sensor_gain_y) / 10;
		v[3] = dir;
		window_end = (start.y + dir * 4000L) * dir;
		scan_start = start.y;
	}

	do {
		if ((v[0] < 0 ? -v[0] : v[0]) < vmax)
			v[0] += v[2];
		if ((v[1] < 0 ? -v[1] : v[1]) < vmax)
			v[1] += v[3];
		adc = ADC_VALUE();
		if (filtered) {
			thr = sensor_filter_step(axis, adc);
			if (!*fail && !found)
				scan_check_limits(axis, wide, speed, edge, v, fail);
		} else {
			motion_get_pos(&p.x);
			if ((axis == 0 ? p.x : p.y) * dir > window_end && !*fail && !found) {
				scan_record_edge(edge, &p, v);
				*fail = 1;
			}
		}
		if (cancel_request && !found && !*fail) {
			motion_get_pos(&p.x);
			scan_record_edge(edge, &p, v);
			*fail = 1;
		}
		if (adc < thr && !found && !*fail) {
			motion_get_pos(&p.x);
			scan_record_edge(edge, &p, v);
			found = 1;
		}
		motor_run_continuous(1, v[0], v[1]);
	} while (v[0] != 0 || v[1] != 0);
	motor_run_continuous(1, 0, 0);
	sensor_off();
	wait_motion_done();

	/* convert the motor position of the edge into device logical coords */
	motion_get_pos_wait(&saved.x);
	motor_target_x = edge->x;
	motor_target_y = edge->y;
	sync_device_pos();
	*edge = dev_pos;
	motor_target_x = saved.x;
	motor_target_y = saved.y;
	sync_device_pos();
}

/*
 * 0x21620: diagonal scan for the automatic first-mark search.  Both
 * motors run at the same speed.  Portrait: speed > 0 runs (-X, +Y),
 * speed <= 0 runs (+X, -Y).  Landscape: (+X, +Y) / (-X, -Y).  Fails when
 * the sensor leaves the travel range; always uses the adaptive threshold.
 */
void sensor_scan_diag(int axis_swap, int speed, char *fail, pt_t *edge)
{
	int v[4] = { 0, 0, 0, 0 };
	int vmax = speed > 0 ? speed : -speed;
	unsigned adc, thr;
	char found = 0;
	pt_t p, saved;

	sensor_on();
	sensor_filter_init();
	*fail = 0;
	if (axis_swap == 0) {
		v[2] = speed > 0 ? -1 : 1;
		v[3] = speed > 0 ? 1 : -1;
	} else {
		v[2] = v[3] = speed > 0 ? 1 : -1;
	}
	do {
		if ((v[0] < 0 ? -v[0] : v[0]) < vmax)
			v[0] += v[2];
		if ((v[1] < 0 ? -v[1] : v[1]) < vmax)
			v[1] += v[3];
		adc = ADC_VALUE();
		thr = sensor_filter_step(0, adc);
		if (!*fail && !found) {
			motion_get_pos(&p.x);
			p.x -= scan_off_x;
			p.y -= scan_off_y;
			if (p.x < -9 || p.y < -9 || p.x > travel_max_x + 10 || p.y > travel_max_y + 10) {
				scan_record_edge(edge, &p, v);
				*fail = 1;
			}
		}
		if (cancel_request && !found && !*fail) {
			motion_get_pos(&p.x);
			scan_record_edge(edge, &p, v);
			*fail = 1;
		}
		if (adc < thr && !found && !*fail) {
			motion_get_pos(&p.x);
			scan_record_edge(edge, &p, v);
			found = 1;
		}
		motor_run_continuous(1, v[0], v[1]);
	} while (v[0] != 0 || v[1] != 0);
	motor_run_continuous(1, 0, 0);
	sensor_off();
	wait_motion_done();
	motion_get_pos_wait(&saved.x);
	motor_target_x = edge->x;
	motor_target_y = edge->y;
	sync_device_pos();
	*edge = dev_pos;
	motor_target_x = saved.x;
	motor_target_y = saved.y;
	sync_device_pos();
}

/*
 * 0x2114c: find the centre line of a mark bar that crosses the logical
 * X axis ahead of the sensor (dir = +1 or -1).
 *
 *   1. fast filtered scan in `dir` until the bar is hit (edge E1);
 *   2. back off to E1 - dir*400 (2 mm), slow precise scan forward: E2;
 *   3. go to E1 + dir*600 (3 mm, past the bar), slow scan back: E3;
 *   4. centre = (E2 + E3) / 2, written to *out.  regmark_result = fail.
 *
 * In landscape the logical X axis is motor Y and runs negative.
 * `wide` is 1 only for the TB70 calibration (+-50 mm window).
 */
void regmark_locate_x(signed char dir, char wide, pt_t *out)
{
	int axis = landscape ? 1 : 0;
	int sign = landscape ? -1 : 1;
	char fail;
	pt_t start, e1, e2 = { 0, 0 }, e3;   /* original leaves e2 uninitialised on early failure */

	cancel_request = 0;
	wait_motion_done();
	sync_device_pos();
	start = dev_pos;

	sensor_scan_axis(axis, 1, dir * sensor_scan_speed * sign, &fail, &e1, wide);
	regmark_result = fail;
	if (fail) {
		*out = e2;   /* callers ignore *out when regmark_result != 0 */
		return;
	}
	move_clamped_wait(e1.x - dir * 400L, start.y, 0);
	wait_motion_done();
	sensor_scan_axis(axis, 0, dir * sign, &fail, &e2, wide);
	regmark_result = fail;
	if (fail) {
		*out = e2;
		return;
	}
	move_clamped_wait(e1.x + dir * 600L, start.y, 0);
	wait_motion_done();
	sensor_scan_axis(axis, 0, -(dir * sign), &fail, &e3, wide);
	regmark_result = fail;
	if (fail) {
		*out = e2;
		return;
	}
	out->x = (e2.x + e3.x) / 2;
	out->y = (e2.y + e3.y) / 2;
}

/* 0x2135c: same for a bar crossing the logical Y axis. */
void regmark_locate_y(signed char dir, char wide, pt_t *out)
{
	int axis = landscape ? 0 : 1;
	char fail;
	pt_t start, e1, e2 = { 0, 0 }, e3;   /* original leaves e2 uninitialised on early failure */

	cancel_request = 0;
	wait_motion_done();
	sync_device_pos();
	start = dev_pos;

	sensor_scan_axis(axis, 1, dir * sensor_scan_speed, &fail, &e1, wide);
	regmark_result = fail;
	if (fail) {
		*out = e2;
		return;
	}
	move_clamped_wait(start.x, e1.y - dir * 400L, 0);
	wait_motion_done();
	sensor_scan_axis(axis, 0, dir, &fail, &e2, wide);
	regmark_result = fail;
	if (fail) {
		*out = e2;
		return;
	}
	move_clamped_wait(start.x, e1.y + dir * 600L, 0);
	wait_motion_done();
	sensor_scan_axis(axis, 0, -dir, &fail, &e3, wide);
	regmark_result = fail;
	if (fail) {
		*out = e2;
		return;
	}
	out->x = (e2.x + e3.x) / 2;
	out->y = (e2.y + e3.y) / 2;
}

/*
 * 0x21546: move (pen up/down) to a logical position, clamping Y to the
 * machine range, and wait.  Returns 1 if Y had to be clamped -- the
 * searches treat that as "mark outside the reachable area".
 */
int move_clamped_wait(long x, long y, int pen)
{
	long p[2];
	int clamped = 0;

	apply_orientation(x, y, p);
	if (p[1] < mach_ymin) {
		p[1] = mach_ymin;
		clamped = 1;
	}
	if (p[1] > mach_ymax) {
		p[1] = mach_ymax;
		clamped = 1;
	}
	distcorr_to_user(p);
	move_last_x = p[0];
	move_last_y = p[1];
	queue_segment(p[0], p[1], pen);
	pen_return_to_saved();
	motion_drain();
	wait_motion_done();
	return clamped;
}


/* =====================================================================
 * 3. Mark search
 *
 * A mark is an "L" of two bars, each regmark_length mm long (TB51, 5..20,
 * default 10).  Every bar is sampled once with locate_x / locate_y; the
 * sample is a point on the bar's centre line.
 *
 *   P0 Y-bar of mark 1      P1 X-bar of mark 1
 *   P2 Y-bar of mark 2      P3 X-bar of mark 2      (mark 2 along +X)
 *   P4 X-bar of mark 3      P5 Y-bar of mark 3      (mark 3 along +Y, 3-mark only)
 *
 * ("Y-bar" = bar parallel to Y, found by scanning along X.)
 * ===================================================================== */

/* 0x15a1a: where to start looking for marks 2/3 relative to mark 1.
 * If a nominal distance d was given (TB23,d,...) start 10 mm short of it;
 * otherwise scan from right next to mark 1 until the next bar. */
void regmark_compute_spans(void)
{
	long len = regmark_length * 200L;

	if (regmark_dist_x - 2000 < len) {
		span_x1 = 0;
		span_x2 = len / 2;
	} else {
		span_x1 = span_x2 = regmark_dist_x - 2000;
	}
	if (regmark_dist_y - 2000 < len) {
		span_y1 = 0;
		span_y2 = len / 2;
	} else {
		span_y1 = span_y2 = regmark_dist_y - 2000;
	}
}

/* 0x15b16: distance between mark 1 and the new feed origin */
long regmark_origin_margin(void)
{
	long m = 0;

	if ((CFG_HI & CFG_TYPE_MASK) == CFG_TYPE2)
		m = regmark_length * 200L / 2;
	if (!landscape) {
		if (regmark_offset_x < 0)
			m -= regmark_offset_x;
	} else if (regmark_offset_y < 0) {
		m -= regmark_offset_y;
	}
	return m + 4000;                                /* + 20 mm */
}

/*
 * 0x15212: type-1 marks (TB52,1).  Every scan runs in the + direction:
 * the arms of each L point away from the page (outward corners).
 * Returns 0x10 when the 2-mark search ends early; regmark_result is the
 * real status.
 */
char regmark_find_type1(void)
{
	long half = regmark_length * 200L / 2;
	pt_t *P = regmark_samples;
	pt_t a, b;

	regmark_locate_x(1, 0, &P[0]);                 /* Y-bar of mark 1 */
	if (regmark_result)
		return regmark_result;
	b = a = P[0];
	a.x += span_x1;
	if (move_clamped_wait(a.x, a.y, 0))
		return regmark_result = 1;
	regmark_locate_x(1, 0, &P[2]);                 /* Y-bar of mark 2 */
	if (regmark_result)
		return regmark_result;
	a = P[2];
	a.x += half;
	if (move_clamped_wait(a.x, a.y, 0))
		return regmark_result = 1;
	regmark_locate_y(1, 0, &P[3]);                 /* X-bar of mark 2, +X of its Y-bar */
	if (regmark_result)
		return regmark_result;
	b.x -= half;
	if (move_clamped_wait(b.x, b.y, 0))
		return regmark_result = 1;
	regmark_locate_y(1, 0, &P[1]);                 /* X-bar of mark 1, -X of its Y-bar */
	if (regmark_result)
		return regmark_result;
	if ((CFG_HI & CFG_SEARCH_MASK) == CFG_SEARCH_2MARK)
		return 0x10;
	a = P[1];
	a.y += span_y1;
	if (move_clamped_wait(a.x, a.y, 0))
		return regmark_result = 1;
	regmark_locate_y(1, 0, &P[4]);                 /* X-bar of mark 3 */
	if (regmark_result)
		return regmark_result;
	a = P[4];
	a.y += half;
	if (move_clamped_wait(a.x, a.y, 0))
		return regmark_result = 1;
	regmark_locate_x(1, 0, &P[5]);                 /* Y-bar of mark 3 */
	return regmark_result;
}

/*
 * 0x154ca: type-2 marks (default).  The arms of each L point into the page
 * (inward corners): mark 1 is found by scanning back (-X, -Y), mark 2 by
 * scanning +X, mark 3 by scanning +Y.
 */
char regmark_find_type2(void)
{
	long half = regmark_length * 200L / 2;
	pt_t *P = regmark_samples;
	pt_t a, b;
	long p0x, p1y;

	regmark_locate_x(-1, 0, &P[0]);                /* Y-bar of mark 1, behind the sensor */
	a = P[0];
	p0x = P[0].x;
	if (regmark_result)
		return regmark_result;
	a.x += half;
	if (move_clamped_wait(a.x, a.y, 0))
		return regmark_result = 1;
	regmark_locate_y(-1, 0, &P[1]);                /* X-bar of mark 1 */
	p1y = P[1].y;
	if (regmark_result)
		return regmark_result;
	if (move_clamped_wait(p0x + span_x2, a.y, 0))
		return regmark_result = 1;
	regmark_locate_x(1, 0, &P[2]);                 /* Y-bar of mark 2 */
	b = P[2];
	if (regmark_result)
		return regmark_result;
	b.x -= half;
	if (move_clamped_wait(b.x, b.y, 0))
		return regmark_result = 1;
	regmark_locate_y(-1, 0, &P[3]);                /* X-bar of mark 2 */
	if (regmark_result)
		return regmark_result;
	if ((CFG_HI & CFG_SEARCH_MASK) == CFG_SEARCH_2MARK)
		return 0x10;
	if (move_clamped_wait(a.x, p1y + span_y2, 0))
		return regmark_result = 1;
	regmark_locate_y(1, 0, &P[4]);                 /* X-bar of mark 3 */
	if (regmark_result)
		return regmark_result;
	b = P[4];
	b.y -= half;
	if (move_clamped_wait(b.x, b.y, 0))
		return regmark_result = 1;
	regmark_locate_x(-1, 0, &P[5]);                /* Y-bar of mark 3 */
	return regmark_result;
}

/*
 * 0x15bbc: automatic search for mark 1 (TB12x).  From the current
 * position it sweeps diagonals between the X and Y axes, alternating
 * (k*s, 0) -> up-left and (0, k*s) -> down-right, with s = 1.6*L mm, until
 * a diagonal crosses a bar, giving up after 100 mm.  It then crosses the
 * other arm along X or Y and parks the sensor half an arm inside the bar,
 * which is where find_type1/2 expect to start.
 */
long regmark_scan_first(void)
{
	long step = regmark_length * 200L * 8 / 10;     /* 0.8 * L mm */
	long dx = -step, dy = 0;
	int speed = sensor_scan_speed * 10 / 14;
	char fail;
	pt_t hit, e;

	if (speed < 1)
		speed = 1;
	regmark_result = 0;
	for (;;) {
		dx += step * 2;
		if (dx > 20000)
			return regmark_result = 1;
		move_clamped_wait(dx, 0, 0);
		sensor_scan_diag(landscape, speed, &fail, &hit);
		if (!fail) {
			/* hit the Y-bar: cross it along +X from 1.5 L before it */
			hit.x -= regmark_length * 600L / 2;
			if (hit.x < 0)
				hit.x = 0;
			hit.y += regmark_length * 200L / 2;
			move_clamped_wait(hit.x, hit.y, 0);
			sensor_scan_axis(landscape ? 1 : 0, 1, landscape ? -sensor_scan_speed : sensor_scan_speed,
			                 &fail, &e, 1);
			regmark_result = fail;
			if (fail)
				return 1;
			return move_clamped_wait(e.x + regmark_length * 200L / 2, e.y, 0);
		}
		dy += step * 2;
		if (dy > 20000)
			return regmark_result = 1;
		move_clamped_wait(0, dy, 0);
		sensor_scan_diag(landscape, -speed, &fail, &hit);
		if (!fail)
			break;
	}
	/* hit the X-bar: cross it along +Y from 1.5 L before it */
	hit.y -= regmark_length * 600L / 2;
	if (hit.y < 0)
		hit.y = 0;
	hit.x += regmark_length * 200L / 2;
	move_clamped_wait(hit.x, hit.y, 0);
	sensor_scan_axis(landscape ? 0 : 1, 1, sensor_scan_speed, &fail, &e, 1);
	regmark_result = fail;
	if (fail)
		return 1;
	return move_clamped_wait(e.x, e.y + regmark_length * 200L / 2, 0);
}

/* common tail of both searches: undo the page set-up after a failure */
static void regmark_search_failed(char saved_media_mode)
{
	pt_t pos;

	wait_motion_done();
	set_media_mode(saved_media_mode);
	motion_get_pos_wait(&pos.x);
	if (!cancel_request)
		set_carriage_origin(pos.x + feed_total, pos.y);
	else
		set_carriage_origin(0, pos.y);
	feed_total = 0;
	regmark_clear_transform();
	page_length = 200000;                           /* 1000 mm */
	page_width = 40000;                             /* 200 mm */
	mach_xmax = page_length;                        /* __movblk(&page_length, &mach_xmax, 8) */
	mach_ymax = page_width;
}

/* common set-up of both searches */
static char regmark_search_begin(void)
{
	char saved = media_mode;

	motion_flush();
	pen_return_to_saved();
	motion_drain();
	wait_motion_done();
	page_limits_apply(1);
	set_media_mode(1);
	orient_off_y = 0;
	orient_off_x = 0;
	regmark_clear_transform();
	regmark_compute_spans();
	regmark_clear_transform();
	return saved;
}

/* common end: compute transform if enabled, report "%5d" + ETX */
static void regmark_search_end(char saved)
{
	if (regmark_result == 0) {
		if (CFG_HI & CFG_ENABLE) {
			regmark_compute_transform();
			if (regmark_result)
				regmark_search_failed(saved);
			else
				wait_motion_done();
		} else {
			wait_motion_done();
			wait_motion_done();
		}
	} else {
		regmark_search_failed(saved);
	}
	regmark_goto_origin();
	reply_begin(1);
	reply_number_end(5, regmark_result);           /* "    0" = found */
}

/*
 * 0x1575a, TB22 / TB23: the user has put the sensor on mark 1.  The feed
 * origin is moved to `margin` behind the sensor, the page is set to
 * 1100 mm minus that, and the marks are searched.
 */
void regmark_search_manual(void)
{
	char saved = regmark_search_begin();
	pt_t pos;
	long n;

	motion_get_pos_wait(&pos.x);
	regmark_feed_origin = regmark_origin_margin();
	n = pos.x - regmark_feed_origin;
	feed_total += n;
	page_length = 220000 - n;
	page_width = 40000;
	mach_xmax = page_length;
	mach_ymax = page_width;
	set_carriage_origin(regmark_feed_origin, pos.y);
	sync_device_pos();
	cancel_request = 0;
	regmark_scan_mode = 1;                          /* moves now position the sensor */
	move_clamped_wait(dev_pos.x, dev_pos.y, 0);
	if ((CFG_HI & CFG_TYPE_MASK) == CFG_TYPE1)
		regmark_find_type1();
	else
		regmark_find_type2();
	regmark_scan_mode = 0;
	regmark_search_end(saved);
}

/* 0x15fb2, TB122 / TB123: find mark 1 automatically, then as above. */
void regmark_search_auto(void)
{
	char saved = regmark_search_begin();

	cancel_request = 0;
	regmark_scan_mode = 1;
	regmark_scan_first();
	if (regmark_result == 0) {
		if ((CFG_HI & CFG_TYPE_MASK) == CFG_TYPE1)
			regmark_find_type1();
		else
			regmark_find_type2();
	}
	regmark_scan_mode = 0;
	regmark_search_end(saved);
}

/* 0x1596c: put the knife at user (0,0) of the registered page */
void regmark_goto_origin(void)
{
	pt_t p = { 0, 0 };

	if (regmark_active)
		regmark_transform(&p, 0, 0);
	p.x += regmark_off_x;
	p.y += regmark_off_y;
	device_plot_nocal(p.x, p.y, 0);
	pen_return_to_saved();
	motion_drain();
	wait_motion_done();
	sync_device_pos();
}


/* =====================================================================
 * 4. Geometry
 * ===================================================================== */

/* 0x05452 */
void regmark_get_mark(int idx, long *x, long *y)
{
	switch (idx) {
	case 0:
	case 1: *x = regmark_marks[0]; *y = regmark_marks[1]; break;
	case 2: *x = regmark_marks[2]; *y = regmark_marks[3]; break;
	case 3: *x = regmark_marks[4]; *y = regmark_marks[5]; break;
	}
}

/*
 * 0x5082: foot of the perpendicular from point a onto the line b-c,
 * i.e. the corner of an L whose second arm passes through `a` at a right
 * angle.  *parallel = 1 (and (0,0)) if a lies on the line in the
 * degenerate axis-parallel cases.
 */
void regmark_axis_intersect(int *parallel, pt_t *out, pt_t c, pt_t b, pt_t a)
{
	double m, b1, m2, b2, x;

	*parallel = 0;
	if (c.x == b.x) {                              /* line b-c vertical */
		if (a.x == c.x) {
			out->x = out->y = 0;
			*parallel = 1;
		} else {
			out->x = c.x;
			out->y = a.y;
		}
		return;
	}
	if (c.y == b.y) {                              /* line b-c horizontal */
		if (a.y == c.y) {
			out->x = out->y = 0;
			*parallel = 1;
		} else {
			out->x = a.x;
			out->y = c.y;
		}
		return;
	}
	m = ((double)c.y - b.y) / ((double)c.x - b.x);
	b1 = ((double)c.y * b.x - (double)b.y * c.x) / ((double)b.x - c.x);
	m2 = D_MINUS_ONE / m;
	b2 = a.y - m2 * a.x;
	if (m == m2) {
		out->x = out->y = 0;
		*parallel = 1;
		return;
	}
	x = (b2 - b1) / (m - m2);
	out->x = (long)x;
	out->y = (long)(m * x + b1);
}

/* 0x48ba: intersection of line a-b with line c-d (3-mark origin corner) */
void regmark_line_intersect(int *parallel, pt_t *out, pt_t d, pt_t c, pt_t b, pt_t a)
{
	double m1, b1, m2, b2, x;

	*parallel = 0;
	if (d.x == c.x && b.x == a.x) {
		out->x = out->y = 0;
		*parallel = 1;
		return;
	}
	if (b.x != a.x) {
		m2 = ((double)a.y - b.y) / ((double)a.x - b.x);
		b2 = ((double)a.y * b.x - (double)b.y * a.x) / ((double)b.x - a.x);
	}
	if (d.x != c.x) {
		m1 = ((double)c.y - d.y) / ((double)c.x - d.x);
		b1 = ((double)c.y * d.x - (double)d.y * c.x) / ((double)d.x - c.x);
	}
	if (d.x == c.x) {                              /* c-d vertical */
		out->x = d.x;
		out->y = (long)(m2 * d.x + b2);
		return;
	}
	if (b.x == a.x) {                              /* a-b vertical */
		out->x = b.x;
		out->y = (long)(m1 * b.x + b1);
		return;
	}
	if (m1 == m2) {
		out->x = out->y = 0;
		*parallel = 1;
		return;
	}
	x = (b2 - b1) / (m1 - m2);
	out->x = (long)x;
	out->y = (long)(m1 * x + b1);
}

/* 0x46b8: mark corners from the bar samples.  The parallel flag is ignored. */
void regmark_find_corners(void)
{
	pt_t *P = regmark_samples;
	pt_t *M = (pt_t *)regmark_marks;
	int par;

	if ((CFG_HI & CFG_SEARCH_MASK) == CFG_SEARCH_2MARK) {
		/* X-axis line through both X-bars; drop the Y-bar samples onto it */
		regmark_axis_intersect(&par, &M[0], P[1], P[3], P[0]);
		regmark_axis_intersect(&par, &M[1], P[1], P[3], P[2]);
	} else {
		/* origin = X-line (P1,P3) x Y-line (P0,P5); marks 2, 3 projected */
		regmark_line_intersect(&par, &M[0], P[1], P[3], P[0], P[5]);
		regmark_axis_intersect(&par, &M[1], P[1], P[3], P[2]);
		regmark_axis_intersect(&par, &M[2], P[0], P[5], P[4]);
	}
}

/* 0x45e4 (the firmware leaves rm_yy = 0 and rm_xy = 1.0 here; harmless,
 * the transform is only used while regmark_active is set) */
void regmark_clear_transform(void)
{
	regmark_active = 0;
	rm_scale_on = 0;
	rm_flag_ac = 0;
	rm_yx = D_ZERO;
	rm_xx = D_ONE;
	rm_yy = D_ZERO;
	rm_xy = D_ONE;
	rm_yy_fx = 0;        /* fixed copies follow the doubles: yx = yy = 0, xx = xy = 1.0 */
	rm_yx_fx = 0;
	rm_xy_fx = 0x4000;
	rm_xx_fx = 0x4000;
	rm_sx = D_ONE;
	rm_sy = D_ONE;
	rm_sx_fx = 0x4000;
	rm_sy_fx = 0x4000;
	regmark_off_x = 0;
	regmark_off_y = 0;
	rm_unk_a6 = 0;
}

/*
 * 0x5bc0: mark1->mark2 distance.  With TB55 enabled and a nominal
 * distance given (TB2x/12x x), store the scale nominal/measured
 * (accepted 0.6667..2.0, else 1.0) and return the nominal distance;
 * otherwise return measured + 3 mm.  Used as the clip-area size.
 */
long regmark_scale_x(double dx, double dy)
{
	double meas = sqrt(dx * dx + dy * dy);

	if (regmark_dist_x == 0 || !(CFG_HI & CFG_SCALE)) {
		rm_sx = D_ONE;
		rm_sx_fx = 0x4000;
	} else {
		rm_sx = regmark_dist_x / meas;
		if (rm_sx < SCALE_MIN || rm_sx > SCALE_MAX) {
			regmark_scale_fail_hook();      /* empty */
			rm_sx = D_ONE;
			rm_sx_fx = 0x4000;
		} else {
			rm_sx_fx = (int)(rm_sx * FX14);
		}
	}
	if (rm_sx_fx != 0x4000) {
		rm_scale_on = 1;
		return regmark_dist_x;
	}
	return (long)(meas + AREA_PAD);
}

/* 0x5d40: same for mark1->mark3 */
long regmark_scale_y(double dx, double dy)
{
	double meas = sqrt(dx * dx + dy * dy);

	if (regmark_dist_y == 0 || !(CFG_HI & CFG_SCALE)) {
		rm_sy = D_ONE;
		rm_sy_fx = 0x4000;
	} else {
		rm_sy = regmark_dist_y / meas;
		if (rm_sy < SCALE_MIN || rm_sy > SCALE_MAX) {
			regmark_scale_fail_hook();
			rm_sy = D_ONE;
			rm_sy_fx = 0x4000;
		} else {
			rm_sy_fx = (int)(rm_sy * FX14);
		}
	}
	if (rm_sy_fx != 0x4000) {
		rm_scale_on = 1;
		return regmark_dist_y;
	}
	return (long)(meas + AREA_PAD);
}

static void regmark_fail_park(void)
{
	regmark_result = 1;
	device_plot_nocal(0, 0, 0);
	pen_return_to_saved();
	motion_drain();
}

/*
 * 0x54e4: build the registration transform from the mark corners.
 *
 *   X axis = unit vector mark1->mark2 (must point +X within +-5 deg)
 *   Y axis = 2 marks: X axis turned by +90 deg
 *            3 marks: unit vector mark1->mark3 (must point +Y within +-5 deg),
 *                     so a skewed (non-square) print is followed too
 *
 *   x' = rm_xx*x + rm_xy*y,   y' = rm_yy*y + rm_yx*x
 *
 * The page origin (regmark_off) is mark 1 plus the TB54 offset rotated
 * into the mark frame.  Plotting is clipped to the rectangle spanned by
 * the marks (rm_area_*): 0..dist+3 mm - offset, Y unbounded (1000 mm) with
 * 2 marks.  On success the knife goes to the new origin.
 */
void regmark_compute_transform(void)
{
	long m1x, m1y, m2x, m2y, m3x, m3y, ox, oy;
	double dx, dy, ex, ey, len, slope;
	long save_ox, save_oy;
	pt_t off;

	regmark_find_corners();
	regmark_clear_transform();
	rm_area_xmin = 0;
	rm_area_ymin = 0;
	rm_last_outcode = 0;
	regmark_get_mark(1, &m1x, &m1y);
	regmark_get_mark(2, &m2x, &m2y);
	dx = (double)(m2x - m1x);
	dy = (double)(m2y - m1y);
	if (dx == D_ZERO && dy == D_ZERO) {
		regmark_result = 1;
		regmark_clear_transform();
		device_plot_nocal(0, 0, 0);
		pen_return_to_saved();
		motion_drain();
		return;
	}
	if (dx == D_ZERO)
		dx = D_MINUS_ONE;                       /* forces the failure below */
	else
		slope = dy / dx;
	if (dx <= D_ZERO || slope > TAN5 || slope < -TAN5) {
		regmark_fail_park();
		return;
	}
	len = sqrt(dx * dx + dy * dy);

	if ((CFG_HI & CFG_SEARCH_MASK) == CFG_SEARCH_2MARK) {
		rm_yx = dy / len;                       /* sin */
		rm_xx = dx / len;                       /* cos */
		rm_yx_fx = (int)(rm_yx * FX14);
		rm_xx_fx = (int)(rm_xx * FX14);
		rm_yy = rm_xx;
		rm_xy = -rm_yx;
		rm_yy_fx = rm_xx_fx;
		rm_xy_fx = -rm_yx_fx;
		rm_area_xmax = regmark_scale_x(dx, dy) - regmark_offset_x;
		rm_area_ymax = 200000;
	} else {
		regmark_get_mark(3, &m3x, &m3y);
		rm_yx = dy / len;
		rm_xx = dx / len;
		rm_yx_fx = (int)(rm_yx * FX14);
		rm_xx_fx = (int)(rm_xx * FX14);
		ex = (double)(m3x - m1x);
		ey = (double)(m3y - m1y);
		if (ey == D_ZERO)
			ey = D_MINUS_ONE;
		else
			slope = ex / ey;
		if (ey <= D_ZERO || slope > TAN5 || slope < -TAN5) {
			regmark_fail_park();
			return;
		}
		len = sqrt(ex * ex + ey * ey);
		if (len == D_ZERO) {
			regmark_fail_park();
			return;
		}
		rm_yy = ey / len;
		rm_xy = ex / len;
		rm_yy_fx = (int)(rm_yy * FX14);
		rm_xy_fx = (int)(rm_xy * FX14);
		rm_area_xmax = regmark_scale_x(dx, dy) - regmark_offset_x;
		rm_area_ymax = regmark_scale_y(ex, ey) - regmark_offset_y;
	}

	regmark_get_mark(0, &ox, &oy);
	regmark_transform(&off, regmark_offset_x, regmark_offset_y);
	ox += off.x;
	oy += off.y;
	regmark_active = 1;
	save_ox = orient_off_x;
	save_oy = orient_off_y;
	origin_reset(1);
	orient_off_x = save_ox;
	orient_off_y = save_oy;
	regmark_off_x = ox;
	regmark_off_y = oy;
	device_plot_raw(0, 0, 0);
	pen_return_to_saved();
	motion_drain();
	sync_device_pos();
}

/* 0x5ecc: lengthen p1-p2 by 3000 units (15 mm) at both ends */
void regmark_extend_segment(long *p1, long *p2)
{
	double dx = (double)(p1[0] - p2[0]);
	double dy = (double)(p1[1] - p2[1]);
	double len = sqrt(dx * dx + dy * dy);
	long a0, a1, b0, b1;

	a0 = (long)(p1[0] + (double)((p1[0] - p2[0]) * 3000) / len);
	a1 = (long)(p1[1] + (double)((p1[1] - p2[1]) * 3000) / len);
	b0 = (long)(p2[0] + (double)((p2[0] - p1[0]) * 3000) / len);
	b1 = (long)(p2[1] + (double)((p2[1] - p1[1]) * 3000) / len);
	p1[0] = a0;
	p1[1] = a1;
	p2[0] = b0;
	p2[1] = b1;
}


/* =====================================================================
 * 5. Use in the device pipeline
 *
 * plot_device (0x1c572) adds the origin offsets, then, while
 * regmark_active, clips every segment against rm_area_* (regmark_clip_plot)
 * before device_plot; apply_regmark_and_cal (0x1c918) applies
 * regmark_transform and adds regmark_off.  With TB55 scaling the point is
 * first multiplied by rm_sx/rm_sy.
 * ===================================================================== */

/* 0x1c6c4 */
void regmark_transform(pt_t *out, long x, long y)
{
	if (rm_scale_on) {
		x = (long)(x * rm_sx);
		y = (long)(y * rm_sy);
	}
	out->x = (long)(x * rm_xx + y * rm_xy);
	out->y = (long)(y * rm_yy + x * rm_yx);
}

/* 0x1c242 */
int regmark_area_outcode(long *p)
{
	int c = 0;

	if (p[0] > rm_area_xmax)
		c = 1;
	else if (p[0] < rm_area_xmin)
		c = 2;
	if (p[1] > rm_area_ymax)
		c |= 4;
	else if (p[1] < rm_area_ymin)
		c |= 8;
	return c;
}

/* 0x1c39a: move p onto x along p-q, return its new outcode */
int regmark_clip_x(pt_t *p, pt_t *q, long x)
{
	if (q->x != p->x)
		p->y += muldiv(q->y - p->y, x - p->x, q->x - p->x);
	p->x = x;
	return regmark_area_outcode(&p->x);
}

/* 0x1c414 */
int regmark_clip_y(pt_t *p, pt_t *q, long y)
{
	if (q->y != p->y)
		p->x += muldiv(q->x - p->x, y - p->y, q->y - p->y);
	p->y = y;
	return regmark_area_outcode(&p->x);
}

/* 0x1c2d6: clip p towards q; if that brings p inside, plot to it */
int regmark_clip_segment(pt_t *p, pt_t *q, int pen, int code)
{
	if (code & 1)
		code = regmark_clip_x(p, q, rm_area_xmax);
	else if (code & 2)
		code = regmark_clip_x(p, q, rm_area_xmin);
	if (code & 4)
		code = regmark_clip_y(p, q, rm_area_ymax);
	else if (code & 8)
		code = regmark_clip_y(p, q, rm_area_ymin);
	if (code == 0)
		device_plot_noclip(p->x, p->y, pen);
	return code;
}

/* 0x1c146: Cohen-Sutherland against the registered area */
void regmark_clip_plot(long x, long y, int pen)
{
	pt_t from = { rm_last_x, rm_last_y }, to = { x, y };
	int c0 = rm_last_outcode, c1;

	rm_last_x = x;
	rm_last_y = y;
	rm_last_outcode = c1 = regmark_area_outcode(&to.x);
	if (c0 & c1)
		return;                                 /* both outside on the same side */
	if (c0 != 0)
		c0 = regmark_clip_segment(&from, &to, 0, c0);   /* pen-up to the entry point */
	if (c0 != 0)
		return;
	if (c1 != 0)
		c1 = regmark_clip_segment(&to, &from, pen, c1);
	if (c1 == 0)
		device_plot_noclip(to.x, to.y, pen);    /* (repeats the clipped end point) */
}


/* =====================================================================
 * 6. Commands
 * ===================================================================== */

/* 0x20616: power-on defaults */
void regmark_init(void)
{
	unsigned char b;

	CFG_HI &= ~CFG_SEARCH_MASK;
	CFG_HI &= ~CFG_SCALE;
	CFG_HI = (CFG_HI & ~CFG_TYPE_MASK) | CFG_TYPE2;
	CFG_HI &= ~CFG_ENABLE;
	regmark_length = 10;
	regmark_offset_x = regmark_offset_y = 0;
	regmark_dist_x = regmark_dist_y = 0;
	b = eeprom_read_byte(0x300029);
	sensor_gain_x = b & 0x0f;                       /* edge = below gain/10 of paper */
	sensor_gain_y = b >> 4;
	sensor_scan_speed = eeprom_read_byte(0x300028) << 1;
	regmark_load_calibration();
	regmark_scan_mode = 0;
	landscape = 0;
}

/*
 * 0x206d6: sensor position relative to the knife.  Nominally 15 mm along
 * -Y (3000 units); trimmed by service g14 (EEPROM 0x20/0x22, 0.1 mm steps)
 * and by the user calibration TB72 (EEPROM 0x24/0x26, 0.05 mm steps).
 */
void regmark_load_calibration(void)
{
	scan_off_x = 0;
	scan_off_y = -3000;
	scan_off_x -= eeprom_read_word(0x300020) * 20L;
	scan_off_y -= eeprom_read_word(0x300022) * 20L;
	scan_off_x -= eeprom_read_word(0x300024) * 10L;
	scan_off_y -= eeprom_read_word(0x300026) * 10L;
}

/* 0x1f930: TB0 / TB99 */
char regmark_enable(int on)
{
	CFG_HI = (CFG_HI & ~CFG_ENABLE) | (on ? CFG_ENABLE : 0);
	return on != 0;
}

/* 0x1f79c: TB50 orientation change, keeping the media position */
void regmark_set_orientation(char new_landscape)
{
	long ox, oy, a;
	int had_offset;

	if (landscape == new_landscape)
		return;
	ox = orient_off_x;
	oy = orient_off_y;
	had_offset = ox != 0 || oy != 0;
	landscape = new_landscape;
	regmark_off_y = 0;
	regmark_off_x = 0;
	regmark_clear_transform();
	page_limits_apply(1);
	if (had_offset) {
		if (!landscape)
			ox = page_width - ox;
		else
			oy = page_width - oy;
		a = ox - (ox - oy);                     /* = oy */
		set_area_extent(a, (ox - oy) + a);      /* = (oy, ox): swapped */
	}
}

/* 0x1f8aa: TB50[,n] */
void cmd_TB50_orientation(void)
{
	long n = 0;

	hook_nop();
	motion_flush();
	wait_motion_done();
	if (!end_of_params())
		n = read_int();
	if (!regmark_active)
		regmark_set_orientation(n != 0);
}

/* 0x1f956: TB51,len -- arm length, 5..20 mm, else ignored.
 * (If the parameter is missing the original tests an uninitialised local.) */
void cmd_TB51_mark_length(void)
{
	int mm = 0;

	if (!end_of_params())
		mm = (int)(read_coord() / 200);
	if (mm > 4 && mm < 21)
		regmark_length = (char)mm;
}

/* 0x1f9a0: TB52[,n] -- n == 1 (or missing): type 1, otherwise type 2 */
void cmd_TB52_mark_type(void)
{
	long n = 1;

	if (!end_of_params())
		n = read_int();
	CFG_HI = (CFG_HI & ~CFG_TYPE_MASK) | (n == 1 ? CFG_TYPE1 : CFG_TYPE2);
}

/* 0x1f9ee: TB53,w -- bar width 60..200 units (0.3..1.0 mm). Stored, never used. */
void cmd_TB53_mark_width(void)
{
	int w = 0;

	if (!end_of_params())
		w = (int)read_coord();
	if (w > 59 && w < 201)
		regmark_width = (char)w;
}

/* 0x1fa30: TB54[,x[,y]] -- page origin relative to mark 1 (mark frame) */
void cmd_TB54_mark_offset(void)
{
	if (end_of_params())
		return;
	regmark_offset_x = read_coord();
	if (!end_of_params())
		regmark_offset_y = read_coord();
}

/* 0x1fa6c: TB55,n -- scale to the nominal mark distances given with TB2x/12x */
void cmd_TB55_flag(void)
{
	if (!end_of_params())
		CFG_HI = (CFG_HI & ~CFG_SCALE) | (read_int() ? CFG_SCALE : 0);
}

/* 0x1fa9c */
void cmd_TB70_calibrate(void)
{
	regmark_calibrate();
}

/* 0x1fab4: TB71 -> "x,y" (%5d) */
void cmd_TB71_get_calibration(void)
{
	reply_begin(1);
	reply_number_comma(5, eeprom_read_word(0x300024));
	reply_number_end(5, eeprom_read_word(0x300026));
}

static int clamp40(int v)
{
	return v > 40 ? 40 : v < -40 ? -40 : v;
}

/* 0x1faf8: TB72,x,y -- sensor/knife trim, -40..40 x 0.05 mm */
void cmd_TB72_set_calibration(void)
{
	int x = clamp40((int)read_int());
	int y = clamp40((int)read_int());

	eeprom_write_word(x, 0x300024);
	eeprom_write_word(y, 0x300026);
	regmark_load_calibration();
}

/* 0x1f90a: TB98 -> "     1" */
void cmd_TB98_query(void)
{
	reply_begin(2);
	reply_number_end(6, 1);
}

static void read_nominal_distance(void)
{
	regmark_dist_x = 0;
	regmark_dist_y = 0;
	if (end_of_params())
		return;
	regmark_dist_x = read_coord();
	if (!end_of_params())
		regmark_dist_y = read_coord();
}

/* 0x1fb80: TB22 / TB23 [,dx,dy] -- type = 1 (2 marks) or 2 (3 marks) */
void cmd_TB2x_search_manual(int type)
{
	CFG_HI = (CFG_HI & ~CFG_SEARCH_MASK) | ((type & 3) << 4);
	read_nominal_distance();
	regmark_search_manual();
}

/* 0x1fbe6: TB122 / TB123 [,dx,dy] */
void cmd_TB12x_search_auto(int type)
{
	CFG_HI = (CFG_HI & ~CFG_SEARCH_MASK) | ((type & 3) << 4);
	read_nominal_distance();
	regmark_search_auto();
}

/* 0x1f6cc: TBn[,...] -- n must fit in one byte */
void cmd_TB(void)
{
	long n = read_int();

	if (n & 0xff00)
		return;
	switch ((unsigned char)n) {
	case 0:   regmark_enable(0); break;
	case 22:
	case 23:  cmd_TB2x_search_manual((int)n - 21); break;
	case 122:
	case 123: cmd_TB12x_search_auto((int)n - 121); break;
	case 50:  cmd_TB50_orientation(); break;
	case 51:  cmd_TB51_mark_length(); break;
	case 52:  cmd_TB52_mark_type(); break;
	case 53:  cmd_TB53_mark_width(); break;
	case 54:  cmd_TB54_mark_offset(); break;
	case 55:  cmd_TB55_flag(); break;
	case 70:  cmd_TB70_calibrate(); break;
	case 71:  cmd_TB71_get_calibration(); break;
	case 72:  cmd_TB72_set_calibration(); break;
	case 98:  cmd_TB98_query(); break;
	case 99:  regmark_enable(99); break;
	}
}

/*
 * 0x1fc4c / 0x1ff2a: TB70 calibration (portrait / landscape).
 *
 * The sensor starts on the centre of a printed "+" cross.  It samples
 * the vertical bar below the crossing (E1), the horizontal bar 8 mm to one
 * side (E2), the vertical bar again 8 mm above (E3) and the horizontal bar
 * 8 mm to the other side (E4), then cuts E1-E3 and E2-E4, each lengthened
 * by 15 mm at both ends, i.e. a "+" where the sensor thinks the cross is.
 * The user measures the offset between printed and cut cross and corrects
 * it with TB72 (read back with TB71).  Ends 30 mm back, at Y = 0.
 * Returns 0 on success, 1 if a bar was not found (no reply).
 */
static int regmark_cut_cross(int landscape_order)
{
	pt_t s, e[4];
	char wide = 1;

	wait_motion_done();
	sync_device_pos();
	s = dev_pos;
	regmark_scan_mode = 1;
	move_clamped_wait(s.x, s.y, 0);
	pen_return_to_saved();
	motion_drain();
	if (!landscape_order) {
		regmark_locate_x(1, wide, &e[0]);
		if (regmark_result) goto fail;
		move_clamped_wait(e[0].x + 1600, e[0].y, 0); pen_return_to_saved(); motion_drain();
		regmark_locate_y(1, wide, &e[1]);
		if (regmark_result) goto fail;
		move_clamped_wait(e[1].x, e[1].y + 1600, 0); pen_return_to_saved(); motion_drain();
		regmark_locate_x(-1, wide, &e[2]);
		if (regmark_result) goto fail;
		move_clamped_wait(e[2].x - 1600, e[2].y, 0); pen_return_to_saved(); motion_drain();
		regmark_locate_y(-1, wide, &e[3]);
		if (regmark_result) goto fail;
	} else {
		regmark_locate_y(1, wide, &e[0]);
		if (regmark_result) goto fail;
		move_clamped_wait(e[0].x, e[0].y + 1600, 0); pen_return_to_saved(); motion_drain();
		regmark_locate_x(-1, wide, &e[1]);
		if (regmark_result) goto fail;
		move_clamped_wait(e[1].x - 1600, e[1].y, 0); pen_return_to_saved(); motion_drain();
		regmark_locate_y(-1, wide, &e[2]);
		if (regmark_result) goto fail;
		move_clamped_wait(e[2].x, e[2].y - 1600, 0); pen_return_to_saved(); motion_drain();
		regmark_locate_x(1, wide, &e[3]);
		if (regmark_result) goto fail;
	}
	regmark_scan_mode = 0;                          /* from here on: knife positions */
	regmark_extend_segment(&e[0].x, &e[2].x);
	regmark_extend_segment(&e[1].x, &e[3].x);
	move_clamped_wait(e[0].x, e[0].y, 0);
	move_clamped_wait(e[2].x, e[2].y, 1);
	move_clamped_wait(e[1].x, e[1].y, 0);
	move_clamped_wait(e[3].x, e[3].y, 1);
	move_clamped_wait(e[3].x, e[3].y, 0);
	s.x -= 6000;
	if (s.x < 0)
		s.x = 0;
	move_clamped_wait(s.x, 0, 0);
	pen_return_to_saved();
	motion_drain();
	return 0;
fail:
	regmark_scan_mode = 0;
	move_clamped_wait(s.x, s.y, 0);
	pen_return_to_saved();
	motion_drain();
	return 1;
}

int regmark_calibrate(void)
{
	return landscape ? regmark_find_and_mark() : regmark_cut_cross(0);
}

int regmark_find_and_mark(void)
{
	return regmark_cut_cross(1);
}


/* =====================================================================
 * Service t: sensor tests (only after FPGRFCC1)
 * ===================================================================== */

/*
 * 0x21dfa: trace the raw sensor along one axis.  Ramps to speed 8, stores
 * (A/D, distance from start) pairs at 0xffcf00; the first sample below
 * baseline*gain/10 is stored negated and followed by a (baseline, 0)
 * marker.  Reverses after 800 samples or 6000 units (30 mm) and stops
 * back at the start.  Returns the number of ints stored.
 */
static int sensor_trace_scan(char axis)
{
	int v[2] = { 0, 0 }, inc[2] = { 0, 0 };
	int *w = &sensor_trace;
	int n = 0, thr, d;
	char reversed = 0, marked = 0;
	pt_t start, p;

	sensor_on();
	sensor_filter_init();
	motion_get_pos(&start.x);
	if (axis == 0) {
		inc[0] = 1;
		thr = (unsigned)(sensor_baseline * sensor_gain_x) / 10;
	} else {
		inc[1] = 1;
		thr = (unsigned)(sensor_baseline * sensor_gain_y) / 10;
	}
	do {
		if (v[0] < 8)
			v[0] += inc[0];
		if (v[1] < 8)
			v[1] += inc[1];
		n++;
		*w = (int)ADC_VALUE();
		if (*w < thr && !marked) {
			*w = -*w;
			marked = 1;
			w[1] = sensor_baseline;
			w[2] = 0;
			w += 2;
		}
		w++;
		motion_get_pos(&p.x);
		d = (int)(axis == 0 ? p.x - start.x : p.y - start.y);
		*w++ = d;
		if (!reversed && (n > 800 || d > 6000)) {
			inc[0] = -inc[0];
			v[0] += inc[0];
			inc[1] = -inc[1];
			v[1] += inc[1];
			reversed = 1;
		}
		motor_run_continuous(1, v[0], v[1]);
	} while (v[0] >= 0 && v[1] >= 0);
	motor_run_continuous(1, 0, 0);
	sensor_off();
	motion_get_pos_wait(&p.x);
	motor_target_x = p.x;
	motor_target_y = p.y;
	sync_device_pos();
	return n * 2;
}

/* 0x21cd6: t3,axis -- replies "    -1" first, then traces */
void sensor_trace_start(void)
{
	long axis = read_int();

	rx_flush_masked();
	reply_begin(1);
	reply_number_end(6, -1);
	if (axis == 0 || axis == 1)
		sensor_trace_len = sensor_trace_scan((char)axis);
	sensor_trace_rd = 0;
}

/* 0x21d42: t4 -- next pair "%6d,%6d", or "-1,-1" at the end */
void sensor_trace_read(void)
{
	int *t = &sensor_trace;

	reply_begin(1);
	if (sensor_trace_rd < sensor_trace_len) {
		reply_number_comma(6, t[sensor_trace_rd]);
		reply_number_end(6, t[sensor_trace_rd + 1]);
		sensor_trace_rd += 2;
	} else {
		reply_number_comma(6, -1);
		reply_number_end(6, -1);
		rx_flush_masked();
	}
}

/* 0x21b96 (service t) */
void svc_t_sensor(void)
{
	char fail;
	pt_t e;
	long axis, speed;

	switch (read_int()) {
	case 0:                                         /* t0: LED on, reply A/D */
		sensor_measure_reply();
		break;
	case 1:                                         /* t1,axis,speed: filtered scan */
		axis = read_int();
		speed = read_int();
		cancel_request = 0;
		sensor_scan_axis((int)axis, 1, (int)speed, &fail, &e, 0);
		break;
	case 2:                                         /* t2,dir: locate a bar along X */
		regmark_locate_x((signed char)read_int(), 0, &e);
		break;
	case 3:
		sensor_trace_start();
		break;
	case 4:
		sensor_trace_read();
		break;
	}
}
