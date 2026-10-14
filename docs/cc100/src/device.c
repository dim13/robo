/*
 * CC100 firmware 2.30 -- device-specific GP-GL commands (F*, TT, TB) and
 * service mode, hand-cleaned from the Ghidra decompilation.
 *
 * int is 16-bit, long is 32-bit (H8S, Renesas C compiler).
 * Coordinates are internal units (0.005 mm, 200 per mm) unless noted.
 * Original ROM addresses are given in brackets.
 */

#include "cc100.h"

/* Shared functions and RAM globals are declared in cc100.h. */

static const char fw_version[] = "V2.30    ";   /* [28cec] */

/* ---- serial EEPROM ------------------------------------------------------ */

#define EE(off) (0x300000L + (off))

enum {
	EE_DIST_CORR_A = 0x1c, EE_DIST_CORR_B = 0x1e,
	EE_REG_CAL_X   = 0x24, EE_REG_CAL_Y   = 0x26,
	EE_SPEED = 0x34, EE_ACCEL = 0x35, EE_FQ2 = 0x36, EE_ORIENT = 0x37,
	EE_FQ3 = 0x38, EE_TRACK_PASSES = 0x39, EE_MEDIA = 0x3a,
	EE_FORCE = 0x3b, EE_TRACK_OFF = 0x3c,
};

static void eeprom_update_byte(char v, long addr)
{
	if (eeprom_read_byte(addr) != (unsigned char)v)
		eeprom_write_byte(v, addr);
}

/* Reply helpers (reply_begin, reply_putc, reply_number*) are in input.c. */

/* [0ffbc]+[1108a]+[1db52]: flush the queue and wait until motion is done */
static void wait_motion(void)
{
	hook_nop();
	motion_flush();
	wait_motion_done();
}

static void reply_term(void)
{
	reply_putc(term1);
	if (term1 != term2)
		reply_putc(term2);
	reply_send();                   /* [114ba] start USB/serial transmit */
}

/* ---- F* commands -------------------------------------------------------- */

/* [0ff4e] FG: firmware version */
void cmd_FG_version(void)
{
	const char *p;

	reply_begin(1);
	for (p = fw_version; *p; p++)
		reply_putc(*p);
	reply_term();
}

/* [0767a] FA: read distance correction */
void cmd_FA_get_distance_corr(void)
{
	reply_begin(2);
	reply_number_comma(5, eeprom_read_word(EE(EE_DIST_CORR_A)));
	reply_number_end(5, eeprom_read_word(EE(EE_DIST_CORR_B)));
}

static int clamp(int v, int lim) { return v > lim ? lim : v < -lim ? -lim : v; }

/* [076be] FBa,b: set distance correction, -200..200 each */
void cmd_FB_set_distance_corr(void)
{
	long dev[2] = { 0, 0 };
	int a, b;

	if (orient_off_x || orient_off_y) {     /* keep logical position across rescale */
		long x = orient_off_x, y = orient_off_y;
		orient_off_x = orient_off_y = 0;
		apply_orientation(x, y, dev);   /* [1c9c0] */
		distcorr_to_user(dev);          /* [0f4b2] */
	}
	a = clamp(read_int(), 200);
	b = clamp(read_int(), 200);
	eeprom_write_word(a, EE(EE_DIST_CORR_A));
	eeprom_write_word(b, EE(EE_DIST_CORR_B));
	load_axis_correction();         /* [12040] */
	sync_device_pos();              /* [18f0e] */
	if (dev[0] || dev[1]) {
		distcorr_to_device(dev);        /* [0f61e] */
		if (landscape) {
			long t = dev[1];
			dev[1] = dev[0];
			dev[0] = page_width - t;
		}
		orient_off_x = dev[0];
		orient_off_y = dev[1];
	}
}

/* [07a3a] FL: query media mode */
void cmd_FL_get_media_mode(void)
{
	reply_begin(1);
	reply_number_end(6, media_mode);
}

/* [07aa4] mode 0: paper (5 mm Y offset, 190 mm), mode 1: mat (no offset, 200 mm) */
void set_media_mode(char mode)
{
	long x, y, *w;
	int had_pos;

	if (media_mode == mode)
		return;
	hook_nop();
	motion_flush();
	wait_motion_done();
	x = orient_off_x;               /* [ffb4f8] logical position */
	y = orient_off_y;
	had_pos = x || y;
	media_mode = mode;
	w = landscape ? &x : &y;        /* the width axis */
	if (mode == 0) {
		motor_y_bias = 1000;    /* [ffb3b4] */
		page_width = 38000;
		*w -= 1000;
		if (*w < 0)
			*w = 0;
		else if (*w > page_width)
			*w = page_width - 1;
	} else {
		motor_y_bias = 0;
		page_width = 40000;
		*w += 1000;
	}
	page_limits_update();           /* [14ff4] */
	pen_reset();                    /* [17e8e] */
	page_limits_apply(1);           /* [150dc] */
	motion_get_pos_wait(&motor_target_x);   /* [1d5e0] -> ffbc3e */
	sync_device_pos();
	if (had_pos)
		set_area_extent(x, y);      /* [07984] */
}

/* [07a64] FMn */
void cmd_FM_set_media_mode(void)
{
	long n = read_int();
	if (!regmark_active)
		set_media_mode(n != 0);
}

/* [201f8] FN[n]: orientation, persisted */
void cmd_FN_orientation(void)
{
	char l = end_of_params() == 0 ? read_int() != 0 : 0;

	eeprom_update_byte(l, EE(EE_ORIENT));
	set_orientation(l);             /* [20268] */
}

/* [07c82] FO[len]: page length; 0 = 1000 mm */
void cmd_FO_page_length(void)
{
	extern long regmark_origin_x;           /* [ffb3fc] */
	extern char page_set;                   /* [ffb3ba] */
	long n = end_of_params() == 0 ? read_coord() : 0;

	if (!regmark_active) {
		page_length = n ? n : 200000;
	} else {
		/* only for manual mark search in landscape */
		if ((regmark_cfg >> 8 & 0x30) != 0x10 || !landscape)
			return;
		page_length = n + regmark_origin_x;
	}
	mach_xmin = mach_ymin = 0;              /* [ffb524] [ffb528] */
	mach_xmax = page_length;                /* ffb50c.. -> ffb52c.. (8-byte copy) */
	mach_ymax = page_width;
	page_set = 1;
}

/* FP (service unlock, [0d46c]) is in input.c. */

/* [07d46] FQn: query stored parameters */
void cmd_FQ_query(void)
{
	unsigned char v;

	switch (read_int()) {
	case 0: v = eeprom_read_byte(EE(EE_SPEED)); if (v == 0 || v > 10) v = 10; break;
	case 1: v = eeprom_read_byte(EE(EE_ACCEL)); if (v == 0 || v > 3) v = 3; break;
	case 2: v = eeprom_read_byte(EE(EE_FQ2));   if (v > 45) v = 0; break;
	case 3: v = eeprom_read_byte(EE(EE_FQ3));   break;
	case 4: v = eeprom_read_byte(EE(EE_ORIENT)); break;
	case 5: v = regmark_active; break;      /* 1 after successful mark search */
	default: return;
	}
	reply_begin(1);
	reply_number_end(5, v);
}

/* [148ce] FT: move to origin, tool up */
void cmd_FT_origin(void)
{
	plot_to(0, 0, 0);
}

/* [148e8] FUlen,w: roll the media back and forth over `len` to form
 * tracks for thick media. Width is read and ignored. */
void cmd_FU_track_feed(void)
{
	long len = read_coord();
	unsigned char passes, i;

	(void)read_coord();
	if (cut_force < 1) {
		if (pen_force[cur_pen - 1] != 3)        /* '*' level 3 = heavy media */
			return;
	} else if (cut_force < 19) {
		return;
	}
	if (track_enhance_off == 1 || (passes = eeprom_read_byte(EE(EE_TRACK_PASSES))) == 0)
		return;
	if (!regmark_active || ((regmark_cfg >> 8 & 0x30) == 0x10 && landscape)) {
		long p[2] = { len, 0 };

		distcorr_to_user(p);
		len = p[0];
		pen_save_position();
		motion_build_profile(0x640064);         /* feed speed */
		for (i = 0; i < passes; i++) {
			motion_line(-feed_total, 0, 0); /* [1d4aa] */
			motion_line(len - feed_total, 0, 0);
		}
		motion_line(-feed_total, 0, 0);
		pen_return_to_saved();
		wait_motion_done();
		apply_pen_params();
	} else {
		/* registration active: shuttle between the stored mark-area corners */
		motion_build_profile(0x640064);
		for (i = 0; i < passes; i++) {
			device_plot(rm_area_xmin, rm_area_ymin, 0);
			pen_return_to_saved();
			if (!landscape)
				device_plot(rm_area_xmax, rm_area_ymin, 0);
			else
				device_plot(rm_area_xmin, rm_area_ymax, 0);
			pen_return_to_saved();
		}
		device_plot(rm_last_x, rm_last_y, 0);
		motion_flush();
		apply_pen_params();
	}
}

/* [07ec8] FV: query media type (default 13) */
void cmd_FV_get_media(void)
{
	unsigned char v = eeprom_read_byte(EE(EE_MEDIA));
	if (v == 0 || v > 13)
		v = 13;
	reply_begin(1);
	reply_number_end(5, v);
}

/* [07f1a] FWn: media type 1..13 */
void cmd_FW_set_media(void)
{
	int n = read_int();
	if (n < 1 || n > 13)
		n = 13;
	wait_motion();
	eeprom_update_byte(n, EE(EE_MEDIA));
	media_type = n;
}

/* [07f7a] FXn[,m]: cutting force 1..30, m ignored */
void cmd_FX_force(void)
{
	int n;

	motion_flush();
	n = read_int();
	(void)read_int();
	if (n < 1)
		n = 1;
	if (n > 29)
		n = 30;
	cut_force = n;
	eeprom_update_byte(n, EE(EE_FORCE));
	pen_params_dirty = 1;           /* [ffb2be] */
	pen_select_hook();              /* [17992] */
	apply_pen_params();             /* [18020] */
	apply_force();                  /* [180b2] */
}

/* [14b12] FYn: 1 = track enhancing off */
void cmd_FY_track_enhance(void)
{
	track_enhance_off = read_int() != 0;
	eeprom_update_byte(track_enhance_off, EE(EE_TRACK_OFF));
}

/* [14068] 10 mm square with a triangle inside, at the current position */
void draw_test_cut_mark(void)
{
	long x = cur_x, y = cur_y;              /* [ffb584] [ffb588] */

	plot_solid(x + 200,  y + 400,  0);      /* [1eafe] */
	plot_solid(x + 1800, y + 400,  1);
	plot_solid(x + 1000, y + 1800, 1);
	plot_solid(x + 200,  y + 400,  1);
	plot_solid(x,        y,        0);
	plot_solid(x + 2000, y,        1);
	plot_solid(x + 2000, y + 2000, 1);
	plot_solid(x,        y + 2000, 1);
	plot_solid(x,        y,        1);
	plot_solid(x,        y,        0);
	wait_motion_done();
	sync_device_pos();
}

/* [141d2] FH: test cut, then step 10.5 mm forward */
void cmd_FH_test_cut(void)
{
	long x, y;

	sync_device_pos();
	x = cur_x;
	y = cur_y;
	pending_up = 0;                 /* [ffbc68] */
	draw_test_cut_mark();
	plot_solid(x + 2100, y, 0);
	motion_flush();
	sync_device_pos();
	tool_set(0);                    /* tool up */
	pen_return_to_saved();
	motion_drain();
}

/* [14246] FI: endless burn-in plot (radial lines, zig-zags, circles,
 * bars), never returns. See notes/device.md. */
void cmd_FI_test_plot_loop(void);

/* TT ([08002]) is in control.c. */

/* ---- TB: registration marks -------------------------------------------- */

/*
 * regmark_cfg [ffc100], bit 15 = MSB:
 *   15     apply transform after search (TB99 set / TB0 clear)
 *   14     TB55 flag
 *   13..12 search type 1/2 (TB22,TB122 / TB23,TB123)
 *   11..10 mark type 1/2 (TB52)
 */
#define RM_APPLY        0x8000
#define RM_FLAG55       0x4000
#define RM_SEARCH(v)    ((v) << 12)
#define RM_SEARCH_MASK  0x3000
#define RM_TYPE(v)      ((v) << 10)
#define RM_TYPE_MASK    0x0c00

/* [1f930] */
char regmark_enable(int on)
{
	regmark_cfg = (regmark_cfg & ~RM_APPLY) | (on ? RM_APPLY : 0);
	return on != 0;
}

static void read_optional_xy(long *x, long *y)
{
	*x = *y = 0;
	if (end_of_params() == 0) {
		*x = read_coord();
		if (end_of_params() == 0)
			*y = read_coord();
	}
}

extern long regmark_org_x, regmark_org_y;       /* [ffc112] [ffc116] mark origin */
extern long mark_off_x, mark_off_y;             /* [ffc102] [ffc106] TB54 mark offset */

/* [1fb80] TB22 / TB23 [,x,y]: manual search, tool positioned by user */
void cmd_TB2x_search_manual(int type)
{
	regmark_cfg = (regmark_cfg & ~RM_SEARCH_MASK) | RM_SEARCH(type);
	read_optional_xy(&regmark_org_x, &regmark_org_y);
	regmark_search_manual();        /* [1575a] replies "%5d" result */
}

/* [1fbe6] TB122 / TB123 [,x,y]: automatic search */
void cmd_TB12x_search_auto(int type)
{
	regmark_cfg = (regmark_cfg & ~RM_SEARCH_MASK) | RM_SEARCH(type);
	read_optional_xy(&regmark_org_x, &regmark_org_y);
	regmark_search_auto();          /* [15fb2] replies "%5d" result */
}

/* [1f8aa] TB50[,n]: orientation for registration, n != 0 = landscape */
void cmd_TB50_orientation(void)
{
	long n;

	wait_motion();
	n = end_of_params() == 0 ? read_int() : 0;
	if (!regmark_active)
		regmark_set_orientation(n != 0);    /* [1f79c] */
}

/* [1f956] TB51,len: mark length 5..20 mm */
void cmd_TB51_mark_length(void)
{
	int mm = 0;     /* TODO: uninitialised in the original when no param */

	if (end_of_params() == 0)
		mm = read_coord() / 200;
	if (mm >= 5 && mm <= 20)
		regmark_length = mm;
}

/* [1f9a0] TB52[,n]: mark type, 1 if n == 1 (default), else 2 */
void cmd_TB52_mark_type(void)
{
	int n = end_of_params() == 0 ? read_int() : 1;

	regmark_cfg = (regmark_cfg & ~RM_TYPE_MASK) | RM_TYPE(n == 1 ? 1 : 2);
}

/* [1f9ee] TB53,w: mark line width 60..200 internal (0.3..1 mm), unused */
void cmd_TB53_mark_width(void)
{
	int w = 0;      /* TODO: uninitialised in the original */

	if (end_of_params() == 0)
		w = (int)read_coord();
	if (w >= 60 && w <= 200)
		regmark_width = w;
}

/* [1fa30] TB54[,x[,y]]: mark offset */
void cmd_TB54_mark_offset(void)
{
	if (end_of_params() == 0) {
		mark_off_x = read_coord();
		if (end_of_params() == 0)
			mark_off_y = read_coord();
	}
}

/* [1fa6c] TB55,n */
void cmd_TB55_flag(void)
{
	if (end_of_params() == 0)
		regmark_cfg = (regmark_cfg & ~RM_FLAG55) | (read_int() ? RM_FLAG55 : 0);
}

/* [1fab4] TB71: calibration offsets */
void cmd_TB71_get_calibration(void)
{
	reply_begin(2);
	reply_number_comma(5, eeprom_read_word(EE(EE_REG_CAL_X)));
	reply_number_end(5, eeprom_read_word(EE(EE_REG_CAL_Y)));
}

/* [1faf8] TB72,x,y: calibration offsets -40..40 (x10 internal units) */
void cmd_TB72_set_calibration(void)
{
	int x = clamp(read_int(), 40);
	int y = clamp(read_int(), 40);

	eeprom_write_word(x, EE(EE_REG_CAL_X));
	eeprom_write_word(y, EE(EE_REG_CAL_Y));
	regmark_load_calibration();     /* [206d6] */
}

/* [1f90a] TB98: constant reply 1 */
void cmd_TB98_query(void)
{
	reply_begin(1);
	reply_number_end(6, 1);
}

/* [1f6cc] TBn dispatcher */
void cmd_TB(void)
{
	int n = read_int();

	if (n & 0xff00)
		return;
	switch (n) {
	case 0:   regmark_enable(0); break;
	case 22:
	case 23:  cmd_TB2x_search_manual(n - 21); break;
	case 122:
	case 123: cmd_TB12x_search_auto(n - 121); break;
	case 50:  cmd_TB50_orientation(); break;
	case 51:  cmd_TB51_mark_length(); break;
	case 52:  cmd_TB52_mark_type(); break;
	case 53:  cmd_TB53_mark_width(); break;
	case 54:  cmd_TB54_mark_offset(); break;
	case 55:  cmd_TB55_flag(); break;
	case 70:  regmark_calibrate(); break;   /* [1fa9c] -> [1fc4c] */
	case 71:  cmd_TB71_get_calibration(); break;
	case 72:  cmd_TB72_set_calibration(); break;
	case 98:  cmd_TB98_query(); break;
	case 99:  regmark_enable(99); break;
	}
}

/* [20616] power-on / reset defaults */
void regmark_init(void)
{
	unsigned char b;

	regmark_cfg = (regmark_cfg & ~(RM_APPLY | RM_FLAG55 | RM_SEARCH_MASK | RM_TYPE_MASK))
	            | RM_TYPE(2);
	regmark_length = 10;
	mark_off_x = mark_off_y = 0;
	regmark_org_x = regmark_org_y = 0;
	b = eeprom_read_byte(EE(0x29));
	sensor_gain_x = b & 0x0f;            /* [ffc124] */
	sensor_gain_y = b >> 4;              /* [ffc126] */
	sensor_scan_speed = eeprom_read_byte(EE(0x28)) << 1;
	regmark_load_calibration();
	regmark_scan_mode = 0;                       /* [ffc222] */
	landscape = 0;
}

/* [206d6] sensor-to-tool offset from EEPROM */
void regmark_load_calibration(void)
{
	scan_off_x = 0;                       /* [ffc11c] */
	scan_off_y = -3000;                   /* [ffc120] */
	scan_off_x -= eeprom_read_word(EE(0x20)) * 20;
	scan_off_y -= eeprom_read_word(EE(0x22)) * 20;
	scan_off_x -= eeprom_read_word(EE(EE_REG_CAL_X)) * 10;
	scan_off_y -= eeprom_read_word(EE(EE_REG_CAL_Y)) * 10;
}

/*
 * [1575a] / [15fb2] regmark_search_manual / regmark_search_auto (summary):
 *   wait, media mode 1, clear transform, set page 1100 mm x 200 mm,
 *   (manual: start at current position; auto: scan for the first mark [15bbc]),
 *   detect remaining marks with the type-1 [15212] or type-2 [154ca] detector,
 *   if RM_APPLY build the affine transform [054e4] (sets regmark_active),
 *   on failure return to start and restore 1000 mm x 200 mm page,
 *   finally: reply_begin(1); reply_number_end(5, regmark_result);
 */

/* ---- service mode ------------------------------------------------------- */

/*
 * Lowercase letters are dispatched in gpgl_loop only while service_mode != 0
 * (set by FPGRFCC1, cleared by reset). See notes/device.md for the table.
 * The 'g' family uses: gN,0 -> reply "%5d" current value; gN,1,v -> store v.
 */

/* [0d512] demo_run: entered by a long key press (IRQ2 handler [12ff2]);
 * after a mechanical self test it feeds the ROM GP-GL program at
 * [0db9a .. 0f096) (5372 bytes, a papercraft box net) into the interpreter
 * by pointing the input buffer at it. */
void demo_run(void)
{
	demo_mode = 1;                  /* [ffc2da] peek() re-runs the demo when idle */
	/* ... mechanical self test elided, see decompiled.c ... */
	rx_rd_ptr = (unsigned char *)0xdb9a;    /* ROM GP-GL program */
	rx_count = 0xf096 - 0xdb9a;

	gpgl_loop();
}
