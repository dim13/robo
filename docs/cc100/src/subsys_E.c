/*
 * CC100 firmware 2.30 -- readable C for selected routines in 0x00400-0x10fff.
 * Hand-cleaned from the Ghidra decompile (H8S, int = 16 bit, long = 32 bit).
 * Original addresses are given as / * 0xNNNN * / comments. See notes/subsys_E.md.
 *
 * Trivial wrappers and empty handlers are only listed in the table at the end.
 */

#include "cc100.h"

typedef unsigned char  u8;
typedef unsigned short u16;
typedef unsigned long  u32;

/* ---- externals named elsewhere (symbols.txt) ---- */
extern unsigned int rx_getc(void);                     /* 0x11368 */
extern unsigned int rx_peek(void);                     /* 0x11228 */
extern void reply_begin(int via);                      /* 0x113e4 */
extern void reply_putc(char c);                        /* 0x1141e */
extern void reply_send(void);                          /* 0x114ba */
extern long read_int(void);                            /* 0x11b72 */
extern void reply_number_comma(int w, long v);         /* 0x133be */
extern void reply_number_end(int w, long v);           /* 0x133ee */
extern void plot_to(long x, long y, int pen);          /* 0x1db80 */
extern void plot_solid(long x, long y, int pen);       /* 0x1eafe */
extern void clip_reset(void);                          /* 0x110ba */
extern void rx_flush_masked(void);                     /* 0x11852 */
/* eeprom_read_byte: 0x127d8 */
extern void tool_up_wait(void);                        /* 0x1d6f8 */
extern void motion_idle(void);                         /* 0x17f2e */
extern void device_move(long x, long y, int pen);      /* 0x1d4aa */
extern unsigned int pen_speed_code(u8 v);              /* 0x181da */
extern void hook_nop(void);                            /* 0x0ffbc */
extern void term_changed_hook(char t);                 /* 0x100ca, empty */

extern char fw_version_string[];                       /* 0x28cec */
extern long cur_x, cur_y;                              /* 0xffb584/88 */
extern char term1, term2;                              /* 0xffb359/5a */
extern int  cur_pen;                                   /* 0xffb414 */
extern u8   pen_params_dirty, fk_flag;

/* jog_keys: 0xffc424 */
extern int  kana_enabled;                              /* 0xffb26c */

/* per-pen tables (8 entries), 0xffb416.. */
extern int  pen_speed[8], pen_tab_b426[8], pen_force[8], pen_cutter_offset[8],
            pen_cutter_q[8], pen_FE[8], pen_FF_a[8], pen_FF_b[8];
extern long pen_blade_rot[8];                          /* 0xffb466: swivel threshold, 1/1000 deg */

/* =========================================================================
 * Work buffer: 6 KiB at 0xffc800 split into 6 regions
 * ========================================================================= */
extern u8   workbuf[0x1800];                           /* 0xffc800 */
extern long workbuf_sizes[6];                          /* 0xffb104 */
extern u8  *workbuf_ptrs[7];                           /* 0xffb0e8.. 0xffb100 */
extern u8  *workbuf_end4;                              /* 0xffb11c */
static const long workbuf_default_sizes[6] = { 3072, 1024, 0, 0, 0, 0 }; /* ROM 0x2660a */

/* 0x6ba8 */
void workbuf_fix_sizes(void)
{
	long sum = 0;
	int i;

	for (i = 0; i < 6; i++)
		sum += workbuf_sizes[i];
	workbuf_sizes[0] += 0x1800 - sum;      /* region 0 absorbs the rest */
}

/* 0x6aac */
void workbuf_layout(void)
{
	u8 *p = workbuf;

	workbuf_ptrs[0] = p;  p += workbuf_sizes[0];
	workbuf_ptrs[1] = p;  p += workbuf_sizes[1];
	workbuf_ptrs[2] = p;  p += workbuf_sizes[2];
	workbuf_end4    = p - 4;
	workbuf_ptrs[3] = p;  p += workbuf_sizes[3];
	workbuf_ptrs[4] = p;
	workbuf_ptrs[5] = p;  p += workbuf_sizes[5];  /* sizes[4] is not used */
	workbuf_ptrs[6] = p;
}

/* 0x6c3c */
void workbuf_init(void)
{
	int i;

	for (i = 0; i < 6; i++)
		workbuf_sizes[i] = workbuf_default_sizes[i];
	workbuf_fix_sizes();
	workbuf_layout();
	clip_reset();
	rx_flush_masked();
}

/* =========================================================================
 * Geometry helpers
 * ========================================================================= */

/* 0x7378 */
long labs_(long v) { return v < 0 ? -v : v; }

/* 0x6c76: octagonal distance approximation (no sqrt), used by bezier_draw */
long approx_dist(long dx, long dy)
{
	long a = labs_(dx), b = labs_(dy);

	if (a < b) {
		if (b < a * 4)
			return b + a / 2 - (b / 4) / 2;
		return b;
	}
	if (a < b * 4)
		return a + (b / 2 - (a / 4) / 2);
	return a;
}

/*
 * 0x4100: sort intersection points (x,y pairs) along the clipped segment.
 * mode bit0: key is y (segment steeper than 45 deg), bit1: descending.
 */
void clip_sort_hits(char mode, long n, long *pts)
{
	int key = mode & 1, desc = mode & 2;
	long i, j, t;

	if (n < 2)
		return;
	for (i = 0; i < (n - 1) * 2; i += 2)
		for (j = i + 2; j <= (n - 1) * 2; j += 2) {
			long a = pts[i + key], b = pts[j + key];
			if (desc ? a < b : b < a) {
				t = pts[i];     pts[i]     = pts[j];     pts[j]     = t;
				t = pts[i + 1]; pts[i + 1] = pts[j + 1]; pts[j + 1] = t;
			}
		}
}

/* 0x5452: fetch a detected registration mark */
extern long regmark_marks[6];                          /* 0xffb0ae: m1.xy, m2.xy, m3.xy */
void regmark_get_mark(int idx, long *x, long *y)
{
	if (idx > 255)
		return;                                /* high byte must be 0 */
	switch (idx) {
	case 0: case 1: *x = regmark_marks[0]; *y = regmark_marks[1]; break;
	case 2:         *x = regmark_marks[2]; *y = regmark_marks[3]; break;
	case 3:         *x = regmark_marks[4]; *y = regmark_marks[5]; break;
	}
}

/* 0xd4c8: move in millimetres (200 internal units per mm), pen up */
void move_mm(long x_mm, long y_mm)
{
	device_move(x_mm * 200, y_mm * 200, 0);
}

/* =========================================================================
 * Catmull-Rom spline segment for Y / _  (16.16 fixed point)
 * ========================================================================= */

/* 0x9ebc: coefficients for one axis of the window p[0..3]; returns a step hint */
long spline_coeffs(long *c, const long *p)
{
	long d1 = p[1] - p[0], d2 = p[2] - p[0], d3 = p[3] - p[0];

	c[0] = (d1 * 3 - d2 * 3 + d3) * 4;
	c[1] = (d2 * 4 - d1 * 5 - d3) * 2;
	c[2] = d2;
	c[3] = p[1];
	return labs_(d1 - d2) / 64;
}

/* 0x9fc4: Horner evaluation, t in 16.16, rounded multiply */
long spline_eval_fx(const long *c, unsigned int t)
{
	long v = c[0];
	int i;

	for (i = 1; i < 4; i++) {
		u32 m = (u32)labs_(v);
		u32 r = (m >> 16) * t + (((m & 0xffff) * t + 0x8000) >> 16);
		v = (v < 0 ? -(long)r : (long)r) + c[i];
	}
	return v;
}

/*
 * 0x9d24: draw the segment p1..p2 of a 4-point window, then shift the window.
 * t runs 0..0x7fff (i.e. 0..0.5 in 16.16), which supplies the 1/2 factors
 * of the Catmull-Rom basis.
 */
void spline_segment(long *xs, long *ys, int pen)
{
	long cx[4], cy[4];
	u32 nx = spline_coeffs(cx, xs), n = spline_coeffs(cy, ys);
	long step;
	int t;

	xs[0] = xs[1]; xs[1] = xs[2]; xs[2] = xs[3];
	ys[0] = ys[1]; ys[1] = ys[2]; ys[2] = ys[3];
	if (n < nx)
		n = nx;
	if (n < 4)
		n = 4;
	step = 0x10000L / n;
	t = 0;
	do {
		plot_to(spline_eval_fx(cx, t), spline_eval_fx(cy, t), pen);
		pen = -1;                              /* continue current pen state */
		t += (int)((step + 1) >> 1);
	} while (t >= 0);                              /* until 16-bit overflow past 0x7fff */
	plot_to(spline_eval_fx(cx, 0x8000), spline_eval_fx(cy, 0x8000), -1); /* TODO: exact final t */
}

/* =========================================================================
 * Pen tables (0x10332) and pen selection (0x1054c)
 * ========================================================================= */

/* 0x10332 */
void pen_tables_init(void)
{
	signed char speed = eeprom_read_byte(0x300034);
	signed char force = eeprom_read_byte(0x300035);
	signed char fe    = eeprom_read_byte(0x300038);
	int i;

	if (speed < 1 || speed > 10) speed = 10;
	if (force < 1 || force > 3)  force = 3;
	if (fe < 0 || fe > 1)        fe = 0;
	for (i = 0; i < 8; i++) {
		pen_speed[i]         = speed;
		pen_tab_b426[i]      = 1;
		pen_force[i]         = force;
		pen_cutter_offset[i] = 0;
		pen_cutter_q[i]      = 1;
		pen_blade_rot[i]     = 30000;      /* 30.000 deg swivel threshold */
		pen_FE[i]            = fe;
		pen_FF_a[i]          = 2;
		pen_FF_b[i]          = 2;
	}
	fk_flag = 0;
	cut_force = eeprom_read_byte(0x30003b);
	media_type = eeprom_read_byte(0x30003a);
	if ((signed char)media_type < 1 || media_type > 13)
		media_type = 13;
	track_enhance_off = eeprom_read_byte(0x30003c);
}

/* cutter-offset working state, 0xffb288.. */
extern long cutoff_radius, cutoff_min_angle, cutoff_angle_prev, cutoff_angle;
extern long cutoff_lead_x, cutoff_lead_y;              /* 0xffb298/9c */
extern long cutoff_vec_x, cutoff_vec_y;                /* 0xffb2a0/a4 */
extern int  overcut_a;                    /* 0xffb2b8/ba */
extern u16  cutoff_speed;                              /* 0xffb2b0 */
extern u8 cutoff_flag_c1, cutoff_mode_hi;

/* 0x1054c: load the current pen's parameters into the working set */
long select_pen_params(void)
{
	int p = cur_pen - 1;

	if (p < 0)
		p = 0;
	cutoff_angle = cutoff_angle_prev = 0;
	cutoff_lead_x = cutoff_lead_y = 0;
	pen_params_dirty = 1;
	cutoff_flag_c1 = 0;
	/* 0x17992 */
	cutoff_mode   = (u8)pen_FE[p];                 /* low byte of FE entry */
	cutoff_radius = (long)pen_cutter_offset[p] * 5;
	if (cutoff_radius > 240)
		cutoff_radius = 240;                   /* 1.2 mm max */
	cutoff_speed     = pen_speed_code((u8)pen_cutter_q[p]);
	cutoff_min_angle = pen_blade_rot[p];
	overcut_a   = pen_FF_a[p] * 20;
	overcut_len = pen_FF_b[p] * 20;
	cutoff_mode_hi = 0;
	overcut_pending = 0;
	blade_init_active = 0;
	return 0;
}

/* =========================================================================
 * Data-dump mode (service command 'o')
 * ========================================================================= */
extern char  dump_term1, dump_term2;                   /* 0xffb16e/6f */
extern long  dump_x, dump_y;                           /* 0xffb170/74 */
/* dump_active: 0xffb178 */
extern char  dump_text[4];                             /* 0xffb179 */
extern long  font_width, font_height, font_line_pitch, font_char_pitch; /* 0xffb184.. */
extern const char *dump_reply_tab[18];                 /* 0xffc230, from ROM 0x29644 */
extern void  font_defaults(void);                      /* 0xadae */
/* font_print: 0xaed6 */
extern void  tool_up(void);                            /* 0x17f1e */

/* 0xa5de: canned reply; first character of the table string is a dummy */
void dump_reply(int idx)
{
	const char *s = dump_reply_tab[idx];

	if (*s == '\0')
		return;
	reply_begin(1);
	while (*++s)
		reply_putc(*s);
	reply_putc(dump_term1);
	if (dump_term1 != dump_term2)
		reply_putc(dump_term2);
	reply_send();
}

/* 0xab7a: format a byte as two hex digits and a space */
void dump_hex_byte(u8 c)
{
	static const char hex[] = "0123456789ABCDEF";

	dump_text[0] = hex[c >> 4];
	dump_text[1] = hex[c & 15];
	dump_text[2] = ' ';
	dump_text[3] = 0;
}

/* 0xac58: returns 1 when the byte is shown as hex */
int dump_format_char(u8 c)
{
	if ((c & 0x7f) < 0x20 || c == 0x7f ||
	    (kana_enabled && c >= 0xa0) || (!kana_enabled && c >= 0xe0)) {
		dump_hex_byte(c);
		return 1;
	}
	dump_text[0] = c;
	dump_text[1] = 0;
	return 0;
}

/* 0xaad4: advance the print head along the line, wrap after 180 mm */
void dump_advance(int wide)
{
	dump_y += font_char_pitch;
	if (wide)
		dump_y += font_char_pitch;
	if (dump_y > 36000) {
		dump_y = 0;
		dump_x += font_line_pitch;
	}
	plot_to(dump_x, dump_y, 0);
	motion_idle();
	hook_nop();
}

/* 0xacd6: print one received byte; hex bytes at 2/3 size, half pitch */
void dump_print_char(u8 c)
{
	long w = font_width, h = font_height, pitch = font_char_pitch;
	int hex = dump_format_char(c);

	if (hex) {
		font_width  = w * 2 / 3;
		font_height = h * 2 / 3;
		font_char_pitch = pitch / 2;
	}
	font_print(dump_text);
	font_width = w;
	font_height = h;
	font_char_pitch = pitch;
	dump_advance(hex);
}

/* 0xa5b8: every byte the dump loop reads is also printed */
unsigned int dump_getc(void)
{
	unsigned int c = rx_getc();

	dump_print_char((u8)c);
	return c & 0xff;
}

/* 0xa676 */
void dump_set_term(void)
{
	dump_term1 = (char)dump_getc();
	dump_term2 = (char)dump_getc();
}

/* 0xa696: skip label text up to a terminator */
void dump_skip_label(void)
{
	char c;

	do
		c = (char)dump_getc();
	while (c != dump_term1 && c != dump_term2);
}

/* 0xa6d2: '(' followed by 0x80 ... 0x80: skip a framed block */
void dump_skip_frame(char c)
{
	if ((u8)c != 0x80)
		return;
	while ((u8)dump_getc() != 0x80)
		dump_getc();
	dump_getc();
}

/* 0xa6fc: T... -> TB98 / TB71 / TB22 / TB23 */
void dump_cmd_T(void)
{
	char c;

	if (dump_getc() != 'B')
		return;
	do
		c = (char)dump_getc();
	while (c == ' ');
	if (c == '9') {
		if (dump_getc() == '8')
			dump_reply(8);
	} else if (c == '7') {
		if (dump_getc() == '1')
			dump_reply(10);
	} else if (c == '2') {
		c = (char)dump_getc();
		if (c == '2' || c == '3')
			dump_reply(9);
	}
}

/* 0xa792: F... -> FA / FG / FL / FQn */
void dump_cmd_F(void)
{
	char c = (char)dump_getc();
	const char *s;

	switch (c) {
	case 'A':
		dump_reply(7);
		break;
	case 'G':
		reply_begin(1);
		for (s = fw_version_string; *s; s++)
			reply_putc(*s);
		reply_putc(dump_term1);
		if (dump_term1 != dump_term2)
			reply_putc(dump_term2);
		reply_send();
		break;
	case 'L':
		dump_reply(11);
		break;
	case 'Q':
		c = (char)dump_getc();
		if (c >= '0' && c <= '5')
			dump_reply(12 + (c - '0'));
		break;
	}
}

/* 0xa8a6: never returns */
void dump_loop(void)
{
	u8 c;

	dump_term1 = dump_term2 = 3;                   /* ETX */
	for (;;) {
		c = (u8)dump_getc();
		if (c < 0x21 || (c > 0x2a && c < 0x2f) || (c > 0x2f && c < 0x3a) || c > 0x5f)
			continue;
		switch (c) {
		case 'G': dump_reply(0); break;
		case 'C': dump_reply(1); break;
		case '?': dump_reply(2); break;
		case '[': dump_reply(3); break;
		case 'U': dump_reply(4); break;
		case '@': dump_reply(5); break;
		case '=': dump_set_term(); break;
		case 'P':
		case 'K': dump_skip_label(); break;
		case '(': dump_skip_frame((char)c); break; /* passes '(' itself: never matches 0x80 (firmware bug) */
		case 'T': dump_cmd_T(); break;
		case 'F': dump_cmd_F(); break;
		case 'B':
			if (rx_peek() == 'S') {
				dump_getc();
				dump_reply(6);
			}
			break;
		}
	}
}

/* 0xaab2: service command 'o' */
void svc_o_dump_mode(void)
{
	dump_active = 1;
	font_defaults();
	/* 0xaa9a: dump_start */
	font_defaults();
	dump_x = dump_y = 2000;                        /* 10 mm, 10 mm */
	plot_to(dump_x, dump_y, 0);
	/* 0xaa14: dump_banner */
	plot_to(dump_x, dump_y, 0);
	font_print(fw_version_string);
	dump_y = 0;
	dump_x += font_line_pitch;
	tool_up_wait();
	plot_to(dump_x, dump_y, 0);
	motion_idle();
	tool_up();
	dump_loop();
}

/* =========================================================================
 * Stroke-font byte-code interpreter (0xb96a), structure only
 * ========================================================================= */
extern int  font_sp;                                   /* 0xffb1d4 */
extern const u8 *font_stack[];                         /* 0xffb1d6 */
extern const u8 *const font_subglyphs[];               /* ROM 0x285a0 */
extern int  font_pen;                                  /* 0xffb1be */
extern void font_unpack_xy(unsigned int b, long unused, long *out); /* 0xc450 */
extern void font_draw_to(long *pt);                    /* 0xc4d8 */
extern void font_op_arc(void);                         /* 0xbbe2, ops F0-F4 */
extern void font_op_curve(int op);                     /* 0xbcae/bd38/bdf0/be78, ops FA-FD */
extern void font_op_ef(void);                          /* 0xbb70 */
extern int  font_op_return(void);                      /* 0xbf6e */

void font_interp(const u8 *prog)
{
	long pt[2];
	u8 op;
	int more;

	font_stack[font_sp] = prog;
	while (font_sp >= 0) {
		/* TODO: the original keeps a pending-arc counter (local_18) and calls
		 * font_continue (0xc24c) instead of fetching while it is non-zero. */
		more = 1;
		while (more) {
			op = *font_stack[font_sp]++;
			switch (op) {
			case 0xee: font_pen = -1; break;             /* next point pen up */
			case 0xef: font_op_ef(); break;
			case 0xf0: case 0xf1: case 0xf2: case 0xf3: case 0xf4:
				font_op_arc(); more = 0; break;
			case 0xf5: case 0xf6: case 0xf7: case 0xf8: case 0xf9:
				/* push sub-glyph: the operand selects from font_subglyphs */
				font_sp++;
				font_stack[font_sp] = font_subglyphs[op - 0xf5]; /* TODO: operand decoding */
				break;
			case 0xfa: case 0xfb: case 0xfc: case 0xfd:
				font_op_curve(op); more = 0; break;
			case 0xfe: font_pen = 0; break;              /* pen down */
			case 0xff: more = font_op_return(); break;   /* pop / end */
			default:
				font_unpack_xy(op, 0, pt);           /* hi nibble x, lo nibble y */
				font_draw_to(pt);
				more = 0;
			}
		}
	}
}

/* =========================================================================
 * USB control requests (0xfb6c, 0xfe12, 0xfdf6)
 * ========================================================================= */
extern u8  usb_setup_pkt[8];                           /* 0xffb000 */
/* usb_ctl_end: 0xffc4a2/a6/ae */
/* usb_ep0_state: 0xffc43e: 1 = IN data, 2 = OUT/none */
extern const u8 usb_dev_desc[18];                      /* ROM 0x29166 */
extern const u8 usb_cfg_desc[32];                      /* ROM 0x29178 */
extern const struct { u16 len; const u8 *p; } usb_str_desc[]; /* ROM 0x29142 */
extern u8 *ieee1284_id;  extern u16 ieee1284_len;      /* 0xffc2de / 0xffc2dc */
extern u8  usb_port_status;                            /* 0xffc2e2 */
/* usb_tx_remaining: 0xffb2ec */
extern volatile u8 USB_EP0_CFG;                        /* 0xc00086 */
extern unsigned int usb_wlength(void);                 /* 0x22986 */
extern void usb_stall(void);                           /* 0x1f6be */

static void usb_reply(const u8 *p, unsigned int len)
{
	usb_ctl_buf = (u8 *)p;
	usb_ctl_ptr = usb_ctl_buf + usb_wlength();     /* clip to wLength */
	usb_ctl_end = usb_ctl_buf + len;
	if (usb_ctl_end < usb_ctl_ptr)
		usb_ctl_ptr = usb_ctl_end;
}

/* 0xfe12: printer class requests */
void usb_class_request(void)
{
	switch (usb_setup_pkt[1]) {
	case 0:                                        /* GET_DEVICE_ID */
		usb_reply(ieee1284_id, ieee1284_len);  /* TODO: end uses the id length, as in the original */
		break;
	case 1:                                        /* GET_PORT_STATUS: selected, no error */
		usb_port_status = 0x18;
		usb_reply(&usb_port_status, ieee1284_len); /* sic: end bound uses the id length */
		break;
	case 2:                                        /* SOFT_RESET */
		usb_ctl_buf = usb_ctl_ptr = &usb_setup_pkt[8];
		usb_tx_remaining = 0;
		USB_EP0_CFG = 0x30;
		break;
	default:
		usb_stall();
	}
}

/* 0xfdf6: vendor request 0 sets the jog keys from wValue */
char usb_vendor_request(void)
{
	if (usb_setup_pkt[1] == 0)
		jog_keys = usb_setup_pkt[2];
	return (char)usb_setup_pkt[1];
}

/* 0xfb6c */
int usb_std_request(void)
{
	switch (usb_setup_pkt[0] & 0x60) {
	case 0x00:
		if (usb_setup_pkt[1] == 6) {           /* GET_DESCRIPTOR */
			if (usb_setup_pkt[3] == 1)
				usb_reply(usb_dev_desc, sizeof usb_dev_desc);
			else if (usb_setup_pkt[3] == 3)
				usb_reply(usb_str_desc[usb_setup_pkt[2]].p, usb_str_desc[usb_setup_pkt[2]].len);
			else
				usb_reply(usb_cfg_desc, sizeof usb_cfg_desc);
		} else if (usb_setup_pkt[1] == 7) {    /* SET_DESCRIPTOR: receive into 0xffb008 */
			usb_ctl_ptr = usb_ctl_buf = &usb_setup_pkt[8];
			/* 0xffc4c2/c6 = descriptor area for the data stage */
		}
		break;
	case 0x20: usb_class_request(); break;
	case 0x40: usb_vendor_request(); break;
	default:   usb_stall(); break;
	}
	usb_ep0_state = (usb_setup_pkt[0] & 0x80) ? 1 : 2;
	return usb_ep0_state;
}

/* =========================================================================
 * Misc
 * ========================================================================= */

/* 0x7636: the '=' (terminator) handler -- present but unreferenced in 2.30 */
void cmd_term_unused(void)
{
	term1 = (char)rx_getc();
	term2 = (char)rx_getc();
	term_changed_hook(term1);
}

/* 0x101fe: service 'e a,b,c,d' -> "%10d,%10d,%10d" */
void svc_e_calc(void)
{
	long a = read_int(), b = read_int(), c = read_int(), d = read_int();

	reply_begin(1);
	reply_number_comma(10, cmp64(a, b, c, d));  /* 64-bit compare (a:b) vs (c:d) */
	reply_number_comma(10, 0);
	reply_number_end(10, 0);
}

/*
 * Summary of routines not written out above (see notes/subsys_E.md):
 *
 *  0x46b8 regmark_find_corners      double line intersections of detected mark edges
 *  0x48ba regmark_line_intersect    general line/line intersection (doubles)
 *  0x5082 regmark_axis_intersect    intersection with an axis-parallel line
 *  0x54e4 regmark_compute_transform mark1/mark2(/mark3) -> direction cosines, scale, origin
 *  0x5bc0 regmark_scale_x / 0x5d40 regmark_scale_y   nominal vs. measured distance, 2.14 scale
 *  0x5ecc regmark_extend_segment    lengthen a segment by 3000/len at both ends
 *  0x6140 clip_segment_hits         segment vs. clip polygon edges, up to 30 hits
 *  0x6950 clip_edge_local           edge into segment-local frame (swap if steep)
 *  0x739c hw_init_bus_ports         bus controller / port register setup
 *  0x7984 set_area_extent           0xffb51c/20 = media extent - origin, then resync
 *  0x834c circle3_draw              WP circumcentre (doubles) -> arc_draw
 *  0x9824 arc_draw_float            unreferenced single-float arc renderer
 *  0xa12a fx_mul16                  unreferenced 16.16 multiply
 *  0xadae..0xc994                   stroke font engine (metrics, lookup, primitives)
 *  0xf096 distcorr_compute, 0xf330/0xf4b2/0xf61e  FA/FB distance-correction scaling
 *  0xf78a..0xf9f8, 0xfa9c           USB EP0/EP-IN low-level stages (controller at 0xc00000)
 *  0x102ae plot_soft_reset          A-reset, 0x13a8c, terminators (opt.), clip_reset
 *  0x10686..0x10fd8                 drag-knife offset compensation (see notes)
 *  0xd448 hook_pen_param = hook_nop(); 0xd458 hook_sync = hook_nop() + wait_motion
 *  0xffbc hook_nop, 0xffc8 hook_B, 0x100ca term_changed_hook        empty
 *  0xffd4 DP/MP, 0xffe6 EP/OP, 0x10004 RP, 0x1002e (P, 0x10046 N,
 *  0x1006a %, 0x100be (                                              empty command handlers
 */
