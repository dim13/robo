/*
 * CC100 firmware 2.30 -- GP-GL plotting / geometry commands.
 * Hand-cleaned reconstruction of the Ghidra decompilation. Comments give the ROM address.
 * Target types: int = 16 bit, long = 32 bit. Float/double ops call the compiler runtime
 * (0x22a66 dsub, 0x22a8a dadd, 0x234a0 dmul, 0x22eee ddiv, 0x23358 long->double,
 * 0x23174 double->long, 0x23202 float->double, 0x16700 sqrt); written here as plain C math.
 */

#include "cc100.h"
#include <math.h>
#include <stdlib.h>

/* ---- state (RAM) ---- */
int  steps_per_unit;            /* 0xffb262 : 20/10/5/2, table 0x28d28 */
int  num_frac1, num_frac2;  /* 0xffb360/2 : first two fraction digits of last number */
long cur_x, cur_y;          /* 0xffb584/8 : current position, user units */
long dev_x, dev_y;          /* 0xffb53c/40: position after factor/rotation/orientation */
int  line_type;             /* 0xffbc84 */
long line_pitch;            /* 0xffbc72 */
int  line_type_arg;         /* 0xffb3a0 : last L argument */
long line_pitch_arg;        /* 0xffb39c : last B argument */
int  user_lt_n;             /* 0xffbc86 */
long user_lt_a, user_lt_b, user_lt_c;   /* 0xffbc8a/8e/92 */
unsigned long user_lt[6];   /* 0xffbc96 : user pattern, bit31 = pen down */
long factor_p = 1, factor_q = 1, factor_r = 1;  /* 0xffb4d4/d8/dc */
int  factor_unity;      /* 0xffb5a4 : bit0 x factor == 1, bit1 y factor == 1 */
long offset_x, offset_y;    /* 0xffb4e0/4 */
long rot_cx, rot_cy;        /* 0xffb504/8 */
int  rot_enabled;           /* 0xffb120 */
float rot_sin, rot_cos = 1.0f;          /* 0xffb5b2/6 */
long win_ll_x, win_ll_y, win_ur_x, win_ur_y;          /* 0xffb564..0xffb570 */
long clip_xmin, clip_ymin, clip_xmax, clip_ymax;      /* 0xffb554..0xffb560 */
long hw_max_x, hw_max_y;    /* 0xffb574/8 */
int  orient_flags;      /* 0xffb5c6 */
char ur_set;                /* 0xffb3b9 */
char landscape;             /* 0xffb400  axes swapped; also selects which clip edge is unbounded */
int  axis_type;             /* 0xffb3a2 */
long feed_total;            /* 0xffb3b0 : accumulated chart feed */
long *clip_sp, *clip_limit; /* 0xffb2c4 / 0xffbc78 : clip polygon stack (grows down) */


/* ===================================================================== */
/* Number parsing                                                        */
/* ===================================================================== */

/*
 * read_number (0x11a36), read_int (0x11b72), read_coord (0x11bc6),
 * read_point (0x13ab2) and end_of_params (0x0cb3a) are in input.c.
 */

/* 0x11d50: angle in 1/100 degree */
long read_angle(void)
{
    long v = read_number() * 100;
    int f = num_frac1 * 10 + num_frac2;
    return v + (v < 0 ? -f : f);
}

/* 0x0806a: optional trailing angle, 0 if absent */
long read_opt_angle(void)
{
    return end_of_params() ? 0 : read_angle();
}

/* 0x0d412: ZP, FR -- discard parameters up to the next command character */
void skip_command(void)
{
    for (;;) {
        int k = cmd_class(rx_peek());
        if (k >= 0 && k <= 0x30)
            return;
        rx_getc();
    }
}

/* ===================================================================== */
/* Plot pipeline                                                         */
/* ===================================================================== */

/* 0x16130: a*b/c with 64-bit intermediate (TODO: rounding mode) */
long muldiv(long a, long b, long c);

/* 0x1e7a8: Cohen-Sutherland outcode */
int window_outcode(const long p[2])
{
    int code = p[0] > clip_xmax ? 1 : p[0] < clip_xmin ? 2 : 0;

    if (p[1] > clip_ymax)
        code |= 4;
    else if (p[1] < clip_ymin)
        code |= 8;
    if (!ur_set)
        code &= landscape ? ~4 : ~1;  /* roll axis unbounded until Z is given */
    return code;
}

/* 0x1e86a */
void rotate_point(long *x, long *y)
{
    double dx = *x - rot_cx, dy = *y - rot_cy;
    *x = rot_cx + (long)(dx * rot_cos - dy * rot_sin);
    *y = rot_cy + (long)(dx * rot_sin + dy * rot_cos);
    /* TODO: verify sign convention of the second term in 0x1e86a */
}

/* 0x1e23e: user -> device transform, window clip, emit */
void plot_transform(long x, long y, int pen)
{
    long p[2], prev[2] = { dev_x, dev_y };
    int c0, c1;

    if (!(factor_unity & 1))
        x = muldiv(x, factor_p, factor_r);
    if (!(factor_unity & 2))
        y = muldiv(y, factor_q, factor_r);
    if (rot_enabled)
        rotate_point(&x, &y);
    if (orient_flags) {
        if (orient_flags != 0x60) {
            long t = x;
            x = y;
            y = t;
        }
        if (!(orient_flags & 0x40))
            x = hw_max_x - x;
        if (!(orient_flags & 0x20))
            y = hw_max_y - y;
    }
    p[0] = dev_x = x;
    p[1] = dev_y = y;

    c0 = window_outcode(prev);
    c1 = window_outcode(p);
    /* Cohen-Sutherland: trivially reject when both share an outside side; otherwise clip the
     * start point onto the window (0x1e9a6/0x1e9fc/0x1ea52/0x1eaa8 = intersect with
     * x=xmax/x=xmin/y=ymax/y=ymin), move there pen-up, then clip the end point the same way.
     * ponytail: intersection helpers summarized, see out.c for the exact arithmetic. */
    if (c0 & c1)
        return;
    /* ... clipping elided ... */
    plot_device(p[0], p[1], pen);
}

/* 0x1eb88: split against clip polygons ('>'), alternating pen at each crossing */
void plot_clip_poly(long x, long y, int pen);   /* see out.c; calls plot_transform per piece */

/* 0x1df5e: line-type (dash) expansion, then plot_clip_poly */
void plot_linetype(long x, long y, int pen);   /* see notes/plot.md "Coordinate pipeline" */

/* 0x1db80 */
void plot_to(long x, long y, int pen)
{
    plot_linetype(x, y, pen);
}

/* 0x13ada / 0x13b16 */
void plot_abs(int pen)
{
    long x, y;
    read_point(&x, &y);
    plot_to(x, y, pen);
}

void plot_rel(int pen)
{
    long x, y;
    read_point(&x, &y);
    plot_to(cur_x + x, cur_y + y, pen);
}

/* ===================================================================== */
/* Line commands                                                         */
/* ===================================================================== */

void cmd_D(void) { while (!end_of_params()) plot_abs(1); }   /* 0x13bca */
void cmd_E(void) { while (!end_of_params()) plot_rel(1); }   /* 0x13be8 */
void cmd_M(void) { plot_abs(0); }                             /* 0x13d3e */
void cmd_O(void) { plot_rel(0); }                             /* 0x13d50 */
void cmd_J(void) { read_int(); }                              /* 0x13c58: pen number ignored */

/* 0x0ffd4 DP/MP, 0x0ffe6 EP/OP, 0x10004 RP, 0x1006a '%', 0x100be '(', 0x1002e '(P',
 * 0x10046 'N', 0x0ffc8 BS: empty functions. Their parameters fall through to the main
 * loop, which ignores numeric characters. */

/* ===================================================================== */
/* Line type / scale                                                     */
/* ===================================================================== */

extern void dash_restart(void);     /* 0x1dc18 */

/* 0x1dbc0 */
void set_line_type(int type, long pitch)
{
    if (type == 100)
        return;
    if (type >= -9 && type <= 9) {
        line_type = type;
        if (pitch != 0)
            line_pitch = pitch < 0 ? -pitch : pitch;
    }
    dash_restart();
}

/* 0x2032e: user line type 100 */
void set_user_line_type(int n)
{
    user_lt[0] = user_lt_a | 0x80000000;   /* pen down a */
    user_lt[1] = user_lt_b;                /* pen up b   */
    line_pitch = user_lt_a + user_lt_b;
    if (n != 1) {
        user_lt[2] = user_lt_c | 0x80000000;
        user_lt[3] = user_lt_b;
        line_pitch += user_lt_c + user_lt_b;
        if (n != 2) {
            user_lt[4] = user_lt_c | 0x80000000;
            user_lt[5] = user_lt_b;
            line_pitch += user_lt_c + user_lt_b;
        }
    }
    dash_restart();
}

/* 0x13c68: L p   |   L 100,n,a,b[,c] */
void cmd_L(void)
{
    long p = read_int();
    long v;

    if (p != 100) {
        line_type_arg = (int)p;
        set_line_type(line_type_arg, line_pitch_arg);
        return;
    }
    user_lt_n = (int)read_int();
    if (user_lt_n > 3 || user_lt_n <= 0)
        return;
    if ((v = read_coord()) <= 0)
        return;
    user_lt_a = v;
    if ((v = read_coord()) <= 0)
        return;
    user_lt_b = v;
    if (user_lt_n > 1) {
        if ((v = read_coord()) <= 0)
            return;
        user_lt_c = v;
    }
    line_type = line_type_arg = 100;
    set_user_line_type(user_lt_n);
}

extern void cmd_BZ(void);   /* 0x71d6, below */

/* 0x13b7e: B l | BS | BZ ... */
void cmd_B(void)
{
    if (rx_peek() == 'S') {
        rx_getc();            /* BS: 0x0ffc8, no-op */
        return;
    }
    if (rx_peek() == 'Z') {
        rx_getc();
        cmd_BZ();
        return;
    }
    line_pitch_arg = read_coord();
    set_line_type(line_type_arg, line_pitch_arg);
}

/* ===================================================================== */
/* Transform commands                                                    */
/* ===================================================================== */

extern void sync_user_pos(void);    /* 0x18e44: cur = inverse_transform(dev) */
extern void update_window(void);    /* 0x19444 */
extern void angle_to_sincos(long centideg, float sc[2]);   /* 0x100e2 */

/* 0x075e2: / x,y,theta */
void cmd_rotate(void)
{
    float sc[2];
    long a;

    read_point(&rot_cx, &rot_cy);
    a = read_angle();
    rot_enabled = a != 0;
    angle_to_sincos(a, sc);
    rot_sin = sc[0];
    rot_cos = sc[1];
    sync_user_pos();
}

/* 0x07522: & p,q,r */
void set_factor(long p, long q, long r);

void cmd_factor(void)
{
    long p = read_int(), q = read_int(), r = read_int();
    set_factor(p, q, r);
}

/* 0x1919a */
void set_factor(long p, long q, long r)
{
    if (!p || !q || !r)
        return;
    if (labs(p / r) >= 9 || labs(q / r) >= 9)
        return;
    if (labs(r / p) > 4096 || labs(r / q) > 4096)
        return;
    factor_p = p;
    factor_q = q;
    factor_r = r;
    /* factor_unity: bit pattern 0x5555 when x factor is exactly 1, |= 0xaaaa when y is.
     * TODO: the firmware computes this twice with identical code. */
    factor_unity = 0;
    if ((double)p / r == 1.0)
        factor_unity = 0x5555;
    if ((double)q / r == 1.0)
        factor_unity |= 0xaaaa;
    sync_user_pos();
}

/* 0x17838: ^ x,y  -- keeps physical pen position */
void cmd_offset(void)
{
    long x, y;

    read_point(&x, &y);
    dev_x += offset_x - x;
    dev_y += offset_y - y;
    offset_x = x;
    offset_y = y;
    sync_user_pos();
}

/* 0x19444 */
void update_window(void)
{
    orient_flags = 0;
    if (win_ll_x < win_ur_x) {
        clip_xmin = win_ll_x;
        clip_xmax = win_ur_x;
    } else {
        clip_xmin = win_ur_x;
        clip_xmax = win_ll_x;
        orient_flags |= 0x40;
    }
    if (win_ll_y < win_ur_y) {
        clip_ymin = win_ll_y;
        clip_ymax = win_ur_y;
    } else {
        clip_ymin = win_ur_y;
        clip_ymax = win_ll_y;
        orient_flags |= 0x40;   /* sic: firmware sets bit 6 here too */
    }
    sync_user_pos();
}

/* 0x177da: \ x,y */
void cmd_lower_left(void)
{
    read_point(&win_ll_x, &win_ll_y);
    update_window();
}

/* 0x1777c: Z x,y */
void cmd_upper_right(void)
{
    read_point(&win_ur_x, &win_ur_y);
    update_window();
    ur_set = 1;
}

/* 0x13c22: H */
void cmd_H(void)
{
    motion_flush();
    motion_line_wait(-feed_total, 0, 0);     /* undo chart feed */
    wait_motion_done();
    sync_device_pos();
}

/* ===================================================================== */
/* Clipping polygon '>'                                                  */
/* ===================================================================== */

extern long *clip_base;     /* 0xffb0f0 */
extern long *clip_heap;     /* 0xffb0ec */

/* 0x110ba */
void clip_reset(void)
{
    clip_sp = clip_base - 1;
    *clip_sp-- = 0;                 /* polygon count */
    clip_limit = clip_heap;         /* 0x11102 */
    *clip_limit = 0;
}

/* 0x1112c: > x1,y1,...  (no parameters: clear) */
void cmd_clip(void)
{
    long *count, x, y;

    if (end_of_params()) {
        clip_reset();
        return;
    }
    clip_limit = clip_heap;         /* 0x11102 */
    *clip_limit = 0;
    clip_base[-1]++;                /* one more polygon */
    count = clip_sp;
    *count = 0;
    clip_sp--;
    while (!end_of_params() && clip_sp > clip_limit) {
        read_point(&x, &y);
        (*count)++;
        *clip_sp-- = x;
        *clip_sp-- = y;
    }
    clip_sp--;
}

/* ===================================================================== */
/* Axis 'X'                                                              */
/* ===================================================================== */

extern void plot_solid(long x, long y, int pen);   /* 0x1eafe */

/* 0x13d62: relative pen-down step along/across the axis */
void axis_draw_rel(long along, long across)
{
    if (axis_type == 0 || axis_type == 2) {         /* Y axis: swap */
        long t = along;
        along = across;
        across = t;
    }
    plot_solid(cur_x + along, cur_y + across, 1);
}

/* 0x13eb4: X p,q,r[,t1,t2] */
void cmd_X(void)
{
    long x0 = cur_x, y0 = cur_y;
    long q, t1 = 200, t2 = 200;
    int r, n;

    axis_type = (int)read_int();
    q = read_coord();
    r = n = (int)read_int();
    if (r == 0)
        return;
    if (!end_of_params())
        t1 = read_coord();
    if (!end_of_params())
        t2 = read_coord();
    for (;;) {
        axis_draw_rel(0, t1);                 /* tick */
        axis_draw_rel(0, -t1 - t2);
        axis_draw_rel(0, t2);
        if (--n < 0)
            break;
        if (axis_type >= 2) {                 /* 0x13dc2: exact position x0 + q*(r-n) */
            double s = (double)q * (r - n);     /* TODO: 0x13dc2 computes this in double from q, n, r */
            long ax = x0, ay = y0;
            if (axis_type == 2)
                ay += (long)s;
            else
                ax += (long)s;
            plot_solid(ax, ay, 1);
        } else
            axis_draw_rel(q, 0);
    }
}

/* ===================================================================== */
/* Circles, curves                                                       */
/* ===================================================================== */

struct arc {            /* stack frame layout used by 0x8e22 / 0x9070 */
    long x0, y0;        /* centre */
    long r1, r2;        /* start / end radius (spiral if different) */
    double th1, th2;    /* centidegrees */
    double pad;         /* zeroed */
    double d;           /* chord step (centideg), 0 = auto, <0 = -segments */
    int pen0;           /* pen for first point: 0, or 1 for ']' (0xffff once drawing) */
    int flag;
};

/* 0x8e22: returns 0 if arc degenerated to a dot */
static int arc_prepare(struct arc *a)
{
    long rmax = labs(a->r1) > labs(a->r2) ? labs(a->r1) : labs(a->r2);
    double span = a->th2 - a->th1;

    if (rmax < 10) {
        plot_to(a->x0, a->y0, 0);
        plot_to(a->x0, a->y0, 1);
        return 0;
    }
    if (a->d == 0) {
        a->d = 315000.0 / sqrt((double)rmax);
        if (a->d > 5000.0)
            a->d = 5000.0;
    } else if (a->d < 0) {
        a->d = span * 100.0 / a->d;   /* -n -> n segments (d was read x100) */
        if (a->d < 0)
            a->d = -a->d;
        if (a->d == 0)
            a->d = 1.0;
    }
    if ((long)span * (long)a->d < 0)     /* step takes sign of span */
        a->d = -a->d;
    return 1;
}

/* 0x9070 */
static void arc_draw(struct arc *a)
{
    double span;
    long i, n;
    float sc[2];

    if (!arc_prepare(a))
        return;
    span = a->th2 - a->th1;
    n = (long)ceil(span / a->d);    /* TODO: rounding of 0x23b3c */
    if (a->pen0)
        a->pen0 = -1;
    for (i = 0; i <= n; i++) {
        double th = (i == n) ? a->th2 : a->th1 + i * a->d;
        double r = span != 0 ? (a->r1 - a->r2) * (a->th2 - th) / span + a->r2 : a->r1;
        angle_to_sincos((long)th, sc);
        plot_to(a->x0 + (long)(sc[1] * r), a->y0 + (long)(sc[0] * r), a->pen0);
        a->pen0 = -1;   /* pen down after first point */
    }
}

extern char settings_368;    /* 0xffb368 */
extern int  arc_pen_2b6;     /* 0xffb2b6 */

/* 0x08c36: W x0,y0,r1,r2,t1,t2[,d]   |  WP ... */
void cmd_W(void)
{
    struct arc a = { 0 };

    if (rx_peek() == 'P') {
        rx_getc();
        extern void cmd_WP(void);
        cmd_WP();
        return;
    }
    read_point(&a.x0, &a.y0);
    read_point(&a.r1, &a.r2);
    a.th1 = read_angle();
    a.th2 = read_angle();
    a.d = read_opt_angle();
    a.pen0 = 0;
    a.flag = 1;
    if ((settings_368 & 0x40) && cutoff_radius)
        a.pen0 = arc_pen_2b6;
    arc_draw(&a);
}

/* 0x0828e: WP a,x1,y1,x2,y2,x3,y3[,d] */
void cmd_WP(void)
{
    long x1, y1, x2, y2, x3, y3;
    int a = (int)read_int();

    if (a < 0)
        a = 0;
    read_point(&x1, &y1);
    read_point(&x2, &y2);
    read_point(&x3, &y3);
    circle3_draw(x1, y1, x2, y2, x3, y3, a, (double)read_opt_angle());
}

/* 0x08d14: ] r1,r2,t1,t2[,d] -- centre chosen so the current point is on the arc at t1 */
void cmd_rel_circle(void)
{
    struct arc a = { 0 };
    float sc[2];

    read_point(&a.r1, &a.r2);
    a.th1 = read_angle();
    a.th2 = read_angle();
    a.d = read_opt_angle();
    a.pen0 = 1;
    a.flag = 1;
    angle_to_sincos((long)a.th1, sc);
    a.x0 = cur_x - (long)(a.r1 * sc[1]);
    a.y0 = cur_y - (long)(a.r1 * sc[0]);
    arc_draw(&a);
}

/* 0x081f8: ) a,x0,y0,r1,r2,t1,t2,t3[,d] */
void cmd_ellipse(void)
{
    struct {
        long x0, y0, r1, r2;
        double th1, th2, th3, d;
        int a, flag;
    } e;

    e.a = (int)read_int();
    if (e.a < 0)
        e.a = 0;
    read_point(&e.x0, &e.y0);
    read_point(&e.r1, &e.r2);
    e.th1 = read_angle();
    e.th2 = read_angle();
    e.th3 = read_angle();       /* axis rotation */
    e.d = read_opt_angle();
    e.flag = 1;
    ellipse_draw(&e);
}

/* 0x0808c: shared Y / _ body. relative: each point offsets the previous one. */
void curve_cmd(int relative, long ox, long oy)
{
    int closed = read_int() != 0 ? 4 : 0;
    int state = 0, moved = 0;
    long x, y;

    spline_feed(closed, 0, 0);                  /* begin: 0 open, 4 closed */
    while (!end_of_params()) {
        read_point(&x, &y);
        ox += x;
        oy += y;
        if (state == 0 && ox == cur_x && oy == cur_y) {
            spline_feed(closed ? closed : 1, ox, oy);   /* start at current point */
            state = 2;
        }
        spline_feed(2, ox, oy);                 /* add point */
        if (!moved && closed != 4) {
            moved = 1;
            if (state != 2)
                plot_solid(ox, oy, 0);          /* pen-up to first point */
        }
        if (!relative)
            ox = oy = 0;
        state = 1;
    }
    spline_feed(3, x, y);                       /* finish */
}

void cmd_Y(void) { curve_cmd(0, 0, 0); }                  /* 0x08cf8 */
void cmd_rel_curve(void) { curve_cmd(1, cur_x, cur_y); }  /* 0x08df8 */

/* 0x06c76: integer distance */
extern long dist(long x1, long y1, long x2, long y2);

/* 0x06cda */
static void bezier_draw(int a, long nseg, const double p[8])
{
    long i;

    if (nseg == 0) {
        long L = dist(p[0], p[1], p[2], p[3]) + dist(p[2], p[3], p[4], p[5]) +
                 dist(p[4], p[5], p[6], p[7]);
        /* TODO: 0x23052 scales L before the thresholds below */
        if (L >= 0x619)
            nseg = 40;
        else if (L < 0x15)
            nseg = 4;
        else
            nseg = (long)sqrt((double)L);   /* TODO: exact formula via 0x23f1e */
    }
    plot_to((long)p[0], (long)p[1], a);
    for (i = 1; i < nseg; i++) {
        double t = (double)i / nseg, u = 1 - t;
        double x = u*u*u*p[0] + 3*u*u*t*p[2] + 3*u*t*t*p[4] + t*t*t*p[6];
        double y = u*u*u*p[1] + 3*u*u*t*p[3] + 3*u*t*t*p[5] + t*t*t*p[7];
        plot_to((long)x, (long)y, 1);
    }
    plot_to((long)p[6], (long)p[7], 1);
}

/* 0x071d6: BZ a,x1,y1,x2,y2,x3,y3,x4,y4[,d] */
void cmd_BZ(void)
{
    long v[8], d;
    double p[8];
    int a = (int)read_int(), i;

    if (a < 0)
        a = 0;
    for (i = 0; i < 8; i += 2)
        read_point(&v[i], &v[i + 1]);
    d = end_of_params() ? 0 : read_int();
    for (i = 0; i < 8; i++)
        p[i] = v[i];
    bezier_draw(a, d, p);
}
