/*
 * CC100 firmware 2.30 -- shared declarations for the hand-cleaned sources.
 *
 * Generated from ../symbols.txt (the source of truth for names and
 * prototypes; regenerate when it changes) plus a few hand-written
 * additions at the end. Target types: int = 16 bit, long = 32 bit
 * (Renesas H8S, advanced mode). Comments give the RAM/ROM address.
 */
#ifndef CC100_H
#define CC100_H

#define ETX 0x03
#define ESC 0x1b
#define FS  0x1c

typedef struct { long x, y; } pt_t;   /* point in internal units */

/* ---- RAM globals ---- */
extern char fw_version_string[10];                   /* 0x028cec */
extern double rm_yx;                             /* 0xffb06a */
extern double rm_xx;                             /* 0xffb072 */
extern double rm_yy;                             /* 0xffb07a */
extern double rm_xy;                             /* 0xffb082 */
extern double rm_sx;                             /* 0xffb092 */
extern double rm_sy;                             /* 0xffb09a */
extern char regmark_active;                      /* 0xffb0aa */
extern char rm_scale_on;                         /* 0xffb0ab */
extern int rot_enabled;                          /* 0xffb120 */
extern int cut_force;                            /* 0xffb122 */
extern char dump_term1;                          /* 0xffb16e */
extern char dump_term2;                          /* 0xffb16f */
extern long dump_x;                              /* 0xffb170 */
extern long dump_y;                              /* 0xffb174 */
extern char dump_active;                         /* 0xffb178 */
extern long font_width;                          /* 0xffb184 */
extern long font_height;                         /* 0xffb188 */
extern long font_line_pitch;                     /* 0xffb18c */
extern long font_char_pitch;                     /* 0xffb190 */
extern short font_id;                            /* 0xffb1a0 */
extern int last_cmd;                             /* 0xffb22c */
extern char service_mode;                        /* 0xffb22e  lowercase commands enabled (FPGRFCC1) */
extern double distcorr_x;                        /* 0xffb24a */
extern double distcorr_y;                        /* 0xffb252 */
extern int steps_per_unit;                       /* 0xffb262  20/10/5/2 steps per user unit */
extern long err_mask_long;                       /* 0xffb282 */
extern int error_mask;                           /* 0xffb284  low word of the long stored by '"' */
extern unsigned char error_status;               /* 0xffb286  write-only, never reported */
extern long cutoff_radius;                       /* 0xffb288 */
extern long cutoff_min_angle;                    /* 0xffb28c */
extern long cutoff_angle_prev;                   /* 0xffb290 */
extern long cutoff_angle;                        /* 0xffb294 */
extern long cutoff_tip_x;                        /* 0xffb2a8 */
extern long cutoff_tip_y;                        /* 0xffb2ac */
extern long seg_merge_dist;                      /* 0xffb2b2 */
extern short overcut_len;                        /* 0xffb2ba */
extern unsigned char fk_flag;                    /* 0xffb2bc */
extern char cutoff_mode;                         /* 0xffb2bd */
extern unsigned char pen_params_dirty;           /* 0xffb2be */
extern char overcut_pending;                     /* 0xffb2bf */
extern char blade_init_active;                   /* 0xffb2c2 */
extern unsigned char *rx_buf_base;               /* 0xffb2d6  = 0xffc800 */
extern long rx_buf_size;                         /* 0xffb2da */
extern unsigned char *rx_wr_ptr;                 /* 0xffb2de  written by the ISRs */
extern unsigned char *rx_rd_ptr;                 /* 0xffb2e2  parser */
extern volatile long rx_count;                   /* 0xffb2e6 */
extern int tx_len;                               /* 0xffb2ea */
extern short usb_tx_remaining;                   /* 0xffb2ec */
extern short usb_tx_offset;                      /* 0xffb2ee */
extern volatile char tx_busy;                    /* 0xffb357 */
extern char term1;                               /* 0xffb359  always ETX */
extern char term2;                               /* 0xffb35a  always ETX */
extern volatile char rx_starved;                 /* 0xffb35e  1 while the parser waits for data */
extern int num_frac1;                            /* 0xffb360 */
extern int num_frac2;                            /* 0xffb362 */
extern double distcorr_inv_x;                    /* 0xffb36a */
extern double distcorr_inv_y;                    /* 0xffb372 */
extern volatile int delay_counter;               /* 0xffb37c  decremented by timer_tick_isr */
extern volatile char usb_suspended;              /* 0xffb397  really the soft power-off flag */
extern long line_pitch_arg;                      /* 0xffb39c */
extern int line_type_arg;                        /* 0xffb3a0 */
extern int axis_type;                            /* 0xffb3a2 */
extern long chart_feed_len;                      /* 0xffb3aa */
extern char track_enhance_off;                   /* 0xffb3af */
extern long feed_total;                          /* 0xffb3b0 */
extern long motor_y_bias;                        /* 0xffb3b4 */
extern char media_mode;                          /* 0xffb3b8 */
extern char ur_set;                              /* 0xffb3b9 */
extern char landscape;                           /* 0xffb400 */
extern int cur_pen;                              /* 0xffb414 */
extern int pen_speed[8];                            /* 0xffb416 */
extern int pen_force[8];                            /* 0xffb436 */
extern int pen_cutter_offset[8];                    /* 0xffb446 */
extern int pen_cutter_q[8];                         /* 0xffb456 */
extern long pen_blade_rot[8];                       /* 0xffb466 */
extern int pen_FE[8];                               /* 0xffb486 */
extern int pen_FF_a[8];                             /* 0xffb496 */
extern int pen_FF_b[8];                             /* 0xffb4a6 */
extern char media_type;                          /* 0xffb4c8 */
extern long factor_p;                            /* 0xffb4d4 */
extern long factor_q;                            /* 0xffb4d8 */
extern long factor_r;                            /* 0xffb4dc */
extern long offset_x;                            /* 0xffb4e0 */
extern long offset_y;                            /* 0xffb4e4 */
extern long origin_x;                            /* 0xffb4e8 */
extern long origin_y;                            /* 0xffb4ec */
extern long regmark_off_x;                       /* 0xffb4f0 */
extern long regmark_off_y;                       /* 0xffb4f4 */
extern long orient_off_x;                        /* 0xffb4f8 */
extern long orient_off_y;                        /* 0xffb4fc */
extern long rot_cx;                              /* 0xffb504 */
extern long rot_cy;                              /* 0xffb508 */
extern long page_length;                         /* 0xffb50c */
extern long page_width;                          /* 0xffb510 */
extern long mach_xmin;                           /* 0xffb524 */
extern long mach_ymin;                           /* 0xffb528 */
extern long mach_xmax;                           /* 0xffb52c */
extern long mach_ymax;                           /* 0xffb530 */
extern long mach_last_x;                         /* 0xffb534 */
extern long mach_last_y;                         /* 0xffb538 */
extern long dev_x;                               /* 0xffb53c */
extern long dev_y;                               /* 0xffb540 */
extern long clip_xmin;                           /* 0xffb554 */
extern long clip_ymin;                           /* 0xffb558 */
extern long clip_xmax;                           /* 0xffb55c */
extern long clip_ymax;                           /* 0xffb560 */
extern long win_ll_x;                            /* 0xffb564 */
extern long win_ll_y;                            /* 0xffb568 */
extern long win_ur_x;                            /* 0xffb56c */
extern long win_ur_y;                            /* 0xffb570 */
extern long cur_x;                               /* 0xffb584 */
extern long cur_y;                               /* 0xffb588 */
extern int factor_unity;                         /* 0xffb5a4 */
extern float rot_sin;                            /* 0xffb5b2 */
extern float rot_cos;                            /* 0xffb5b6 */
extern int orient_flags;                         /* 0xffb5c6 */
extern int mach_last_outcode;                    /* 0xffb5c8 */
extern int cur_outcode;                          /* 0xffb5ca */
extern int port600010_shadow;                    /* 0xffb5de */
extern int jog_speed_x;                          /* 0xffb5f0 */
extern int jog_speed_y;                          /* 0xffb5f2 */
extern char jog_dir_x;                           /* 0xffbbd6 */
extern char jog_dir_y;                           /* 0xffbbd7 */
extern long rm_area_xmin;                        /* 0xffbc20 */
extern long rm_area_ymin;                        /* 0xffbc24 */
extern long rm_area_xmax;                        /* 0xffbc28 */
extern long rm_area_ymax;                        /* 0xffbc2c */
extern long rm_last_x;                           /* 0xffbc30 */
extern long rm_last_y;                           /* 0xffbc34 */
extern int rm_last_outcode;                      /* 0xffbc38 */
extern long feed_min_x;                          /* 0xffbc3a */
extern long motor_target_x;                      /* 0xffbc3e */
extern long motor_target_y;                      /* 0xffbc42 */
extern int tool_state;                           /* 0xffbc46 */
extern long pending_up_x;                        /* 0xffbc4a */
extern long pending_up_y;                        /* 0xffbc4e */
extern long deferred_x;                          /* 0xffbc54 */
extern long deferred_y;                          /* 0xffbc58 */
extern int deferred_pen;                         /* 0xffbc64 */
extern int tool_force_applied;                   /* 0xffbc66 */
extern char pending_up;                          /* 0xffbc68 */
extern long line_pitch;                          /* 0xffbc72 */
extern int line_type;                            /* 0xffbc84 */
extern int user_lt_n;                            /* 0xffbc86 */
extern long user_lt_a;                           /* 0xffbc8a */
extern long user_lt_b;                           /* 0xffbc8e */
extern long user_lt_c;                           /* 0xffbc92 */
extern long travel_max_x;                        /* 0xffbefa */
extern long travel_max_y;                        /* 0xffbefe */
extern char reply_via_sci;                       /* 0xffbf0f  last data came from RS-232 */
extern unsigned char sci_esc_state;              /* 0xffbf11 */
extern volatile int motor_busy_x;                         /* 0xffbf14 */
extern volatile long motor_pos_x;                         /* 0xffbf26 */
extern volatile int motor_busy_y;                         /* 0xffbf58 */
extern volatile long motor_pos_y;                         /* 0xffbf6a */
extern int regmark_cfg;                          /* 0xffc100 */
extern int sensor_scan_speed;                    /* 0xffc11a */
extern long scan_off_x;                          /* 0xffc11c */
extern long scan_off_y;                          /* 0xffc120 */
extern int sensor_gain_x;                        /* 0xffc124 */
extern int sensor_gain_y;                        /* 0xffc126 */
extern int sensor_baseline;                      /* 0xffc128 */
extern char regmark_scan_mode;                   /* 0xffc222 */
extern char regmark_length;                      /* 0xffc223 */
extern char regmark_width;                       /* 0xffc224 */
extern char regmark_result;                      /* 0xffc225 */
extern volatile char cancel_request;             /* 0xffc226  ESC 03 (USB) / ESC 02 (SCI) */
extern int sensor_trace_len;                     /* 0xffc228 */
extern int sensor_trace_rd;                      /* 0xffc22a */
extern int sensor_threshold;                     /* 0xffc22c */
extern int libc_errno;                           /* 0xffc22e */
extern char demo_mode;                           /* 0xffc2da */
extern char usb_cfg_pending;                     /* 0xffc2ee */
extern long home_off_x;                          /* 0xffc394 */
extern long home_off_y;                          /* 0xffc398 */
extern char deferred_valid;                      /* 0xffc422 */
extern volatile unsigned char jog_keys;          /* 0xffc424  ESC 00 n, active-low arrow bits */
extern short usb_ep0_state;                      /* 0xffc43e */
extern unsigned char *usb_ctl_ptr;                         /* 0xffc4a2 */
extern unsigned char *usb_ctl_buf;                         /* 0xffc4a6 */
extern unsigned char *usb_ctl_end;                         /* 0xffc4ae */
extern int sensor_trace;                         /* 0xffcf00 */

/* ---- functions ---- */
void clip_sort_hits(char mode, long n, long *pts);           /* 0x04100 */
void regmark_find_corners(void);                             /* 0x046b8 */
void regmark_line_intersect(int *parallel, pt_t *out, pt_t d, pt_t c, pt_t b, pt_t a);                  /* 0x048ba */
void regmark_axis_intersect(int *parallel, pt_t *out, pt_t c, pt_t b, pt_t a);                  /* 0x05082 */
void regmark_get_mark(int idx, long *x, long *y);            /* 0x05452 */
void regmark_compute_transform(void);                        /* 0x054e4 */
long regmark_scale_x(double dx, double dy);                                  /* 0x05bc0 */
long regmark_scale_y(double dx, double dy);                                  /* 0x05d40 */
void regmark_scale_fail_hook(void);                          /* 0x05ec0 */
void regmark_extend_segment(long *p1, long *p2);             /* 0x05ecc */
void clip_segment_hits(long poly_top, long *out, long x0, long y0, long x1, long y1, int *n_hits, int *n_touch, int *flags); /* 0x06140 */
long clip_edge_local(void);                                  /* 0x06950 */
void workbuf_layout(void);                                   /* 0x06aac */
void workbuf_fix_sizes(void);                                /* 0x06ba8 */
void workbuf_init(void);                                     /* 0x06c3c */
long approx_dist(long dx, long dy);                          /* 0x06c76 */
void cmd_BZ(void);                                           /* 0x071d6 */
long labs_(long v);                                          /* 0x07378 */
void hw_init_bus_ports(int keep_port);                       /* 0x0739c */
void cmd_speed(void);                                        /* 0x074bc */
void cmd_factor(void);                                       /* 0x07522 */
void cmd_force(void);                                        /* 0x07570 */
void cmd_rotate(void);                                       /* 0x075e2 */
void cmd_term_unused(void);                                  /* 0x07636 */
void cmd_nop_T(void);                                        /* 0x0766e */
void cmd_FA_get_distance_corr(void);                         /* 0x0767a */
void cmd_FB_set_distance_corr(void);                         /* 0x076be */
void cmd_FJ(void);                                           /* 0x07832 */
void set_area_extent(long x, long y);                        /* 0x07984 */
void cmd_FL_get_media_mode(void);                            /* 0x07a3a */
void cmd_FM_set_media_mode(void);                            /* 0x07a64 */
void set_media_mode(char mode);                              /* 0x07aa4 */
void cmd_FO_page_length(void);                               /* 0x07c82 */
void cmd_FQ_query(void);                                     /* 0x07d46 */
void cmd_FV_get_media(void);                                 /* 0x07ec8 */
void cmd_FW_set_media(void);                                 /* 0x07f1a */
void cmd_FX_force(void);                                     /* 0x07f7a */
void cmd_TT(void);                                           /* 0x08002 */
long read_opt_angle(void);                                   /* 0x0806a */
void curve_cmd(int relative, long x0, long y0);              /* 0x0808c */
void cmd_ellipse(void);                                      /* 0x081f8 */
void cmd_WP(void);                                           /* 0x0828e */
void circle3_draw(long x1, long y1, long x2, long y2, long x3, long y3, int a, double d); /* 0x0834c */
void cmd_W(void);                                            /* 0x08c36 */
void cmd_Y(void);                                            /* 0x08cf8 */
void cmd_rel_circle(void);                                   /* 0x08d14 */
void cmd_rel_curve(void);                                    /* 0x08df8 */
void ellipse_draw(void *ell);                                /* 0x092e8 */
void arc_draw_float(long *arc);                              /* 0x09824 */
void spline_feed(int op, long x, long y);                    /* 0x09992 */
void spline_segment(long *xs, long *ys, int pen);            /* 0x09d24 */
long spline_coeffs(long *c, const long *p);                     /* 0x09ebc */
long spline_eval_fx(const long *c, unsigned int t);             /* 0x09fc4 */
long fx_mul16(long a, int b);                                /* 0x0a12a */
void cmd_FC(void);                                           /* 0x0a268 */
void cmd_FD(void);                                           /* 0x0a316 */
void cmd_FE(void);                                           /* 0x0a3cc */
void cmd_FF(void);                                           /* 0x0a46e */
void cmd_FK(void);                                           /* 0x0a580 */
unsigned int dump_getc(void);                                /* 0x0a5b8 */
void dump_reply(int idx);                                    /* 0x0a5de */
void dump_set_term(void);                                    /* 0x0a676 */
void dump_skip_label(void);                                  /* 0x0a696 */
void dump_skip_frame(char c);                                /* 0x0a6d2 */
void dump_cmd_T(void);                                       /* 0x0a6fc */
void dump_cmd_F(void);                                       /* 0x0a792 */
void dump_loop(void);                                        /* 0x0a8a6 */
void dump_home(void);                                        /* 0x0a9d2 */
void dump_banner(void);                                      /* 0x0aa14 */
void dump_start(void);                                       /* 0x0aa9a */
void svc_o_dump_mode(void);                                  /* 0x0aab2 */
void dump_advance(int wide);                                 /* 0x0aad4 */
void dump_hex_byte(unsigned char c);                         /* 0x0ab7a */
int dump_format_char(unsigned char c);                       /* 0x0ac58 */
void dump_print_char(unsigned char c);                       /* 0x0acd6 */
void font_defaults(void);                                    /* 0x0adae */
int font_select(int id);                                     /* 0x0ae22 */
void font_draw_char(unsigned char c);                        /* 0x0ae44 */
void font_draw_string(char *s);                              /* 0x0ae5c */
void font_print(char *s);                                    /* 0x0aed6 */
void font_scale_point(void);                                 /* 0x0aef2 */
int font_begin(void);                                        /* 0x0b14a */
void font_setup_matrix(void);                                /* 0x0b212 */
void font_glyph(unsigned char c);                            /* 0x0b352 */
char * font_find_glyph(int code);                            /* 0x0b7ec */
char * font_find_glyph2(unsigned int code);                  /* 0x0b84a */
char * font_lookup(unsigned int code, long table);           /* 0x0b890 */
int font_step(int op);                                       /* 0x0b904 */
void font_run_glyph(int code);                               /* 0x0b950 */
void font_interp(const unsigned char *prog);                                /* 0x0b96a */
int font_op_draw_next(void);                                     /* 0x0bb5a */
void font_op_ef(void);                                       /* 0x0bb70 */
void font_op_arc(void);                                      /* 0x0bbe2 */
void font_op_call(long *sp, int n);                          /* 0x0bc76 */
void font_op_fa(void);                                       /* 0x0bcae */
void font_op_fb(void);                                       /* 0x0bd38 */
void font_op_fc(void);                                       /* 0x0bdf0 */
void font_op_fd(void);                                       /* 0x0be78 */
int font_op_move_next(void);                                   /* 0x0bf5a */
int font_op_return(void);                                    /* 0x0bf6e */
void font_op_helper1(void);                                  /* 0x0bfe8 */
void font_op_helper2(void);                                  /* 0x0c014 */
void font_op_helper3(void);                                  /* 0x0c0a8 */
void font_arc_step(void);                                    /* 0x0c160 */
void font_continue(void);                                    /* 0x0c24c */
void font_arc(void);                                         /* 0x0c2b0 */
void font_read_xy(void);                                     /* 0x0c412 */
void font_unpack_xy(unsigned int b, long unused, long *out); /* 0x0c450 */
void font_read_byte(char **pp, long unused, unsigned long *out); /* 0x0c494 */
void font_draw_to(long *pt);                                 /* 0x0c4d8 */
unsigned long font_map_code(unsigned long code);             /* 0x0c5fc */
void font_metrics_a(unsigned char c, int *out);              /* 0x0c93a */
void font_metrics_b(unsigned char c, int *out);              /* 0x0c968 */
unsigned long font_metrics(unsigned long code, int *out);    /* 0x0c994 */
int end_of_params(void);                                     /* 0x0cb3a */
void gpgl_loop(void);                                        /* 0x0cbf2 */
void skip_command(void);                                     /* 0x0d412 */
void hook_pen_param(void);                                   /* 0x0d448 */
void hook_sync(void);                                        /* 0x0d458 */
void cmd_FP(void);                                           /* 0x0d46c */
void move_mm(long x_mm, long y_mm);                          /* 0x0d4c8 */
void demo_run(void);                                         /* 0x0d512 */
void distcorr_compute(void);                                 /* 0x0f096 */
void distcorr_apply_page(void);                              /* 0x0f330 */
void distcorr_to_user(long *pt);                             /* 0x0f4b2 */
void distcorr_to_device(long *pt);                           /* 0x0f61e */
unsigned int usb_ep_in_fill(long unused, unsigned int ep);   /* 0x0f78a */
void usb_ep_in_start(void);                                  /* 0x0f8ee */
long usb_ep0_setup(long unused, unsigned char ep);           /* 0x0f942 */
void usb_ep0_in(unsigned long status, unsigned char ep);     /* 0x0f9f8 */
unsigned long usb_ep0_out(unsigned long status, unsigned char ep); /* 0x0fa9c */
int usb_std_request(void);                           /* 0x0fb6c */
char usb_vendor_request(void);                               /* 0x0fdf6 */
void usb_class_request(void);                              /* 0x0fe12 */
void cmd_FG_version(void);                                   /* 0x0ff4e */
void hook_nop(void);                                         /* 0x0ffbc */
void hook_B(void);                                           /* 0x0ffc8 */
void cmd_nop_DP_MP(void);                                    /* 0x0ffd4 */
void cmd_nop_EP_OP(void);                                    /* 0x0ffe6 */
void cmd_nop_LP(void);                                       /* 0x0fff8 */
void cmd_nop_RP(void);                                       /* 0x10004 */
void cmd_nop_KP(void);                                       /* 0x10010 */
void cmd_nop_RC(void);                                       /* 0x10022 */
void cmd_nop_userpat(void);                                  /* 0x1002e */
void cmd_nop_I(void);                                        /* 0x1003a */
void cmd_nop_N(void);                                        /* 0x10046 */
void cmd_nop_SA(void);                                       /* 0x10052 */
void chart_feed_hook(void);                                  /* 0x1005e */
void cmd_nop_hatch(void);                                    /* 0x1006a */
void cmd_nop_A(void);                                        /* 0x10076 */
void cmd_nop_dollar(void);                                   /* 0x10082 */
void cmd_nop_SP(void);                                       /* 0x1008e */
void cmd_nop_Q(void);                                        /* 0x1009a */
void cmd_nop_S(void);                                        /* 0x100a6 */
void cmd_nop_R(void);                                        /* 0x100b2 */
void cmd_nop_paren(void);                                    /* 0x100be */
void term_changed_hook(char t);                              /* 0x100ca */
void angle_to_sincos(long centideg, float *sincos);          /* 0x100e2 */
long hypot_l(long dx, long dy);                              /* 0x10180 */
void svc_e_calc(void);                                       /* 0x101fe */
void plot_soft_reset(int keep_term);                         /* 0x102ae */
void cmd_error_mask(void);                                   /* 0x102d4 */
void set_error(unsigned int bits);                           /* 0x10308 */
void pen_tables_init(void);                                  /* 0x10332 */
void load_cutter_offset(void);                               /* 0x104e6 */
long select_pen_params(void);                                /* 0x1054c */
unsigned long cutoff_on_move(long x, long y, int pen);       /* 0x10686 */
void cutoff_swivel(long angle);                              /* 0x10718 */
void cutoff_set_lead(long angle, int pen);                   /* 0x10c1e */
void cutoff_move(long dx, long dy, int pen);                 /* 0x10ca4 */
void cutoff_end(long x, long y, int pen);                    /* 0x10ce0 */
void cutoff_blade_init(int pen);                             /* 0x10d90 */
long cutoff_need_move(long x, long y, int pen);              /* 0x10f4e */
void cutoff_overcut_restore(void);                           /* 0x10fba */
void cutoff_overcut(void);                                   /* 0x10fd8 */
void motion_flush(void);                                     /* 0x1108a */
void clip_reset(void);                                       /* 0x110ba */
void clip_begin(void);                                       /* 0x11102 */
void cmd_clip(void);                                         /* 0x1112c */
unsigned int rx_peek(void);                                  /* 0x11228 */
unsigned int rx_getc(void);                                  /* 0x11368 */
void reply_begin(int nfields);                               /* 0x113e4 */
void reply_putc(char c);                                     /* 0x1141e */
void usb_reply_send(void);                                   /* 0x1145a */
void reply_send(void);                                       /* 0x114ba */
void usb_bulk_out_rx(void);                                  /* 0x11556 */
void usb_rx_flow_check(void);                                /* 0x11796 */
void rx_flush(void);                                         /* 0x1181c */
void rx_flush_masked(void);                                  /* 0x11852 */
void rx_init(void);                                          /* 0x11882 */
int is_terminator(unsigned int c);                           /* 0x11902 */
int cmd_class(unsigned char c);                              /* 0x11936 */
int read_digit(void);                                        /* 0x1199e */
long read_number(void);                                      /* 0x11a36 */
long read_int(void);                                         /* 0x11b72 */
long read_coord(void);                                       /* 0x11bc6 */
long read_angle(void);                                       /* 0x11d50 */
void eeprom_format(int version);                             /* 0x11dc2 */
void eeprom_load_defaults(void);                             /* 0x11e3e */
void eeprom_format_hook(void);                               /* 0x12034 */
void load_axis_correction(void);                             /* 0x12040 */
long clear_device_offsets(void);                             /* 0x121aa */
long settings_load(void);                                    /* 0x121da */
char load_step_size(void);                                   /* 0x12264 */
void load_config_word(void);                                 /* 0x1228e */
void config_word_read(void);                                 /* 0x122c0 */
void config_word_save(void);                                 /* 0x12302 */
void eeprom_check_format(void);                              /* 0x12332 */
void uwire_sk(char level);                                   /* 0x12380 */
void uwire_cs(char level);                                   /* 0x123a0 */
void uwire_di(char level);                                   /* 0x123c0 */
void uwire_shift_out(int nbits, unsigned long bits);         /* 0x123e0 */
void uwire_command(int nbits, unsigned long bits);           /* 0x12440 */
void eeprom_ewen(void);                                      /* 0x1246a */
void eeprom_ewds(void);                                      /* 0x12492 */
void eeprom_erase_all(void);                                 /* 0x124ba */
void eeprom_write_byte(unsigned char value, long addr);      /* 0x12786 */
unsigned char eeprom_read_byte(long addr);                   /* 0x127d8 */
void eeprom_write_word(int value, long addr);                /* 0x12828 */
int eeprom_read_word(long addr);                             /* 0x1287a */
void eeprom_write_long(long value, long addr);               /* 0x128ca */
long eeprom_read_long(long addr);                            /* 0x12922 */
void eeprom_set_bit(char bit, long addr);                    /* 0x1298a */
void eeprom_clear_bit(char bit, long addr);                  /* 0x129c0 */
void eeprom_toggle_bit(char bit, long addr);                 /* 0x129f8 */
void tick_timer_init(void);                                  /* 0x12a2e */
void idle_timer_reset(void);                                 /* 0x12a88 */
void idle_timer_reset2(void);                                /* 0x12ab4 */
void svc_j_counters(void);                                   /* 0x12ad0 */
void thermal_protect(void);                                  /* 0x12b18 */
void timer_tick_isr(void);                                   /* 0x12d7a */
void delay_ticks(int ticks);                                 /* 0x12e66 */
void delay_ticks_jog(int ticks);                             /* 0x12e88 */
void power_off(int error);                                   /* 0x12ec0 */
void power_down_hw(void);                                    /* 0x12f64 */
int power_up_hw(void);                                       /* 0x12fb8 */
void power_key_isr(void);                                    /* 0x12ff2 */
void spin_delay(int n);                                      /* 0x13164 */
void spin_nop(void);                                         /* 0x1318a */
void init_terminators(void);                                 /* 0x131c0 */
void set_lower_left(long x, long y);                         /* 0x131f2 */
int set_upper_right(long x, long y);                         /* 0x1321e */
void reply_number(int width, long value);                    /* 0x13252 */
void reply_number_comma(int width, long value);              /* 0x133be */
void reply_number_end(int width, long value);                /* 0x133ee */
void settings_apply(void);                                   /* 0x13436 */
void test_key_led(void);                                     /* 0x1344e */
void test_key_port(void);                                    /* 0x13486 */
void test_motion_xy(void);                                   /* 0x134be */
void test_motion_loop(void);                                 /* 0x13516 */
void test_circle(void);                                      /* 0x1355e */
void test_jog(void);                                         /* 0x136b4 */
void test_sincos_bench(void);                                /* 0x136f6 */
void factory_test_select(void);                              /* 0x13736 */
void enter_gpgl(void);                                       /* 0x137c4 */
void boot_wait_key(void);                                    /* 0x137ec */
void warm_restart(void);                                     /* 0x139ce */
long muldiv_l(long a, long b, long c);                       /* 0x13a28 */
void muldiv_test(void);                                      /* 0x13a64 */
void reset_line_type(void);                                  /* 0x13a8c */
void read_point(long *x, long *y);                           /* 0x13ab2 */
void plot_abs(int pen);                                      /* 0x13ada */
void plot_rel(int pen);                                      /* 0x13b16 */
void cmd_B(void);                                            /* 0x13b7e */
void cmd_D(void);                                            /* 0x13bca */
void cmd_E(void);                                            /* 0x13be8 */
void cmd_chart_feed(void);                                   /* 0x13c06 */
void cmd_H(void);                                            /* 0x13c22 */
void cmd_J(void);                                            /* 0x13c58 */
void cmd_L(void);                                            /* 0x13c68 */
void cmd_M(void);                                            /* 0x13d3e */
void cmd_O(void);                                            /* 0x13d50 */
void axis_draw_rel(long d, long t);                          /* 0x13d62 */
void axis_tick_exact(long x0, long y0, long q, int n, int r); /* 0x13dc2 */
void cmd_X(void);                                            /* 0x13eb4 */
void cmd_set_origin(void);                                   /* 0x13ffa */
void set_origin_here(void);                                  /* 0x14008 */
void draw_test_cut_mark(void);                               /* 0x14068 */
void cmd_FH_test_cut(void);                                  /* 0x141d2 */
void cmd_FI_test_plot_loop(void);                            /* 0x14246 */
void burnin_circles(long x);                                 /* 0x146ee */
void burnin_rect(long x);                                    /* 0x1476a */
void burnin_circle_at(long x, long y);                       /* 0x14806 */
void cmd_FT_origin(void);                                    /* 0x148ce */
void cmd_FU_track_feed(void);                                /* 0x148e8 */
void cmd_FY_track_enhance(void);                             /* 0x14b12 */
void axis_profile_step(void *axis);                          /* 0x14b5e */
void home_carriage(void);                                    /* 0x14ef0 */
void page_limits_update(void);                               /* 0x14ff4 */
void page_limits_apply(int arg);                             /* 0x150dc */
void origin_reset(int arg);                                  /* 0x151be */
char regmark_find_type1(void);                               /* 0x15212 */
char regmark_find_type2(void);                               /* 0x154ca */
void regmark_search_manual(void);                            /* 0x1575a */
void regmark_goto_origin(void);                              /* 0x1596c */
void regmark_compute_spans(void);                            /* 0x15a1a */
long regmark_origin_margin(void);                            /* 0x15b16 */
long regmark_scan_first(void);                               /* 0x15bbc */
void regmark_search_auto(void);                              /* 0x15fb2 */
long muldiv(long a, long b, long c);                         /* 0x16130 */
long div64_32(long hi, long lo);                             /* 0x161ae */
int cmp64(long ahi, long alo, long bhi, long blo);           /* 0x1631a */
int vector_angle(long *v);                                   /* 0x164c6 */
unsigned long isqrt32(unsigned long x);                      /* 0x165b0 */
void d_sqrt(void *res, void *unused, void *out, unsigned long hi, unsigned long lo); /* 0x16700 */
void svc_g_param(void);                                      /* 0x16bc0 */
void svc_g0_cfg_bit1(void);                                  /* 0x16c8e */
void svc_g1_cfg_bit6(void);                                  /* 0x16cfc */
void svc_g2_axis_correction(void);                           /* 0x16d6a */
void svc_g3_cfg_bit7(void);                                  /* 0x16e22 */
void svc_g4_motion_param(void);                              /* 0x16e90 */
void svc_g5_cfg_bit4(void);                                  /* 0x16f0c */
void svc_g6_step_size(void);                                 /* 0x16f7a */
void svc_g7_solenoid_profile(void);                          /* 0x16ff8 */
void svc_g10_power_save(void);                               /* 0x17082 */
void svc_g11_force_table(void);                              /* 0x170e8 */
void svc_g12_regmark_param(void);                            /* 0x171b2 */
void svc_g13_regmark_nibbles(void);                          /* 0x17220 */
void svc_g14_sensor_offset(void);                            /* 0x172dc */
void svc_g15_regmark_calibrate(void);                        /* 0x17386 */
void axis_correction_pattern(void);                          /* 0x173b4 */
void svc_i_feed_test(void);                                  /* 0x174fc */
void svc_m_burn_in(void);                                    /* 0x17556 */
void svc_g18_track_passes(void);                             /* 0x175de */
void reply_point(int cmd, long x, long y);                   /* 0x17648 */
void cmd_read_status2(void);                                 /* 0x1769e */
void cmd_gin(int cmd);                                       /* 0x176d6 */
void cmd_call_gin(void);                                     /* 0x17740 */
void cmd_read_upper_right(void);                             /* 0x17752 */
void cmd_upper_right(void);                                  /* 0x1777c */
void cmd_read_lower_left(void);                              /* 0x177b0 */
void cmd_lower_left(void);                                   /* 0x177da */
void cmd_read_offset(void);                                  /* 0x1780e */
void cmd_offset(void);                                       /* 0x17838 */
void load_force_table(void);                                 /* 0x178c4 */
int solenoid_init(void);                                     /* 0x17900 */
void pen_select_hook(void);                                  /* 0x17992 */
void svc_z_port(void);                                       /* 0x1799e */
char solenoid_set(int level);                                /* 0x17c1c */
void pen_reset(void);                                        /* 0x17e8e */
long pen_save_position(void);                                /* 0x17ea6 */
void pen_restore(void);                                      /* 0x17f1e */
void pen_return_to_saved(void);                              /* 0x17f2e */
void select_pen(int pen);                                    /* 0x17fac */
void apply_pen_params(void);                                 /* 0x18020 */
void apply_force(void);                                      /* 0x180b2 */
void set_force_level(char level);                            /* 0x180ec */
int force_level_value(char level);                           /* 0x181da */
int force_hold_level(void);                                  /* 0x1820a */
long clear_tool_offset(void);                                /* 0x18236 */
long gin_pen_state(void);                                    /* 0x18262 */
void set_speed(int speed);                                   /* 0x1829e */
void set_force(int force);                                   /* 0x18316 */
void set_cutter_offset(int offset);                          /* 0x183c0 */
void set_cutter_offset_q(int q);                             /* 0x18440 */
void set_blade_rotation(long angle100);                      /* 0x184a4 */
void set_FE(int on);                                         /* 0x1850a */
void set_FF(long ab);                                        /* 0x18598 */
void select_pen_wrap(int pen);                               /* 0x1860e */
void xform_orientation(double *x, double *y);                    /* 0x1863e */
void xform_rotate(void *x, void *y);                         /* 0x186e6 */
void xform_factor(double *x, double *y);                         /* 0x18864 */
void xform_device_offset(void *x, void *y);                  /* 0x18930 */
void xform_device_offset_wrap(void *x, void *y);             /* 0x189ea */
void xform_window_origin(double *x, double *y);                  /* 0x18a12 */
void xform_user_offset(double *x, double *y);                    /* 0x18a78 */
long device_to_logical(void);                                /* 0x18ae0 */
void sync_user_pos(void);                                    /* 0x18e44 */
void sync_device_pos(void);                                  /* 0x18f0e */
void xform_call_long(void *fn, long *x, long *y);            /* 0x18f86 */
void window_reset(void);                                     /* 0x1900c */
void plot_state_reset(void);                                 /* 0x190bc */
void set_factor(long p, long q, long r);                     /* 0x1919a */
void update_window(void);                                    /* 0x19444 */
void device_to_user_pos(long *x, long *y);                   /* 0x19518 */
void svc_c_motion_param(void);                               /* 0x19758 */
long svc_h_speed_profile(void);                              /* 0x19764 */
long motion_build_profile(unsigned long speeds);             /* 0x19780 */
void motion_drain(void);                                     /* 0x19acc */
int motion_line_to(long x, long y, int pen_down);            /* 0x19b9c */
unsigned long motion_plan_segment(long x, long y, int max_angle); /* 0x19f8c */
void motion_init(void);                                      /* 0x1af64 */
unsigned long motion_start_next(void);                       /* 0x1b162 */
int atan_tenths(long dy, long dx);                           /* 0x1bba6 */
void timer_x_isr(void);                                      /* 0x1bc6a */
void timer_y_isr(void);                                      /* 0x1bcc4 */
void jog_x_step(void);                                       /* 0x1bd1e */
void jog_y_step(void);                                       /* 0x1bf32 */
int regmark_area_outcode(long *p);                           /* 0x1c242 */
int  regmark_clip_x(pt_t *p, pt_t *q, long x);               /* 0x1c39a */
int  regmark_clip_y(pt_t *p, pt_t *q, long y);               /* 0x1c414 */
void device_plot_origin(void);                               /* 0x1c4ba */
void plot_device(long x, long y, int pen);                   /* 0x1c572 */
int machine_outcode(const long *p);                                /* 0x1cbc6 */
void machine_clip_x(long *p, const long *q, long lim);               /* 0x1cd30 */
void machine_clip_y(long *p, const long *q, long lim);               /* 0x1cdaa */
void track_feed_min_and_queue(long x, long y, int pen);      /* 0x1ce28 */
void user_to_motor_point(long *x, long *y);                  /* 0x1ce76 */
void move_to_machine_origin(void);                           /* 0x1d00a */
void queue_point(long x, long y, int pen);                   /* 0x1d0bc */
void queue_segment(long x, long y, int pen);                 /* 0x1d10c */
void queue_segment_now(long x, long y, int pen);             /* 0x1d31a */
void sub_home_offset(long *x, long *y);                      /* 0x1d3c2 */
void motion_emit(long x, long y, int pen);                   /* 0x1d3fe */
void motion_line(long x, long y, int pen);                   /* 0x1d4aa */
void motion_line_wait_set(long x, long y, int pen);          /* 0x1d4f6 */
void motion_line_wait(long x, long y, int pen);              /* 0x1d550 */
void motion_move_steps(long x, long y, int pen);             /* 0x1d582 */
void motion_get_pos_wait(long *xy);                          /* 0x1d5e0 */
void motion_get_pos(long *xy);                               /* 0x1d5fe */
void svc_param_b620(void);                                   /* 0x1d698 */
void nop_1d6ec(void);                                        /* 0x1d6ec */
void tool_set(int down);                                     /* 0x1d6f8 */
void flush_deferred_segment(void);                           /* 0x1da92 */
long set_carriage_origin(long x, long y);                    /* 0x1dacc */
void wait_motion_done(void);                                 /* 0x1db52 */
void plot_to(long x, long y, int pen);                       /* 0x1db80 */
void set_line_type_raw(int lt);                              /* 0x1dba8 */
void set_line_type(int type, long pitch);                    /* 0x1dbc0 */
void dash_restart(void);                                     /* 0x1dc18 */
void dash_next(void);                                        /* 0x1dc5e */
void dash_next_element(void);                                /* 0x1dd5a */
long seg_length(long x, long y);                             /* 0x1deb6 */
void plot_linetype(long x, long y, int pen);                 /* 0x1df5e */
void plot_transform(long x, long y, int pen);                /* 0x1e23e */
void nop_1e792(void);                                        /* 0x1e792 */
int window_outcode(const long *pt);                                /* 0x1e7a8 */
void rotate_point(long *x, long *y);                         /* 0x1e86a */
long clip_isect_xmax(long x1, long y1, long x2, long y2);    /* 0x1e9a6 */
long clip_isect_xmin(long x1, long y1, long x2, long y2);    /* 0x1e9fc */
long clip_isect_ymax(long x1, long y1, long x2, long y2);    /* 0x1ea52 */
long clip_isect_ymin(long x1, long y1, long x2, long y2);    /* 0x1eaa8 */
void plot_solid(long x, long y, int pen);                    /* 0x1eafe */
void plot_raw(long x, long y, int pen);                      /* 0x1eb2c */
void set_pos_and_plot(long x, long y, int pen);              /* 0x1eb52 */
void plot_clip_poly(long x, long y, int pen);                /* 0x1eb88 */
int jog_read_keys(int *st);                                  /* 0x1ed42 */
void jog_with_keys(void);                                    /* 0x1edd6 */
void sci0_init(void);                                        /* 0x1f458 */
void sci0_rxi_isr(void);                                     /* 0x1f4a4 */
void sci0_eri_isr(void);                                     /* 0x1f592 */
void sci0_txi_isr(void);                                     /* 0x1f5c6 */
void sci0_tei_isr(void);                                     /* 0x1f61c */
void nop_1f62c(void);                                        /* 0x1f62c */
void nop_1f638(void);                                        /* 0x1f638 */
void record_stack_ptr(void);                                 /* 0x1f656 */
void nop_1f660(void);                                        /* 0x1f660 */
void usb_tx_idle(void);                                      /* 0x1f66c */
void nop_1f6be(void);                                        /* 0x1f6be */
void cmd_TB(void);                                           /* 0x1f6cc */
void regmark_set_orientation(char landscape);                /* 0x1f79c */
void cmd_TB50_orientation(void);                             /* 0x1f8aa */
void cmd_TB98_query(void);                                   /* 0x1f90a */
char regmark_enable(int on);                                 /* 0x1f930 */
void cmd_TB51_mark_length(void);                             /* 0x1f956 */
void cmd_TB52_mark_type(void);                               /* 0x1f9a0 */
void cmd_TB53_mark_width(void);                              /* 0x1f9ee */
void cmd_TB54_mark_offset(void);                             /* 0x1fa30 */
void cmd_TB55_flag(void);                                    /* 0x1fa6c */
void cmd_TB70_calibrate(void);                               /* 0x1fa9c */
void cmd_TB71_get_calibration(void);                         /* 0x1fab4 */
void cmd_TB72_set_calibration(void);                         /* 0x1faf8 */
void cmd_TB2x_search_manual(int type);                       /* 0x1fb80 */
void cmd_TB12x_search_auto(int type);                        /* 0x1fbe6 */
int regmark_calibrate(void);                                 /* 0x1fc4c */
void cmd_FN_orientation(void);                               /* 0x201f8 */
void set_orientation(char landscape);                        /* 0x20268 */
void set_user_line_type(int n);                              /* 0x2032e */
void load_line_patterns(int n, void *unused);                /* 0x20478 */
void load_line_pattern(int idx, unsigned int *src);          /* 0x204ce */
void nop_2060a(void);                                        /* 0x2060a */
void regmark_init(void);                                     /* 0x20616 */
void regmark_load_calibration(void);                         /* 0x206d6 */
void sensor_on(void);                                        /* 0x2078a */
void sensor_off(void);                                       /* 0x207bc */
int move_clamped_wait(long x, long y, int pen);              /* 0x21546 */
void sensor_init(void);                                      /* 0x2195c */
void sensor_pot_set_level(int level);                                  /* 0x219fa */
int sensor_autocal_gain(int blink_forever);                  /* 0x21a26 */
void svc_n_threshold(void);                                  /* 0x21af8 */
void svc_a(void);                                            /* 0x21b22 */
void svc_t_sensor(void);                                     /* 0x21b96 */
void sensor_reply_adc(void);                                 /* 0x21c40 */
void sensor_measure_reply(void);                             /* 0x21cb6 */
void sensor_trace_start(void);                               /* 0x21cd6 */
void sensor_trace_read(void);                                /* 0x21d42 */
void usb_isr(void);                                          /* 0x220b2 */
void usb_bus_reset_handler(void);                            /* 0x223f4 */
void usb_suspend_handler(void);                              /* 0x2288c */
void usb_resume_handler(void);                               /* 0x228c0 */
unsigned long be_bytes_to_ulong(const unsigned char *p, char n);   /* 0x2292e */
unsigned long le_bytes_to_ulong(const unsigned char *p, char n);   /* 0x22986 */
long __divmodsi4(long a, long b);                            /* 0x23052 */
long __udivmodsi4(long a, long b);                           /* 0x23078 */
long __mulsi3(long a, long b);                               /* 0x2381e */
void __dcopy(void *src, void *dst);                          /* 0x23836 */
void __movblk(void *src, void *dst, unsigned long n);        /* 0x23854 */

/* ---- hand-written additions (names not typed in symbols.txt) ---- */
extern char tx_buf[];                            /* 0xffb2f3  reply buffer */
extern int sci_tx_idx;                           /* 0xffbf0c */
extern unsigned long usb_ier_shadow;             /* 0xffc426 */
extern unsigned char *workbuf_ptrs[7];           /* 0xffb0e8  6 KiB pool at 0xffc800 */
extern long workbuf_sizes[6];                    /* 0xffb104 */
void apply_orientation(long x, long y, long *out);   /* 0x1c9c0 */
void device_plot(long x, long y, int pen);           /* 0x1c608 */
void device_plot_noclip(long x, long y, int pen);      /* 0x1c492 */
void device_plot_raw(long x, long y, int pen);         /* 0x1c4d4 */
void device_plot_nocal(long x, long y, int pen);       /* 0x1c678 */
void motor_run_continuous(int run, int vx, int vy);    /* 0x1d7ac */
void regmark_clear_transform(void);                    /* 0x045e4 */
void regmark_transform(pt_t *out, long x, long y);     /* 0x1c6c4 */
void save_return_addr(void);                           /* 0x1f644 */
void _INITSCT(void);                                   /* 0x238ec  C runtime: copy .data, clear .bss */
int  main(void);                                       /* 0x13940 */
void motion_clear_counters(void);                      /* 0x1af28  orphan */
long hypot_approx(long a, long b);                     /* 0x1bb22  orphan */
extern long plan_pos_x, plan_pos_y;                    /* 0xffb5e0 / 0xffb5e4 */
extern long last_target_x, last_target_y;              /* 0xffb600 / 0xffb604 */
extern long last_target_x_bl;                          /* 0xffb608 */
#define CONFIG_BYTE (*(unsigned char *)0xffb368) /* persistent config bits */

#endif /* CC100_H */
