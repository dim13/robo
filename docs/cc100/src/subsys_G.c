/*
 * CC100 firmware 2.30 -- device/motion layer, registration-mark sensor, USB helpers.
 * Range 0x1c000-0x26fff. Hand-cleaned from the Ghidra decompile; addresses in comments.
 * Target types: int = 16 bit, long = 32 bit. Units: "internal" = 0.005 mm,
 * motor counts = internal / 10 = 0.05 mm.
 *
 * Runtime library in this range (Renesas H8S C runtime). Doubles go by pointer
 * (er0 = result, er1/er2 = operands); written here as plain C arithmetic.
 *
 *   0x22a66 __dsub   0x22a8a __dadd   0x22eee __ddiv   0x234a0 __dmul   0x240a0 __dcmp
 *   0x22c6a __fsub   0x22c74 __fadd   0x236ac __fmul
 *   0x231f6 __deq    0x238bc __dne    0x23348 __dlt    0x23338 __dle    0x232d8 __dgt
 *   0x23358 __ltod   0x232e4 __itod   0x2339e __ltof   0x23316 __itof
 *   0x23174 __dtol   0x23276 __ftol   0x23202 __ftod   0x230b4 __dtof   0x238ca __dneg
 *   0x2381e __mulsi3 0x23052 __divmodsi4 0x23078 __udivmodsi4 0x23836 __dcopy
 *   0x23854 __movblk 0x238ec _INITSCT 0x22e16 __bitfield_helper
 *   0x23c1a cos  0x23d9c sin  0x23cf8 cosf  0x23e7a sinf  0x23f1e sqrt  0x23ffc sqrtf
 *   0x23948 atan 0x23a26 atan2 (likely)  0x23b3c ceil
 *   0x2414c..0x2667f: extended-precision cores and helpers (__xd_* / __xf_*), see notes.
 */

#include "cc100.h"


/* ---- externals (other address ranges) ---- */
extern void delay_ticks(int ticks);                     /* 0x12e66 */
extern long muldiv(long a, long b, long c);             /* 0x16130: a*b/c */
extern void sync_device_pos(void);                      /* 0x18f0e */
extern void planner_flush(void);                        /* 0x17f2e  TODO name */
extern void planner_kick(void);                         /* 0x19acc  TODO name */
extern void planner_line(long x, long y, int pen);      /* 0x19b9c  motion planner */
extern void tool_drive(int force);                      /* 0x17c1c  solenoid/force PWM */
extern void distance_correct(pt_t *p);               /* 0x0f4b2 */
extern void regmark_motion_emit(long x, long y, int pen); /* 0x10f4e/0x10d90/0x10686/0x10ce0 */
extern void mark_centre(pt_t *a, pt_t *b);        /* 0x05ecc  TODO exact semantics */
/* eeprom_read_byte: 0x127d8 */
/* eeprom_write_byte: 0x12786 */
extern void reply_begin(int nfields);                   /* 0x113e4 */
extern void reply_number_end(int width, long v);        /* 0x133ee */

/* ---- globals ---- */
extern char regmark_active;          /* 0xffb0aa */
extern char landscape;               /* 0xffb400 */
extern long page_width;              /* 0xffb510 */
extern long offset_x, offset_y;      /* 0xffb4e0/4  GP-GL '^' */
/* cancel_request: 0xffc226 */
extern char regmark_result;          /* 0xffc225 */
/* jog_keys: 0xffc424 */
extern int  tool_force;              /* 0xffb4ba */

long   origin_x, origin_y;           /* 0xffb4e8/c */
long   origin2_x, origin2_y;         /* 0xffc2e4/8  TODO */
long   regmark_off_x, regmark_off_y;         /* 0xffb4f0/4 */
long   orient_off_x, orient_off_y;   /* 0xffb4f8/c */
pt_t dev_pos;                     /* 0xffb54c */
long   mach_xmin, mach_ymin, mach_xmax, mach_ymax;   /* 0xffb524..0xffb530 */
long   mach_last_x, mach_last_y;     /* 0xffb534 */
int    mach_last_outcode;            /* 0xffb5c8 */
char   feed_beyond_xmax_clipped;     /* 0xffb3ba: 0 = do not clip x > xmax */
long   rm_area_xmin, rm_area_ymin, rm_area_xmax, rm_area_ymax; /* 0xffbc20..0xffbc2c */
long   rm_last_x, rm_last_y;         /* 0xffbc30 */
int    rm_last_outcode;              /* 0xffbc38 */
double rm_xx, rm_xy, rm_yy, rm_yx;   /* 0xffb072, 0xffb082, 0xffb07a, 0xffb06a */
double rm_sx, rm_sy;                 /* 0xffb092/0xffb09a */
char   rm_scale_on;                  /* 0xffb0ab */
long   feed_min_x;                   /* 0xffbc3a */
long   cutoff_tip_x, cutoff_tip_y; /* 0xffb2a8/c */
long   seg_merge_dist;               /* 0xffb2b2 */
char   deferred_valid;               /* 0xffc422 */
long   deferred_x, deferred_y;       /* 0xffbc54/8 */
int    deferred_pen;                 /* 0xffbc64 */
char   pending_up;                   /* 0xffbc68 */
long   pending_up_x, pending_up_y;   /* 0xffbc4a/e */
long   home_off_x, home_off_y;       /* 0xffc394/8 */
char   regmark_scan_mode;            /* 0xffc222 */
long   scan_off_x, scan_off_y;       /* 0xffc11c/0xffc120 */
long   motor_target_x, motor_target_y; /* 0xffbc3e/42 */
long   motor_y_bias;                 /* 0xffb3b4 */
volatile long motor_pos_x, motor_pos_y;   /* 0xffbf26/0xffbf6a, motor counts */
volatile int  motor_busy_x, motor_busy_y; /* 0xffbf14/0xffbf58 */
int    tool_state;                   /* 0xffbc46, bit 5 = tool down */
int    tool_force_applied;           /* 0xffbc66 */
int    sensor_threshold;             /* 0xffc22c */

#define TOOL_DOWN 0x20

/* ======================================================================== */
/* Cohen-Sutherland clipping (two copies: registration area, machine limits) */
/* ======================================================================== */

enum { OUT_XMAX = 1, OUT_XMIN = 2, OUT_YMAX = 4, OUT_YMIN = 8 };

/* 0x1cbc6 */
int machine_outcode(const long *p)
{
	int c = 0;
	if (p[0] > mach_xmax)
		c = OUT_XMAX;
	else if (p[0] < mach_xmin)
		c = OUT_XMIN;
	if (p[1] > mach_ymax)
		c |= OUT_YMAX;
	else if (p[1] < mach_ymin)
		c |= OUT_YMIN;
	if (!feed_beyond_xmax_clipped)
		c &= ~OUT_XMAX;     /* media may feed past the soft x limit */
	return c;
}

/* 0x1cd30 / 0x1cdaa: move p onto the line x = lim (resp. y = lim) towards q */
void machine_clip_x(long *p, const long *q, long lim)
{
	if (q[0] != p[0])
		p[1] += muldiv(q[1] - p[1], lim - p[0], q[0] - p[0]);
	p[0] = lim;
}

void machine_clip_y(long *p, const long *q, long lim)
{
	if (q[1] != p[1])
		p[0] += muldiv(q[0] - p[0], lim - p[1], q[1] - p[1]);
	p[1] = lim;
}

void track_feed_min_and_queue(long x, long y, int pen);

/* 0x1cc6c: clip endpoint p (outcode c) towards q; plot to it if it lands inside */
static int machine_clip_segment(long *p, const long *q, int pen, int c)
{
	if (c & OUT_XMAX)
		machine_clip_x(p, q, mach_xmax), c = machine_outcode(p);
	else if (c & OUT_XMIN)
		machine_clip_x(p, q, mach_xmin), c = machine_outcode(p);
	if (c & OUT_YMAX)
		machine_clip_y(p, q, mach_ymax), c = machine_outcode(p);
	else if (c & OUT_YMIN)
		machine_clip_y(p, q, mach_ymin), c = machine_outcode(p);
	if (c == 0)
		track_feed_min_and_queue(p[0], p[1], pen);
	return c;
}

/* 0x1caca: draw the segment mach_last -> (x,y) clipped to the machine limits */
static void machine_clip_plot(long x, long y, int pen)
{
	long a[2] = { mach_last_x, mach_last_y };
	long b[2] = { x, y };
	int ca = mach_last_outcode, cb;

	mach_last_x = x;
	mach_last_y = y;
	cb = machine_outcode(b);
	mach_last_outcode = cb;
	if (ca & cb)
		return;                         /* fully outside on one side */
	if (ca)                                 /* re-enter: move pen-up to the entry point */
		ca = machine_clip_segment(a, b, 0, ca);
	if (ca)
		return;
	if (cb)
		cb = machine_clip_segment(b, a, pen, cb);
	if (cb == 0)
		track_feed_min_and_queue(x, y, pen);
	/* TODO: when cb != 0 the clipped end point was already queued by
	 * machine_clip_segment; the original then skips the final call, as here. */
}

/* 0x1c146..0x1c414: the same algorithm against rm_area_* (registration area);
 * on success it continues with device_plot (0x1c492 -> 0x1c608). Not repeated. */

/* ======================================================================== */
/* Device transform                                                         */
/* ======================================================================== */

/* 0x1c824 */
static pt_t apply_origin_offsets(long x, long y, int no_user_offset)
{
	pt_t p;
	if (!no_user_offset) {
		x += offset_x;
		y += offset_y;
	}
	p.x = x + origin_x + origin2_x;
	p.y = y + origin_y + origin2_y;
	return p;
}

/* regmark_transform: see regmark.c */

/* 0x1c918 */
static pt_t apply_regmark_and_cal(pt_t p)
{
	if (regmark_active)
		regmark_transform(&p, p.x, p.y);
	p.x += regmark_off_x;
	p.y += regmark_off_y;
	return p;
}

/* 0x1c9c0: out[0..1] = oriented point */
void apply_orientation(long x, long y, long *out)
{
	dev_pos.x = x;
	dev_pos.y = y;
	x += orient_off_x;
	y += orient_off_y;
	if (landscape) {
		/* TODO: axes swapped against page_width; exact formula unclear in the decompile
		 * (compare user_to_motor_point 0x1ce76 which uses x' = W - x - y ... ). */
		long t = x;
		x = y;
		y = page_width - t;
	}
	out[0] = x;
	out[1] = y;
}

/* 0x1c608 */
void device_plot(long x, long y, int pen)
{
	pt_t p = { x, y };
	long o[2];

	p = apply_regmark_and_cal(p);
	apply_orientation(p.x, p.y, o);
	machine_clip_plot(o[0], o[1], pen);
}

/* 0x1c572 */
void plot_device(long x, long y, int pen)
{
	pt_t p = apply_origin_offsets(x, y, 0);
	if (regmark_active) {
		/* regmark_clip_plot(p.x, p.y, pen);  0x1c146: clip to rm_area_*, then device_plot */
		device_plot(p.x, p.y, pen); /* TODO: area clipping omitted here */
	} else {
		rm_last_x = p.x;
		rm_last_y = p.y;
		device_plot(p.x, p.y, pen);
	}
}

/* ======================================================================== */
/* Segment queue                                                            */
/* ======================================================================== */

void tool_set(int down);
void motion_emit(long x, long y, int pen);

/* 0x1d3c2 */
void sub_home_offset(long *x, long *y)
{
	*x -= home_off_x;
	*y -= home_off_y;
}

/* 0x1d31a */
void queue_segment_now(long x, long y, int pen)
{
	if (pending_up) {
		pending_up = 0;
		motion_emit(pending_up_x - home_off_x, pending_up_y - home_off_y, 0);
	}
	/* 0x17f12: TODO (planner bookkeeping) */
	sub_home_offset(&x, &y);
	motion_emit(x, y, pen);
}

/* 0x1da92 */
void flush_deferred_segment(void)
{
	if (deferred_valid) {
		deferred_valid = 0;
		queue_segment_now(deferred_x, deferred_y, deferred_pen);
	}
}

/* 0x1d10c: pen-up moves are made lazily; tiny pen-down moves right after are deferred */
static long merge_ref_x, merge_ref_y;   /* 0xffbc5c/0xffbc60 */

static long absl(long v) { return v < 0 ? -v : v; }

void queue_segment(long x, long y, int pen)
{
	if (pen == 0) {
		flush_deferred_segment();
		/* 0xffb2bf: one-shot flag, calls 0x10fba (TODO) */
		if (!(*(signed char *)0xffb368 < 0)) {   /* config bit 7 clear: lazy pen-up */
			pending_up = 1;
			pending_up_x = x;
			pending_up_y = y;
			tool_set(0);
		} else {
			queue_segment_now(x, y, 0);
		}
		return;
	}
	if (seg_merge_dist != 0) {
		if (!deferred_valid) {
			merge_ref_x = pending_up ? pending_up_x : cutoff_tip_x;
			merge_ref_y = pending_up ? pending_up_y : cutoff_tip_y;
		}
		if (absl(merge_ref_x - x) < seg_merge_dist &&
		    absl(merge_ref_y - y) < seg_merge_dist) {
			deferred_valid = (char)0xff;
			deferred_x = x;
			deferred_y = y;
			deferred_pen = pen;
			return;
		}
	}
	deferred_valid = 0;
	queue_segment_now(x, y, pen);
}

/* 0x1ce28 */
void track_feed_min_and_queue(long x, long y, int pen)
{
	if (pen && x < feed_min_x)
		feed_min_x = x;
	{
		pt_t p = { x, y };
		distance_correct(&p);           /* 0x1d0bc queue_point */
		queue_segment(p.x, p.y, pen);
	}
}

/* 0x1d582 */
void motion_move_steps(long x, long y, int pen)
{
	motor_target_x = x;
	motor_target_y = y;
	planner_line(x / 10, (y + motor_y_bias) / 10, pen);
}

/* 0x1d4aa */
void motion_line(long x, long y, int pen)
{
	if (regmark_scan_mode) {
		x += scan_off_x;
		y += scan_off_y;
	}
	tool_set(pen);
	motion_move_steps(x, y, pen);
}

/* 0x1d3fe */
void motion_emit(long x, long y, int pen)
{
	pending_up = 0;
	/* 0xffb2c2 = 0 */
	if (cutoff_radius == 0 || regmark_scan_mode) {
		cutoff_tip_x = x;
		cutoff_tip_y = y;
		motion_line(x, y, pen);
	} else {
		regmark_motion_emit(x, y, pen);     /* TODO: 0x10f4e path */
	}
}

/* 0x1d5fe: current position in internal units */
void motion_get_pos(long *xy)
{
	xy[0] = motor_pos_x * 10;
	xy[1] = motor_pos_y * 10 - motor_y_bias;
	/* also copied to the planner start position 0xffb5e0/4, 0xffb600/4/8 */
}

/* 0x1db52 */
void wait_motion_done(void)
{
	planner_flush();
	planner_kick();
	while (motor_busy_x || motor_busy_y)
		delay_ticks(1);
}

/* 0x1d6f8: tool (blade/pen) up or down; down uses the current force */
void tool_set(int down)
{
	int force = 0;

	if (!down) {
		flush_deferred_segment();
		if (!(tool_state & TOOL_DOWN))
			return;
		tool_force_applied = 0;
		tool_state &= ~TOOL_DOWN;
	} else {
		if ((tool_state & TOOL_DOWN) && tool_force_applied == tool_force)
			return;
		force = tool_force;
		tool_force_applied = force;
		tool_state |= TOOL_DOWN;
	}
	planner_kick();
	tool_drive(force);
}

/* ======================================================================== */
/* Rounding                                                                 */
/* ======================================================================== */

/* 0x1f3ea: double -> long, half away from zero (constants +0.5 / -0.5 at 0x290c2/0x290ca) */
long dround(double v)
{
	return (long)(v < 0 ? v - 0.5 : v + 0.5);
}

/* ======================================================================== */
/* Registration-mark sensor                                                 */
/* ======================================================================== */

#define ADC_RESULT() ((unsigned)((*(volatile unsigned char *)0xffff90 << 8) | \
                                 *(volatile unsigned char *)0xffff91) >> 6)

/* sensor_on: see regmark.c */

/* sensor_off: see regmark.c */

/* sensor_pot_shift: see regmark.c */

/* sensor_pot_set_level: see regmark.c */

/* sensor_init: see regmark.c */

/* sensor_autocal_gain: see regmark.c */

/* sensor_reply_adc: see regmark.c */

/* sensor_measure_reply: see regmark.c */

/* ======================================================================== */
/* USB helpers                                                              */
/* ======================================================================== */

/* 0x2292e: big-endian bytes -> integer */
unsigned long be_bytes_to_ulong(const unsigned char *p, char n)
{
	unsigned long v = *p;
	while (--n > 0)
		v = v << 8 | *++p;
	return v;
}

/* 0x22986: little-endian bytes -> integer */
unsigned long le_bytes_to_ulong(const unsigned char *p, char n)
{
	unsigned long v = 0;
	while (n)
		v = v << 8 | p[(unsigned char)--n];
	return v;
}
