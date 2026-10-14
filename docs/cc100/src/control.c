/*
 * CC100 firmware 2.30 -- GP-GL control, query and interface commands.
 * Hand-cleaned from the Ghidra output. Original addresses are given in comments.
 * Types: int = 16 bit, long = 32 bit (Renesas H8S, advanced mode).
 */

#include "cc100.h"


/* globals (RAM) */
extern char  term1, term2;           /* 0xffb359, 0xffb35a: always ETX */
extern int   steps_per_unit;              /* 0xffb262: 20, 10, 5 or 2 steps per user unit */
extern int   num_frac1, num_frac2; /* 0xffb360, 0xffb362 */
extern long  chart_feed_len;         /* 0xffb3aa */
extern int   cur_pen;                /* 0xffb414, 1-based */
extern int   pen_speed[8];           /* 0xffb416 */
extern int   pen_force[8];           /* 0xffb436 */
extern int   pen_cutter_offset[8];   /* 0xffb446 */
extern int   pen_cutter_q[8];        /* 0xffb456 */
extern long  pen_blade_rot[8];       /* 0xffb466, degrees * 100 */
extern int   pen_FE[8];              /* 0xffb486 */
extern int   pen_FF_a[8], pen_FF_b[8]; /* 0xffb496, 0xffb4a6 */
extern int   cut_force;              /* 0xffb122, FX value, 0 = use force level */
extern unsigned char fk_flag;        /* 0xffb2bc */
extern unsigned char pen_params_dirty; /* 0xffb2be */
extern long  offset_x, offset_y;     /* 0xffb4e0, 0xffb4e4 */
extern long  win_ll_x, win_ll_y, win_ur_x, win_ur_y; /* 0xffb564.. user window */
extern long  cur_x, cur_y;           /* 0xffb584, 0xffb588 */
extern int   cur_outcode;            /* 0xffb5ca */
extern int   cur_tool;               /* 0xffb4b8 */

/* parameter store at 0x300000 */
enum { NV_SPEED = 0x300034, NV_FORCE = 0x300035, NV_OFFSET = 0x300036,
       NV_FE = 0x300038, NV_THICKNESS = 0x30003b };


/*
 * Parser and reply primitives (cmd_class, is_terminator, end_of_params,
 * read_digit, read_number, read_int, set_error, init_terminators,
 * reply_number*, cmd_error_mask) are in input.c; read_angle is in plot.c.
 */

/* 0x17648 */
void reply_point(int cmd, long x, long y)
{
	reply_begin(2);
	reply_number_comma(6, x / steps_per_unit);
	reply_number_end(6, y / steps_per_unit);
}

/* 0x1829e: pen argument is ignored, all pens get the same speed */
void set_speed(int speed)
{
	int i;

	motion_flush();
	if (speed != eeprom_read_byte(NV_SPEED))
		eeprom_write_byte(speed, NV_SPEED);
	for (i = 0; i < 8; i++)
		pen_speed[i] = speed;
	pen_params_dirty = 1;
	apply_pen_params();
}

/* 0x18316 */
void set_force(int force)
{
	int i;

	cut_force = 0;                       /* '*' cancels FX */
	if (cut_force != eeprom_read_byte(NV_THICKNESS))
		eeprom_write_byte(cut_force, NV_THICKNESS);
	motion_flush();
	if (force != eeprom_read_byte(NV_FORCE))
		eeprom_write_byte(force, NV_FORCE);
	for (i = 0; i < 8; i++)
		pen_force[i] = force;
	pen_params_dirty = 1;
	apply_pen_params();
	apply_force();
}

/* 0x074bc  !l[,n] */
void cmd_speed(void)
{
	int l = read_int();

	if (l >= 100)
		l -= 100;
	if (l < 1 || l > 10)
		return;
	if (!end_of_params())
		read_int();                  /* pen number, unused */
	set_speed(l);
}

/* 0x07570  *a,f[,n] */
void cmd_force(void)
{
	int f;

	read_int();                          /* a: ignored */
	f = read_int();
	if (f < 1 || f > 3)
		return;
	if (!end_of_params())
		read_int();                  /* pen number, unused */
	set_force(f);
}

/* 0x13c06  F / FS: length stored, feed hook 0x1005e is empty */
void cmd_chart_feed(void)
{
	chart_feed_len = read_coord();
}

/* 0x1780e '?', 0x177b0 '[', 0x17752 'U' */
void cmd_read_offset(void)      { reply_point('?', offset_x, offset_y); }
void cmd_read_lower_left(void)  { reply_point('[', win_ll_x, win_ll_y); }
void cmd_read_upper_right(void) { reply_point('U', win_ur_x, win_ur_y); }

/* 0x1769e '@' */
void cmd_read_status2(void)
{
	int s = cur_outcode & 0x2f;

	reply_begin(1);
	reply_number_end(5, s);
}

/* 0x18262 */
long gin_pen_state(void)
{
	return cur_tool * 10 + !(tool_state & 0x20);
}

/* 0x176d6 'G' (and 'C' via 0x17740) */
void cmd_gin(int cmd)
{
	pen_restore();                       /* 0x17f1e */
	reply_begin(3);
	reply_number_comma(6, cur_x / steps_per_unit);
	reply_number_comma(6, cur_y / steps_per_unit);
	reply_number_end(6, gin_pen_state());
}

void cmd_call_gin(void)
{
	cmd_gin('G');
}

/* 0x14008  SO / NP: parameters are not parsed */
void set_origin_here(void)
{
	extern long reg_mark_x;              /* 0xffb298 */

	hook_nop();
	motion_flush();
	pen_return_to_saved();
	motion_drain();
	wait_motion_done();
	sync_device_pos();
	/* 0xffb288 is non-zero when a registration-mark origin is set */
	set_carriage_origin(cutoff_radius ? reg_mark_x : 0, motor_target_y);
	sync_device_pos();
}

/* 0x08002  TT */
void cmd_TT(void)
{
	int i;

	hook_nop();
	motion_flush();
	wait_motion_done();
	tool_set(0);                         /* tool up */
	set_media_mode(1);                   /* 0x7aa4 */
	settings_apply();                    /* 0x13436 */
	for (i = 0; i < 8; i++)
		pen_cutter_offset[i] = 0;    /* 0xffb446 */
	select_pen_params();                 /* 0x1054c */
	home_carriage();                     /* 0x14ef0 */
	load_cutter_offset();                /* 0x104e6: NV 0x36, 0..45 else 0 */
}

/* 0x0a268  FCp,q[,n] */
void cmd_FC(void)
{
	int p = read_int();
	int q = read_int();

	if (p > 45 || p < 0)
		p = -1;
	if (q > 3 || q < 1)
		q = -1;
	if (!end_of_params())
		read_int();                  /* pen, unused */
	if (p != -1)
		set_cutter_offset(p);        /* 0x183c0, NV 0x36 */
	if (q != -1)
		set_cutter_offset_q(q);      /* 0x18440 */
	pen_params_dirty = 1;
	apply_pen_params();
}

/* 0x0a316  FDθ[,n]  θ in degrees with two decimals */
void cmd_FD(void)
{
	long a = -1;
	int n = 0;

	if (!end_of_params())
		a = read_angle();
	if (a < 0 || a > 60000)
		return;
	if (!end_of_params())
		n = read_int();
	if (n < 0 || n > 8)
		return;
	set_blade_rotation(a);               /* 0x184a4, all pens */
	pen_params_dirty = 1;
	apply_pen_params();
}

/* 0x0a3cc  FEp[,n]  TODO: meaning of p */
void cmd_FE(void)
{
	long p = -1;

	if (!end_of_params())
		p = read_int();
	if (p < 0 || p > 1)
		return;
	if (!end_of_params())
		read_int();                  /* pen, unused */
	set_FE(p);                           /* 0x1850a, NV 0x38 */
	pen_params_dirty = 1;
	apply_pen_params();
}

/* 0x0a46e  FFa,b[,n]  TODO: meaning */
void cmd_FF(void)
{
	long a = -1, b = -1;
	int n = 0;

	if (!end_of_params())
		a = read_int();
	if (a < 0 || a > 9)
		return;
	if (!end_of_params())
		b = read_int();
	if (b < 0 || b > 9)
		return;
	if (!end_of_params())
		n = read_int();
	if (n < 0 || n > 8)
		return;
	set_FF((long)b << 16 | (a & 0xffff)); /* 0x18598: a in r0, b in e0 */
	pen_params_dirty = 1;
	apply_pen_params();
}

/* 0x0a580  FK[n]  TODO: meaning */
void cmd_FK(void)
{
	int n = 0;

	if (!end_of_params())
		n = read_int();
	fk_flag = n != 0;
}

/*
 * No-op handlers (empty functions in ROM):
 *   $ 0x10082, I 0x1003a, K/P 0x10010, Q 0x1009a, R 0x100b2, S 0x100a6,
 *   A 0x10076, T 0x0766e, N 0x10046, ( 0x100be,
 *   LP 0x0fff8, RC 0x10022, SP 0x1008e, SA 0x10052.
 * Dispatcher no-ops: V : ; < = '
 */
