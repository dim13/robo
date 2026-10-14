/*
 * CC100 firmware 2.30 -- motion layer: speed profile, planner, segment queue,
 * stepper interrupts, jogging and tool force.
 *
 * See ../notes/motion.md. Units: 1 count = 0.05 mm = one half-step;
 * timers count at 16 MHz (phi/1); speeds in mm/s; acceleration in mm/s^2.
 * int is 16 bits and long 32 bits on the target. The struct layouts below give the
 * target offsets in comments; they are not meant to match host padding.
 */
#include <math.h>
#include "cc100.h"

/* ---- hardware ---------------------------------------------------------- */

#define REG8(a)  (*(volatile unsigned char *)(a))
#define REG16(a) (*(volatile unsigned int *)(a))
#define DA_SOLENOID   REG8(0xfffdac)    /* 8-bit D/A, tool force */
#define LATCH_STEP_X  REG16(0x600008)   /* X phase latch */
#define LATCH_STEP_Y  REG16(0x600000)   /* Y phase latch */
#define LATCH_CURRENT REG8(0x600010)    /* motor current: X bits 3..2, Y bits 1..0 */
#define TPU_X_TCR     REG8(0xffff20)
#define TPU_X_TIER    REG8(0xffff24)
#define TPU_X_TSR     REG8(0xffff25)
#define TPU_X_TCNT    REG16(0xffff26)
#define TPU_X_TGRA    REG16(0xffff28)
#define TPU_Y_TCR     REG8(0xffff30)
#define TPU_Y_TIER    REG8(0xffff34)
#define TPU_Y_TSR     REG8(0xffff35)
#define TPU_Y_TCNT    REG16(0xffff36)
#define TPU_Y_TGRA    REG16(0xffff38)

#define TIER_RUN  0x41  /* TTGE | TGIEA */
#define TIER_STOP 0x40  /* TTGE only */
#define TCR_PHI1_CLEAR_ON_TGRA 0x20

#define TIMER_HZ      16000000.0  /* ROM 0x28fba */
#define PERIOD_SCALE  2.048e8     /* ROM 0x28fc2 = 256 * 16e6 / 20 */
#define PERIOD_MAX    0xf06e      /* longer periods are split ... */
#define PERIOD_CHUNK  60000       /* ... into chunks of this size */
#define PERIOD_MIN    0x212
#define ACCEL         3000        /* mm/s^2 */
#define RAMP_MAX      40          /* profile rejected if ramp would be this long */
#define SEG_RING      8
#define CORNER_STOP   250         /* turn > 25.0 deg: no blending */
#define TURN_FULL     1800        /* "full stop" turn, 180.0 deg */

/* ---- data structures ---------------------------------------------------- */

/* Per-axis stepper state: axis_x 0xffbf12, axis_y 0xffbf56. */
struct axis {
	int  pad0;                       /* +0x00 */
	volatile int state;              /* +0x02 motor_busy_x/y: 0 idle,1 accel,2 cruise,3 decel,4 const */
	int  dir;                        /* +0x04 +1 / -1 */
	volatile unsigned int *tgra;     /* +0x06 */
	volatile unsigned char *tsr;     /* +0x0a */
	volatile unsigned char *tier;    /* +0x0e */
	int  pad12;                      /* +0x12 */
	long phase;                      /* +0x14 motor_pos_x/y, also phase index */
	long period_left;                /* +0x18 timer counts still to wait */
	long t_prev;                     /* +0x1c ramp time at previous step */
	long accel_left;                 /* +0x20 */
	long ramp_pos;                   /* +0x24 path ramp index, 10-bit fraction */
	long ramp_rate;                  /* +0x28 path counts per axis count, 10-bit fraction */
	long decel_left;                 /* +0x2c */
	long cruise_left;                /* +0x30 */
	long cruise_frac;                /* +0x34 fractional period, 16-bit */
	long cruise_period;              /* +0x38 */
	long frac_acc;                   /* +0x3c starts at 0x8000 (0.5) */
	volatile unsigned int *latch;    /* +0x40 */
};

/* Queued segment, 8 x 0x42 bytes at 0xffb7e2 (seg_ring). Fields per axis. */
struct seg_axis {
	int  state;          /* +0 (x) / +8 (y): 0 none, 1 ramped, 4 constant */
	int  dir;            /* +2 / +a */
	int  reserved;       /* +4 / +c */
	int  ramp_steps;     /* +6 / +e */
};
struct seg_axis_l {
	long start_period;   /* +0x18 / +0x2c */
	long ramp_rate;      /* +0x1c / +0x30 */
	long cruise_steps;   /* +0x20 / +0x34 */
	long time;           /* +0x24 / +0x38 */
	long cruise_period;  /* +0x28 / +0x3c */
};
struct segment {
	struct seg_axis   ax[2];   /* +0x00 */
	int  junction_idx;         /* +0x10 ramp index kept at the start junction */
	int  ramp_idx;             /* +0x12 */
	long duration;             /* +0x14 timer counts */
	struct seg_axis_l al[2];   /* +0x18 */
	unsigned char force;       /* +0x40 D/A value for the segment, 0 = unchanged */
};

/* Speed ramp: 0xffb640, pairs of (time, period*256). */
struct ramp_entry { long t; long p256; };

/* ---- globals (not in cc100.h; names from sym/motion.txt) --------------- */

extern struct axis axis_x, axis_y;               /* 0xffbf12 / 0xffbf56 */
extern struct segment seg_ring[SEG_RING];        /* 0xffb7e2 */
extern volatile long seg_head, seg_tail;         /* 0xffb7d0 / 0xffb7d4 */
extern struct ramp_entry speed_ramp[48];         /* 0xffb640 */
extern int  speed_down, speed_up, accel;         /* 0xffb7d8 / 0xffb7da / 0xffb7dc */
extern int  speed_cur, speed_start, start_speed; /* 0xffb7de / 0xffb7e0 / 0xffb626 */
extern long ramp_steps, ramp_time, max_subseg_len;          /* 0xffb60c / 10 / 14 */
extern long ramp_steps_down, ramp_time_down, max_subseg_down; /* 0xffb628 / 2c / 30 */
extern long ramp_steps_up, ramp_time_up, max_subseg_up;       /* 0xffb634 / 38 / 3c */
extern long ramp_steps_max;                      /* 0xffc39e */
extern int  corner_speed_tab[32];                /* 0xffc3e2 */
extern int  corner_ramp_idx[32];                 /* 0xffbb92 */
extern signed char dir_force_trim[3][24];        /* 0xffbbd8, from ROM 0x28fd2 */
extern const signed char rom_dir_force_trim[3][24]; /* 0x28fd2 */
extern unsigned int step_phase_table[32];        /* 0xffc3a2 */
extern int  last_dir_tenths;                     /* 0xffb61c */
extern int  backlash_cur, backlash_x;            /* 0xffb61e / 0xffb620 */
extern int  motion_loading, motion_pen_state;    /* 0xffbbd2 / 0xffbbd4 */
extern char motor_hold_off;                      /* 0xffb2c1 */
extern int  jog_ramp_idx;                        /* 0xffb622 */
extern long jog_period_left_x;                   /* 0xffb5f8 */
extern int  force_da_value, force_level;         /* 0xffb4ba / 0xffb4bc */
extern int  solenoid_level;                      /* 0xffb40e */
extern char solenoid_down, press_profile_sel;    /* 0xffb410 / 0xffb412 */
extern int  solenoid_press_settle, solenoid_prelift_delay, solenoid_lift_settle; /* 0xffb408/a/c */
extern int  force_table[4];                      /* 0xffc37c */
extern signed char media_force_pct[14];          /* 0xffc385 */
struct da_step { int ticks; unsigned char pad; unsigned char da; };
extern struct da_step press_profiles[2][3][5];   /* 0xffc2f0 */
extern struct da_step lift_profile[5];           /* 0xffc368 */
extern volatile char rx_blocked;                 /* 0xffb2f2 */

static void blend_junction(struct segment *prev, struct segment *cur, int junction);

/* ---- speed profile ------------------------------------------------------ */

/* 0x19780: build speed_ramp for the pen-down (low word) and pen-up (high word) speeds. */
long motion_build_profile(unsigned long speeds)
{
	int vd = (int)(speeds & 0xffff), vu = (int)(speeds >> 16);
	int vmax, v0 = start_speed;
	long n, i;
	double a0, x0;

	speed_down = vd;
	speed_up = vu;
	accel = ACCEL;
	if (vu > vd) {
		vmax = vu;
		if (vd < v0)
			v0 = vd;
	} else {
		vmax = vd;
		if (vu < v0)
			v0 = vu;
	}
	speed_start = v0;
	n = (long)(vmax * vmax - v0 * v0) * 10 / ACCEL;
	if (n >= RAMP_MAX)
		return -1;	/* rejected, old profile stays */
	ramp_steps = n;
	if (ramp_steps_max < n)
		ramp_steps_max = n;
	ramp_steps_down = (long)(vd * vd - v0 * v0) * 10 / ACCEL;
	ramp_steps_up = (long)(vu * vu - v0 * v0) * 10 / ACCEL;

	x0 = (double)v0 / ACCEL;		/* v0/A */
	a0 = x0 * x0;
	for (i = 0; i <= n + 8; i++) {
		double s2 = ((i + 0.5) / 10.0) / ACCEL;	/* 2*s/A, s = (i+0.5)/20 mm */
		double t = sqrt(a0 + s2) - x0;		/* seconds from v0 */
		speed_ramp[i].t = (long)(t * TIMER_HZ + 0.5);
		speed_ramp[i].p256 = (long)(PERIOD_SCALE / (v0 + ACCEL * t) + 0.5);
	}
	ramp_time_down = speed_ramp[ramp_steps_down].t;
	ramp_time_up = speed_ramp[ramp_steps_up].t;
	max_subseg_down = (long)vd * 1600 > 100000 ? 100000 : (long)vd * 1600;
	max_subseg_up = (long)vu * 1600 > 100000 ? 100000 : (long)vu * 1600;
	for (i = 0; i < 32; i++)
		corner_ramp_idx[i] = (int)((long)corner_speed_tab[i] * 10 * corner_speed_tab[i] / ACCEL);
	return 0;
}

/* 0x19764: service command 'h n' -- new start speed, rebuild. */
long svc_h_speed_profile(void)
{
	start_speed = (int)read_int();
	return motion_build_profile(((unsigned long)speed_up << 16) | (unsigned int)speed_down);
}

/* Interpolate ramp time at a 10-bit fractional index. */
static long ramp_time_at(long pos)
{
	long i = pos >> 10;
	long t0 = speed_ramp[i].t;
	return t0 + (((speed_ramp[i + 1].t - t0) * (pos & 0x3ff) + 0x200) >> 10);
}

/* ---- pen speed / force -------------------------------------------------- */

/* 0x18020: '!' level l gives l*10 mm/s for both pen states. */
void apply_pen_params(void)
{
	if (cur_pen != 0) {
		int v = pen_speed[cur_pen - 1] * 10;
		set_force_level((char)pen_force[cur_pen - 1]);
		motion_build_profile(((unsigned long)v << 16) | (unsigned int)v);
	}
	apply_force();
}

/* 0x181da */
int force_level_value(char level)
{
	if (level < 0)
		level = 0;
	if (level > 3)
		level = 3;
	return force_table[(int)level];
}

/* 0x180ec: D/A value and press-profile level from '*' level or FX force. */
void set_force_level(char level)
{
	int f1 = force_table[1], f2 = force_table[2], f3 = force_table[3];

	tool_state |= 0xf;
	if (cut_force < 1) {
		force_da_value = force_level_value(level);
		force_level = level;
		return;
	}
	if (cut_force < 15)
		force_da_value = f1 + (f2 - f1) * (cut_force - 2) / 8;
	else
		force_da_value = f2 + (f3 - f2) * (cut_force - 10) / 17;
	force_level = cut_force < 7 ? 1 : cut_force < 19 ? 2 : 3;
}

static void rx_block(void)
{
	rx_blocked = 1;
	usb_ier_shadow &= ~0x40000UL;	/* bulk OUT interrupt off (0xc000c4 bit 18) */
	usb_ier_shadow |= 10;
}

/* 0x17c1c: tool down with D/A level, or up when level == 0. */
char solenoid_set(int level)
{
	int i;

	if (level == 0) {
		if (solenoid_prelift_delay)
			delay_ticks(solenoid_prelift_delay);
		rx_block();
		for (i = 0; i < 5 && lift_profile[i].ticks; i++) {
			DA_SOLENOID = lift_profile[i].da;
			delay_ticks(lift_profile[i].ticks);
		}
		DA_SOLENOID = 0;
		rx_blocked = 0;
		delay_ticks(solenoid_lift_settle);
		solenoid_down = 0;
		solenoid_level = 0;
		return 0;
	}
	if (!blade_init_active && cut_force < 1) {
		level += media_force_pct[(int)media_type] * level / 100;
		if (level < 0)
			level = 0;
		if (level > 255)
			level = 255;
	}
	if (solenoid_down == 1) {
		solenoid_level = level;
		DA_SOLENOID = (unsigned char)level;
		return 1;
	}
	rx_block();
	delay_ticks(1);
	for (i = 0; i < 5; i++) {
		const struct da_step *s = &press_profiles[(int)press_profile_sel][force_level - 1][i];
		if (s->ticks == 0)
			break;
		DA_SOLENOID = s->da;	/* kick */
		delay_ticks(s->ticks);
	}
	rx_blocked = 0;
	solenoid_level = level;
	DA_SOLENOID = (unsigned char)level;	/* hold */
	delay_ticks(solenoid_press_settle);
	solenoid_down = 1;
	return 1;
}

/* ---- planner ------------------------------------------------------------ */

static int ring_used(void)
{
	long n = seg_head - seg_tail;
	return (int)(n < 0 ? n + SEG_RING : n);
}

static int motors_idle(void)
{
	return motor_busy_x == 0 && motor_busy_y == 0;
}

/* Path length without sqrt: max * (1 + ~0.41 (min/max)^2), error < 1 %. */
static long approx_len(long dx, long dy)
{
	long hi = dx > dy ? dx : dy, lo = dx > dy ? dy : dx, r, f;

	if (hi == 0)
		return 0;
	r = (lo << 8) / hi;
	f = ((r * (r * 0x16d + 0x3b00)) >> 16) + 0x400;
	return (hi * f) >> 10;
}

/* Fill one axis of a record. Returns the axis state. */
static int plan_axis(struct seg_axis *sa, struct seg_axis_l *sl, long d, int dir,
		     long a, int n, long cruise_t, long duration)
{
	sa->dir = dir;
	sa->reserved = 0;
	if (d == 0) {
		sa->state = 0;
		sa->ramp_steps = 0;
		sl->cruise_steps = 0;
		return 0;
	}
	sa->ramp_steps = (int)((long)n * d / a);	/* in doubles in the original */
	if (sa->ramp_steps == 0) {
		/* too short to ramp: constant speed for the whole axis move */
		sl->cruise_steps = d;
		sl->time = duration;
		sl->start_period = duration / d;
		sl->cruise_period = 0;
		sa->state = 4;
		return 4;
	}
	sl->cruise_steps = d - 2L * sa->ramp_steps;
	sl->ramp_rate = (d / 2 + (a << 10)) / d;
	sl->start_period = ramp_time_at(sl->ramp_rate);	/* time of the first axis step */
	{
		long top = ramp_time_at(sl->ramp_rate * sa->ramp_steps);
		sl->time = (ramp_time - top) * 2 + cruise_t;	/* TODO: exact terms, see 0x1a3xx */
		sl->cruise_period = top;
	}
	sa->state = 1;
	return 1;
}

/* 0x19f8c: plan the move from plan_pos to (x, y). turn = corner angle at the start (0.1 deg). */
unsigned long motion_plan_segment(long x, long y, int turn)
{
	long save_steps = ramp_steps, save_time = ramp_time, save_max = max_subseg_len;
	int dir_tenths = atan_tenths(y - plan_pos_y, x - plan_pos_x);
	int checked_resonance = 0;
	long dx, dy, a, parts, part_dx, part_dy;
	int sx, sy;

again:
	dx = x - plan_pos_x;
	dy = y - plan_pos_y;
	sx = dx < 0 ? -1 : 1;
	sy = dy < 0 ? -1 : 1;
	dx = dx < 0 ? -dx : dx;
	dy = dy < 0 ? -dy : dy;
	a = approx_len(dx, dy);
	if (a == 0)
		return 0xffff;
	parts = a > max_subseg_len ? a / max_subseg_len + 1 : 0;
	part_dx = dx;
	part_dy = dy;

	do {
		struct segment *s = &seg_ring[seg_head];
		long d_x = dx, d_y = dy, cruise, cruise_t, duration;
		int n;

		if (parts) {
			d_x = part_dx / parts;
			d_y = part_dy / parts;
			a = approx_len(d_x, d_y);
			parts--;
		}
		n = ramp_steps < (a >> 2) ? (int)ramp_steps : (int)(a >> 2);
		cruise = a - 2L * n;
		/* cruise time: cruise counts at the top-of-ramp period */
		cruise_t = (long)((double)cruise * speed_ramp[n].p256 / 256.0);
		duration = (n == ramp_steps ? ramp_time : speed_ramp[n].t) * 2 + cruise_t;
		if (parts) {
			part_dx -= d_x;
			part_dy -= d_y;
		}

		/* direction-dependent force trim, only with cutter offset and tool down */
		s->force = (unsigned char)solenoid_level;
		if (solenoid_level != 0 && cutoff_radius != 0) {
			int b = dir_tenths / 150;
			if (dir_tenths < 0)
				b--;
			b += 12;
			b = b < 0 ? 0 : b > 23 ? 23 : b;
			s->force += (unsigned char)((dir_force_trim[force_level - 1][b] * solenoid_level * 2) >> 8);
		}

		plan_axis(&s->ax[0], &s->al[0], d_x, sx, a, n, cruise_t, duration);
		/* X resonance band: average X period 22858..39999 counts (X 20..35 mm/s) */
		if (!checked_resonance && s->al[0].cruise_steps != 0) {
			long p = s->al[0].time / s->al[0].cruise_steps;
			checked_resonance = 1;
			if (p > 0x5949 && p < 40000) {
				int v = (int)(p * speed_cur / 40000);
				if (v < speed_start)
					v = speed_start;
				ramp_steps = (long)(v * v - speed_start * speed_start) * 10 / accel;
				ramp_time = speed_ramp[ramp_steps].t;
				max_subseg_len = (long)v * 1600 > 100000 ? 100000 : (long)v * 1600;
				goto again;	/* re-plan at the lower speed */
			}
		}
		plan_axis(&s->ax[1], &s->al[1], d_y, sy, a, n, cruise_t, duration);

		plan_pos_x += sx * d_x;
		plan_pos_y += sy * d_y;
		s->ramp_idx = n;
		s->junction_idx = 0;
		if (seg_head != seg_tail && turn < CORNER_STOP)
			blend_junction(&seg_ring[(seg_head - 1) & 7], s, turn);
		s->duration = duration;

		while (((seg_head + 1) & 7) == seg_tail) {	/* ring full */
			if (motors_idle())
				motion_start_next();
			else
				delay_ticks(5);
		}
		if (motors_idle() && ring_used() > 2)
			motion_start_next();
		seg_head = (seg_head + 1) & 7;
		turn = 0;	/* sub-segments join straight */
	} while (parts);

	ramp_steps = save_steps;
	ramp_time = save_time;
	max_subseg_len = save_max;
	return 0;
}

/*
 * Junction speed: the lower of both ramp indices and the corner-speed index for the turn.
 * The original writes a 0x34-byte blend record (0xffb9f2 + i*0x34) that shortens the previous
 * deceleration and this acceleration per axis; motion_start_next applies it.
 * TODO: per-axis blend arithmetic (0x1a7xx..0x1ae40) is summarised, not transcribed.
 */
static void blend_junction(struct segment *prev, struct segment *cur, int turn)
{
	int j = cur->ramp_idx < prev->ramp_idx ? cur->ramp_idx : prev->ramp_idx;

	if (corner_ramp_idx[turn >> 4] < j)
		j = corner_ramp_idx[turn >> 4];
	cur->junction_idx = j;
}

/* 0x19acc: plan the pending point and wait until everything has moved. */
void motion_drain(void)
{
	if (((seg_head + 1) & 7) == seg_tail && motors_idle())
		motion_start_next();
	if (last_target_x_bl != plan_pos_x || last_target_y != plan_pos_y)
		motion_plan_segment(last_target_x_bl, last_target_y, TURN_FULL);
	while (seg_head != seg_tail) {
		if (motors_idle())
			motion_start_next();
		else
			delay_ticks(1);
	}
	while (!motors_idle())
		delay_ticks(1);
}

/* 0x19b9c: queue a move to (x, y) in counts. Returns the move direction in 0.1 deg. */
int motion_line_to(long x, long y, int pen_down)
{
	int pen = pen_down != 0, dir, turn;
	long dx, dy;

	idle_timer_reset();
	if (pen != motion_pen_state) {
		motion_drain();
		motion_pen_state = pen;
	}
	dx = backlash_cur - plan_pos_x + x;
	dy = y - plan_pos_y;
	dir = atan_tenths(dy, dx);
	if (dx == 0 && dy == 0)
		return 0;

	if (seg_head == seg_tail) {
		turn = TURN_FULL;
	} else {
		turn = dir - last_dir_tenths;
		if (turn < 0)
			turn = -turn;
		if (turn > 1800 && turn <= 3600)
			turn = 3600 - turn;
		if (turn > CORNER_STOP) {
			if ((dx < 0 ? -dx : dx) <= 80) {
				if (motors_idle())
					motion_start_next();
			} else {
				motion_drain();	/* sharp corner on a long move: stop */
			}
		}
	}

	if (pen) {
		ramp_steps = ramp_steps_down;
		max_subseg_len = max_subseg_down;
		ramp_time = ramp_time_down;
		speed_cur = speed_down;
	} else {
		ramp_steps = ramp_steps_up;
		max_subseg_len = max_subseg_up;
		ramp_time = ramp_time_up;
		speed_cur = speed_up;
	}

	/* X backlash: offset goes toward 0 (moving -X) or backlash_x (moving +X) */
	if (dx != 0) {
		int want = dx > 0 ? backlash_x : 0;
		if (backlash_cur != want) {
			if (turn > CORNER_STOP) {
				long adx = dx < 0 ? -dx : dx, ady = dy < 0 ? -dy : dy;
				if (adx < ady * 2 && pen)	/* take up the play first */
					motion_plan_segment(want + last_target_x, last_target_y, TURN_FULL);
				turn = TURN_FULL;
				backlash_cur = want;
			} else {	/* shallow turn: ramp the offset 1..2 counts per move */
				int step = (dx < 0 ? -dx : dx) <= 2 ? 1 : 2;
				if (want > backlash_cur) {
					backlash_cur += step;
					if (backlash_cur > backlash_x)
						backlash_cur = backlash_x;
				} else {
					backlash_cur -= step;
					if (backlash_cur < 0)
						backlash_cur = 0;
				}
			}
		}
	}
	last_target_x_bl = backlash_cur + x;
	motion_plan_segment(backlash_cur + x, y, turn);
	if (!pen && motors_idle())
		motion_start_next();
	last_target_x = x;
	last_target_y = y;
	last_dir_tenths = dir;
	return dir;
}

/* ---- segment start and current control --------------------------------- */

/* Current bits for one axis from its cruise period (timer counts per count). */
static unsigned char current_bits(long period)
{
	if (period > 80000)
		return 2;	/* < 10 mm/s */
	if (period > 16000)
		return 1;	/* 10..50 mm/s */
	return 0;		/* full current */
}

static void axis_load(struct axis *ax, const struct seg_axis *sa, const struct seg_axis_l *sl)
{
	ax->dir = sa->dir;
	ax->ramp_rate = sl->ramp_rate;
	ax->ramp_pos = sl->ramp_rate;
	ax->accel_left = sa->ramp_steps;
	ax->decel_left = sa->ramp_steps;
	ax->cruise_left = sl->cruise_steps;
	ax->period_left = sl->start_period;
	ax->t_prev = sl->start_period;
	ax->frac_acc = 0x8000;
	ax->state = sa->state;
	if (sl->cruise_steps) {
		long base = sl->time;	/* total time over cruise counts */
		ax->cruise_period = base / sl->cruise_steps;
		ax->cruise_frac = ((base - ax->cruise_period * sl->cruise_steps) * 0x1000 / sl->cruise_steps) << 4;
	}
}

static void axis_start(struct axis *ax, volatile unsigned int *tgra, volatile unsigned char *tier)
{
	unsigned int p;

	if (ax->period_left < PERIOD_MAX) {
		p = (unsigned int)ax->period_left;
		ax->period_left = 0;
	} else {
		p = PERIOD_CHUNK;
		ax->period_left -= PERIOD_CHUNK;
	}
	*tgra = p;
	*tier = TIER_RUN;
}

/* 0x1b162: take the next record from the ring and start both axes. */
unsigned long motion_start_next(void)
{
	struct segment *s;
	unsigned char cur;

	if (seg_head == seg_tail) {
		port600010_shadow = 0x0a;	/* both axes hold current */
		LATCH_CURRENT = 0x0a;
		return 0xaffff;
	}
	s = &seg_ring[seg_tail];
	motion_loading = 1;
	axis_load(&axis_x, &s->ax[0], &s->al[0]);
	axis_load(&axis_y, &s->ax[1], &s->al[1]);
	/* TODO: apply blend records of this and the next junction (0x1b1f0..0x1b6a0) */

	/* shorten the start of the faster axis by half a cruise period */
	if (axis_y.accel_left + axis_y.cruise_left + axis_y.decel_left <
	    axis_x.accel_left + axis_x.cruise_left + axis_x.decel_left) {
		axis_y.period_left -= axis_y.cruise_period / 2;
		if (axis_y.period_left < PERIOD_MIN)
			axis_y.period_left = PERIOD_MIN;
	} else {
		axis_x.period_left -= axis_x.cruise_period / 2;
		if (axis_x.period_left < PERIOD_MIN)
			axis_x.period_left = PERIOD_MIN;
	}
	seg_tail = (seg_tail + 1) & 7;
	TPU_X_TIER = TIER_STOP;
	TPU_Y_TIER = TIER_STOP;
	TPU_X_TSR &= 0xfe;
	TPU_Y_TSR &= 0xfe;

	cur = (unsigned char)port600010_shadow & 0xf3;
	if (axis_x.state == 0) {
		if (!motor_hold_off)
			cur |= 8;
		axis_x.period_left = 0;
	} else {
		if (!motor_hold_off && axis_x.cruise_left)
			cur |= current_bits(axis_x.cruise_period) << 2;
		axis_start(&axis_x, &TPU_X_TGRA, &TPU_X_TIER);
	}
	cur &= 0xfc;
	if (axis_y.state == 0) {
		if (!motor_hold_off)
			cur |= 2;
		axis_y.period_left = 0;
	} else {
		if (!motor_hold_off && axis_y.cruise_left)
			cur |= current_bits(axis_y.cruise_period);
		axis_start(&axis_y, &TPU_Y_TGRA, &TPU_Y_TIER);
	}
	port600010_shadow = cur;
	LATCH_CURRENT = cur;

	if (s->force)
		DA_SOLENOID = s->force;
	motion_loading = 0;
	TPU_Y_TCNT = 0;
	TPU_X_TCNT = 0;
	return 0;
}

/* ---- stepper interrupts ------------------------------------------------- */

static void axis_step(struct axis *ax)
{
	ax->phase += ax->dir;
	*ax->latch = step_phase_table[ax->phase & 0x1f];
}

/* 0x14b5e: per-count state machine; programs the next compare value. */
void axis_profile_step(void *axis)
{
	struct axis *ax = axis;
	long t;
	unsigned int p;

	*ax->tsr &= 0xfe;
	if (ax->period_left != 0)
		goto program;	/* still in a long period chunk */

	switch (ax->state) {
	case 0:
		ax->period_left = 0xffff;
		*ax->tier = TIER_STOP;
		return;
	case 1:	/* accelerate */
		axis_step(ax);
		if (--ax->accel_left > 0) {
			ax->ramp_pos += ax->ramp_rate;
			t = ramp_time_at(ax->ramp_pos);
			ax->period_left = t - ax->t_prev;
			ax->t_prev = t;
			goto program;
		}
		if (ax->cruise_left != 0) {
			ax->period_left = ax->cruise_period;
			ax->state = 2;
			goto program;
		}
		ax->state = 3;
		goto decel;
	case 2:	/* cruise */
		axis_step(ax);
		ax->period_left = ax->cruise_period;
		if ((ax->frac_acc += ax->cruise_frac) > 0xffff) {
			ax->period_left++;
			ax->frac_acc &= 0xffff;
		}
		if (--ax->cruise_left < 1)
			ax->state = ax->decel_left ? 3 : 0;
		goto program;
	case 3:	/* decelerate */
		axis_step(ax);
		goto decel;
	case 4:	/* constant to the end */
		axis_step(ax);
		if (--ax->cruise_left < 1) {
			*ax->tier = TIER_STOP;
			ax->state = 0;
		}
		ax->period_left = ax->cruise_period;
		if ((ax->frac_acc += ax->cruise_frac) > 0xffff) {
			ax->period_left++;
			ax->frac_acc &= 0xffff;
		}
		goto program;
	default:
		goto program;
	}
decel:
	if (--ax->decel_left < 1) {
		*ax->tier = TIER_STOP;
		ax->state = 0;
	}
	ax->ramp_pos -= ax->ramp_rate;
	t = ramp_time_at(ax->ramp_pos);
	ax->period_left = ax->t_prev - t;
	ax->t_prev = t;
program:
	p = ax->period_left < PERIOD_MAX ? (unsigned int)ax->period_left : PERIOD_CHUNK;
	*ax->tgra = p;
	ax->period_left -= p;
}

/* 0x1bc6a: X compare interrupt; also services Y when its flag is not pending. */
void timer_x_isr(void)
{
	if (jog_dir_x) {
		jog_x_step();
		return;
	}
	TPU_X_TSR &= 0xfe;
	if (motion_loading == 0 && motor_busy_x != 0) {
		axis_profile_step(&axis_x);
		if (!(TPU_Y_TSR & 1) && motor_busy_y != 0)
			axis_profile_step(&axis_y);
		if (motors_idle())
			motion_start_next();
	}
}

/* 0x1bcc4: mirror image of timer_x_isr. */
void timer_y_isr(void)
{
	if (jog_dir_y) {
		jog_y_step();
		return;
	}
	TPU_Y_TSR &= 0xfe;
	if (motion_loading == 0 && motor_busy_y != 0) {
		axis_profile_step(&axis_y);
		if (!(TPU_X_TSR & 1) && motor_busy_x != 0)
			axis_profile_step(&axis_x);
		if (motors_idle())
			motion_start_next();
	}
}

/* 0x1bd1e: jog interrupt for X. Speed jog_speed_x gives 5*jog_speed_x mm/s. */
void jog_x_step(void)
{
	long target, slower, faster, period;

	TPU_X_TSR &= 0xfe;
	if (jog_period_left_x != 0) {	/* long period in chunks */
		if (jog_period_left_x > PERIOD_MAX) {
			jog_period_left_x -= PERIOD_CHUNK;
			TPU_X_TGRA = PERIOD_CHUNK;
			return;
		}
		TPU_X_TGRA = (unsigned int)jog_period_left_x;
		jog_period_left_x = 0;
		return;
	}
	motor_pos_x += jog_dir_x;
	LATCH_STEP_X = step_phase_table[motor_pos_x & 0x1f];
	if (jog_speed_x == 0 && jog_ramp_idx == 1) {	/* ramped down: stop */
		TPU_X_TIER = TIER_STOP;
		jog_dir_x = 0;
	}
	target = jog_speed_x == 0 ? 160000 : 160000L / jog_speed_x;
	slower = speed_ramp[jog_ramp_idx].t - speed_ramp[jog_ramp_idx - 1].t;
	faster = speed_ramp[jog_ramp_idx + 1].t - speed_ramp[jog_ramp_idx].t;
	period = target;
	if (target < slower) {		/* speed up one ramp index */
		if (++jog_ramp_idx > ramp_steps_max + 4)
			jog_ramp_idx = (int)ramp_steps_max + 4;
		period = speed_ramp[jog_ramp_idx].t - speed_ramp[jog_ramp_idx - 1].t;
	} else if (target > faster) {	/* slow down one ramp index */
		if (--jog_ramp_idx == 0)
			jog_ramp_idx = 1;
		else
			period = speed_ramp[jog_ramp_idx].t - speed_ramp[jog_ramp_idx - 1].t;
	}
	port600010_shadow = (port600010_shadow & 0xf3) | (current_bits(period) << 2);
	LATCH_CURRENT = (unsigned char)port600010_shadow;
	jog_period_left_x = period;
	if (period > PERIOD_MAX) {
		jog_period_left_x = period - PERIOD_CHUNK;
		TPU_X_TGRA = PERIOD_CHUNK;
	} else {
		TPU_X_TGRA = (unsigned int)period;
		jog_period_left_x = 0;
	}
}

/* ---- init --------------------------------------------------------------- */

/* 0x1af64 */
void motion_init(void)
{
	int i, j;

	REG8(0xfffde8) &= 0xdf;
	start_speed = 10;
	motion_loading = 1;
	backlash_x = eeprom_read_byte(0x30002a);
	backlash_cur = 0;
	plan_pos_x = plan_pos_y = 0;
	last_target_x = last_target_y = last_target_x_bl = 0;
	motion_pen_state = 0;
	seg_head = seg_tail = 0;
	LATCH_STEP_X = step_phase_table[0];
	LATCH_STEP_Y = step_phase_table[0];
	TPU_Y_TCR = TCR_PHI1_CLEAR_ON_TGRA;
	TPU_X_TCR = TCR_PHI1_CLEAR_ON_TGRA;
	motion_build_profile(0x00640064UL);	/* 100 mm/s up and down */
	motor_busy_x = motor_busy_y = 0;
	motor_pos_x = motor_pos_y = 0;
	axis_x.tgra = &TPU_X_TGRA;
	axis_y.tgra = &TPU_Y_TGRA;
	axis_x.tsr = &TPU_X_TSR;
	axis_y.tsr = &TPU_Y_TSR;
	axis_x.tier = &TPU_X_TIER;
	axis_y.tier = &TPU_Y_TIER;
	axis_x.latch = &LATCH_STEP_X;
	axis_y.latch = &LATCH_STEP_Y;
	jog_keys = 0;
	jog_dir_x = jog_dir_y = 0;
	jog_speed_x = jog_speed_y = 0;
	port600010_shadow = 0x0a;
	LATCH_CURRENT = 0x0a;
	for (i = 0; i < 3; i++)
		for (j = 0; j < 24; j++)
			dir_force_trim[i][j] = rom_dir_force_trim[i][j];
}
