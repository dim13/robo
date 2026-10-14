/*
 * CC100 firmware 2.30 -- system subsystems in 0x11000-0x1bfff (wave 2, fork F).
 *
 * Hand-cleaned from the Ghidra decompile. Addresses of the original functions are
 * given in comments. int is 16 bits and long is 32 bits on the H8S; fixed-width
 * types are used here to keep that explicit. See notes/subsys_F.md for the overview.
 *
 * Trivial or unresolved functions not reproduced below:
 *   0x12034 0x1318a 0x1615e 0x1617c 0x17f06 0x17f12 0x18224 0x18256 0x18626 0x18918  empty hooks
 *   0x1626a 0x162c8 0x162e6 0x161ae 0x1631a  64-bit helpers for the 'e' self test
 *   0x163a4 0x163d8 0x1640c  wrappers for double routines 0x23d9c / 0x23c1a / 0x23948
 *   0x16440 0x16700 0x16982  double helpers (atan guard, sqrt, series) -- see runtime notes
 *   0x16c8e..0x17386  service 'g' get/set handlers -- table in notes/subsys_F.md section 10
 *   0x15212 0x154ca 0x15bbc 0x1596c 0x15a1a 0x15b16  registration-mark search sequences
 *   0x14ff4 0x150dc 0x151be  page limit / origin reset
 *   0x18ae0 0x19518  device -> logical position (inverse transform)
 *   0x19780 0x19f8c 0x1b162 0x1bd1e 0x1bf32  speed-table build, segment planner, queue start, jog ISRs
 */

#include "cc100.h"

#include <stdint.h>

/* ---- I/O ----------------------------------------------------------------- */

#define REG8(a)  (*(volatile uint8_t  *)(a))
#define REG16(a) (*(volatile uint16_t *)(a))
#define REG32(a) (*(volatile unsigned long *)(a))

#define PORT_UWIRE    REG8(0xFFFF00)  /* bit5 SK, bit6 CS, bit7 DI of the settings EEPROM */
#define PORT_LED      REG8(0xFFFF09)  /* bit3 busy LED (active low) */
#define PORT_UWIRE_DO REG8(0xFFFFB0)  /* bit4 EEPROM DO / ready */
#define PORT_JUMPERS  REG8(0xFFFFB2)  /* bits 5..2 factory test selector */
#define PORT_KEY      REG8(0xFFFFBE)  /* bit0 power key */
#define TMR0_TCR      REG8(0xFFFF68)
#define TMR0_TCSR     REG8(0xFFFF6A)
#define TMR0_TCORA    REG8(0xFFFF6C)
#define DA_SOLENOID   REG8(0xFFFDAC)  /* D/A output driving the tool solenoid */
#define DA_CONTROL    REG8(0xFFFDAE)
#define LATCH_Y_PHASE REG16(0x600000) /* external stepper phase latches */
#define LATCH_X_PHASE REG16(0x600008)
#define LATCH_CURRENT REG16(0x600010) /* motor current: bits 3..2 X, 1..0 Y; 0x0A run, 0x0F off */
#define USB_CTRL      REG8(0xC00080)
#define USB_INT_EN    REG32(0xC000C4) /* bit16 bulk IN, bit18 bulk OUT interrupt enable */

#define EEPROM(off) (0x300000L + (off)) /* firmware's address space for EEPROM bytes */

/* ---- functions named elsewhere (first wave) ------------------------------- */

void     delay_ticks(int ticks);                  /* 0x12e66, 1 tick = 1 ms */
uint8_t  eeprom_read_byte(long addr);             /* 0x127d8 */
void     eeprom_write_byte(uint8_t v, long addr); /* 0x12786 */


void     eeprom_factory_reset(void);              /* 0x11f7e */
void     home_carriage(void);
void     demo_run(void);                          /* 0x0d512 */
void     gpgl_loop(void);                         /* 0x0cbf2 */
void     load_cutter_offset(void);
void     rx_init(void);                           /* 0x11882 */
void     apply_pen_params(void);
void     motion_flush(void);                      /* 0x1108a */
void     usb_init(void);                          /* 0x06c3c */
void     sci0_init(void);                         /* 0x1f458 */
void     motion_init(void);                       /* 0x1af64 */
void     settings_apply(void);                    /* 0x13436 */
int      solenoid_init(void);                     /* 0x17900 */
void     pen_reset(void);                         /* 0x17e8e */
/* motion_start_next: 0x1b162 */
/* motion_plan_segment: 0x19f8c */
void     axis_isr_x_jog(void);                    /* 0x1bd1e */
void     axis_isr_y_jog(void);                    /* 0x1bf32 */
long     pen_save_position(void);                 /* 0x17ea6 */
void     pen_return_to_saved(void);               /* 0x17f2e */
void     motion_drain(void);                      /* 0x19acc */
void     motion_wait(void);                       /* 0x1db52 */
char     solenoid_set(int level);                 /* 0x17c1c */
void     power_off(int error);                    /* 0x12ec0 */
void     spin_delay(int n);                       /* 0x13164 */
void     warm_restart(void);                      /* 0x139ce */
void     enter_gpgl(void);                        /* 0x137c4 */
void     factory_test_select(void);               /* 0x13736 */
void     test_key_led(void), test_key_port(void), test_motion_xy(void); /* 0x1344e 0x13486 0x134be */
void     test_circle(void), test_jog(void), test_sincos_bench(void);    /* 0x1355e 0x136b4 0x136f6 */
int      atan_tenths(long dy, long dx);           /* 0x1bba6 */

/* ---- globals ------------------------------------------------------------- */

/* usb_suspended: 0xFFB397, really the "power off" flag */
/* service_mode: 0xFFB22E */
/* rx_starved: 0xFFB35E */

/* cur_pen: 0xFFB414 */
/* port600010_shadow: 0xFFB5DE */
extern long     heat_x, heat_y, heat_solenoid; /* 0xFFB38A, 0xFFB38E, 0xFFB392 */
extern uint8_t  motion_test_mode;       /* 0xFFB396: suppresses thermal model and idle lift */
/* blade_init_active: 0xFFB2C2 */
/* cut_force: 0xFFB122 */

extern int8_t   media_percent[];        /* 0xFFC385, indexed by media type */
extern uint8_t  solenoid_profile;       /* 0xFFB412, EEPROM 0x2c */
extern uint8_t  speed_index;            /* low byte of 0xFFB4BC */
extern uint8_t  pen_is_down;            /* 0xFFB410 */
extern int16_t  pen_level;              /* 0xFFB40E */
extern int16_t  press_settle, lift_delay, lift_settle; /* 0xFFB408, 0xFFB40A, 0xFFB40C */
extern volatile uint8_t usb_rx_blocked; /* 0xFFB2F2 */
extern unsigned long usb_ier_shadow;         /* 0xFFC426 */
struct kick { int16_t ticks; uint8_t pad; uint8_t level; };
extern struct kick lift_profile[5];         /* 0xFFC368 */
extern struct kick press_profile[2][3][5];  /* 0xFFC2F0 [profile][speed][step] */
extern int16_t  force_table[4];         /* 0xFFC37C, entries 1..3 from EEPROM 0x08/0x0a/0x0c */
extern long     home_off_x, home_off_y; /* 0xFFC394 / 0xFFC398 */
extern uint8_t  pen_params_dirty;

/* ======================================================================== */
/* Settings EEPROM: Microwire 93Cxx, bit-banged                              */
/* ======================================================================== */

static void uwire_pin(uint8_t mask, char level)  /* 0x12380 SK, 0x123a0 CS, 0x123c0 DI */
{
	if (level)
		PORT_UWIRE |= mask;
	else
		PORT_UWIRE &= ~mask;
}

/* 0x123e0: shift nbits out MSB first, data changes while SK is low */
void uwire_shift_out(int nbits, unsigned long bits)
{
	int i;

	uwire_pin(0x20, 0);
	for (i = 0; i < nbits; i++) {
		uwire_pin(0x80, (bits & 0x80000000UL) != 0);
		bits <<= 1;
		uwire_pin(0x20, 1);
		uwire_pin(0x20, 0); /* TODO: exact SK edge order differs slightly in the original */
	}
	uwire_pin(0x80, 0);
}

/* 0x12440 */
void uwire_command(int nbits, unsigned long bits)
{
	uwire_pin(0x40, 1);
	uwire_shift_out(nbits, bits);
	uwire_pin(0x40, 0);
}

void eeprom_ewen(void) { uwire_command(11, 0x98000000UL); } /* 0x1246a: 1 00 11xxxxxx */
void eeprom_ewds(void) { uwire_command(11, 0x80000000UL); } /* 0x12492: 1 00 00xxxxxx */

/* 0x124ba: erase all (ERAL) and wait for the part to report ready on DO */
void eeprom_erase_all(void)
{
	eeprom_ewen();
	uwire_command(11, 0x88000000UL);  /* 1 00 10xxxxxx */
	uwire_pin(0x40, 1);
	while (PORT_UWIRE_DO & 0x10)      /* TODO: polarity as in original: loops while bit4 set */
		;
	uwire_pin(0x40, 0);
	eeprom_ewds();
}

long eeprom_read_long(long addr)                 /* 0x12922: low word first */
{
	return ((long)eeprom_read_word(addr + 2) << 16) | (uint16_t)eeprom_read_word(addr);
}

void eeprom_write_long(long value, long addr)    /* 0x128ca */
{
	eeprom_write_word((int16_t)value, addr);
	eeprom_write_word((int16_t)(value >> 16), addr + 2);
}

/* 0x11e3e: factory defaults of the device-calibration area */
void eeprom_load_defaults(void)
{
	eeprom_write_word(10001, EEPROM(0x02));  /* format marker */
	eeprom_write_byte(14, EEPROM(0x13));     /* axis tweak A */
	eeprom_write_byte(0, EEPROM(0x14));      /* axis tweak B */
	eeprom_write_word(0, EEPROM(0x16));
	eeprom_write_word(164, EEPROM(0x08));    /* force table level 1..3 */
	eeprom_write_word(177, EEPROM(0x0a));
	eeprom_write_word(203, EEPROM(0x0c));
	eeprom_write_byte(0, EEPROM(0x12));
	eeprom_write_byte('x', EEPROM(0x29));    /* reg-mark nibbles 8 / 7 */
	eeprom_write_byte(0, EEPROM(0x18));
	eeprom_write_byte(5, EEPROM(0x28));      /* reg-mark parameter */
	eeprom_write_word(0, EEPROM(0x20));      /* mark sensor offset X */
	eeprom_write_word(0, EEPROM(0x22));      /* mark sensor offset Y */
	eeprom_write_word(5, EEPROM(0x1a));      /* power-save idle minutes */
	eeprom_write_byte(0, EEPROM(0x2a));      /* X backlash */
	eeprom_write_byte(0, EEPROM(0x19));      /* sensor calibration result */
	eeprom_write_byte(0, EEPROM(0x2c));      /* solenoid press profile */
	eeprom_write_byte(3, EEPROM(0x39));      /* track enhancing passes */
	eeprom_write_long(0 /* TODO: current config word */, EEPROM(0x04)); /* 0x12302 */
}

/* 0x11dc2: format / upgrade the EEPROM. 0 = full erase (factory test, never returns) */
void eeprom_format(int version)
{
	if (version == 0) {
		eeprom_erase_all();
		eeprom_load_defaults();
		eeprom_factory_reset();
	} else if (version == 10000 || version == 10001) {
		eeprom_load_defaults();
		eeprom_factory_reset();
	} else if (version == -1) {
		eeprom_factory_reset();
	}
	/* 10002..10004 and anything else: nothing */
	delay_ticks(10);
	if (version == 0)
		for (;;) {               /* blink busy LED forever */
			delay_ticks(500);
			PORT_LED &= ~0x08;
			delay_ticks(500);
			PORT_LED |= 0x08;
		}
}

/* 0x12332: called at every boot */
void eeprom_check_format(void)
{
	int16_t marker = eeprom_read_word(EEPROM(0x02));

	if (marker < 1)
		marker = 9999;
	if (marker < 10001) {
		eeprom_format(marker + 1);
		eeprom_write_word(10001, EEPROM(0x02));
	}
}

/* ======================================================================== */
/* Tick timer, power and thermal protection                                  */
/* ======================================================================== */

/* 0x12a2e: TMR0 = phi/64, clear on compare A, CMIA interrupt, TCORA = 250 -> 1 kHz at 16 MHz */
void tick_timer_init(void)
{
	heat_x = heat_y = heat_solenoid = 0;
	/* 0xFFB380..0xFFB386: idle counters cleared too */
	TMR0_TCORA = 250;
	TMR0_TCR = 0x4A;
	TMR0_TCSR = 0;
}

static void heat_step(long *acc, unsigned current_sel)
{
	switch (current_sel) {
	case 0: *acc += 4; break;                          /* full current */
	case 1: *acc += 2; break;
	case 2: *acc += (*acc < 37801) ? 2 : -2; break;
	case 3: *acc += (*acc < 37801) ? -4 : -6; break;   /* standby */
	}
	if (*acc < 0)
		*acc = 0;
}

/* 0x12b18: run every 100 ticks from the tick ISR (0x12d7a) */
void thermal_protect(void)
{
	int delta;

	if (motion_test_mode)
		return;
	heat_step(&heat_x, (port600010_shadow >> 2) & 3);
	if (heat_x > 54000)
		power_off(1);
	heat_step(&heat_y, port600010_shadow & 3);
	if (heat_y > 54000)
		power_off(1);

	delta = ((int)DA_SOLENOID - 180) / 10;
	if (delta < -2)
		delta = -2;
	heat_solenoid += delta;
	if (heat_solenoid < 0)
		heat_solenoid = 0;
	if (heat_solenoid > 43200)
		power_off(1);
}

/* 0x12f64 */
void power_down_hw(void)
{
	port600010_shadow = 0x0F;
	LATCH_CURRENT = port600010_shadow;
	DA_SOLENOID = 0;
	PORT_LED |= 0x08;
	/* 0xFFFF0E |= 4, 0xFFFF06 &= ~0x10, axis timers stopped: 0xFFFF24/34 = 0x40, 0xFFFEB0 &= ~6 */
}

/* 0x12fb8 */
int power_up_hw(void)
{
	port600010_shadow = 0x0A;
	LATCH_CURRENT = port600010_shadow;
	PORT_LED &= ~0x08;
	/* idle counters 0xFFB380/0xFFB382 cleared */
	return 0;
}

/* 0x12ec0: enter the "off" state; error != 0 is the overheat shutdown */
void power_off(int error)
{
	if (usb_suspended)
		return;
	usb_suspended = 1;
	/* clear 0xFFBC68, 0xFFC422, 0xFFB2BF */
	power_down_hw();
	USB_CTRL |= 1;               /* soft-disconnect from the bus */
	/* 0xFFFF02 &= ~0x40; 0xFFFDE9 |= 0x81: enable the power-key IRQ */
	if (error)
		for (;;) {               /* double flash */
			PORT_LED &= ~0x08; spin_delay(100);
			PORT_LED |= 0x08;  spin_delay(100);
			PORT_LED &= ~0x08; spin_delay(100);
			PORT_LED |= 0x08;  spin_delay(1000);
		}
	for (;;)
		;                        /* wait for power_key_isr */
}

/* 0x13164 */
void spin_delay(int n)
{
	while (n--)
		;                        /* calls the empty 0x1318a */
}

static int key_held_chunks(int limit)
{
	int held, i, chunks = 0;

	do {
		held = 0;
		for (i = 0; i < 100; i++) {
			if (PORT_KEY & 1)
				held = 1;
			spin_delay(1);
		}
		if (chunks < limit)
			chunks++;
	} while (held);
	return chunks;
}

/* 0x12ff2: power key interrupt (vector 18) */
void power_key_isr(void)
{
	int chunks;

	if (!usb_suspended) {        /* on -> off */
		key_held_chunks(0);
		power_off(0);
		return;
	}
	chunks = key_held_chunks(51); /* off -> on */
	usb_init();
	power_up_hw();
	usb_suspended = 0;
	USB_INT_EN &= ~1UL;
	usb_ier_shadow = USB_INT_EN | 10;
	REG8(0xC00086) = 0x36;
	if (chunks > 50) {           /* long press: run the ROM demo */
		home_carriage();
		demo_run();
	}
	warm_restart();
}

/* ======================================================================== */
/* Boot                                                                       */
/* ======================================================================== */

/* 0x13940, called from the reset stub at 0x400 */
int main(void)
{
	/* 0x1f644, 0x739c(0), 0x238ec (runtime init), 0x1f660, 0x1f656: port/clock setup TODO */
	tick_timer_init();
	eeprom_check_format();
	usb_init();
	rx_init();
	/* 0x1f66c */
	sci0_init();                 /* 9600 8N2 */
	usb_suspended = 0;
	motion_test_mode = 0;
	cur_pen = 1;
	motion_init();
	settings_apply();
	solenoid_init();
	pen_reset();
	/* 0x1054c, 0x2195c */
	if (PORT_JUMPERS & 0x3c)
		factory_test_select();
	power_off(0);                /* boot into "off"; power key starts warm_restart */
	return 0;
}

/* 0x13736 */
void factory_test_select(void)
{
	switch ((PORT_JUMPERS >> 2) & 0x0f) {
	case 1: eeprom_format(0); break;
	case 3: test_key_led(); break;
	case 4: test_motion_xy(); break;
	case 5: test_circle(); break;
	case 6: test_jog(); break;
	case 7: test_sincos_bench(); break;
	case 8: /* 0x21a26(1): mark sensor calibration */ break;
	case 9: test_key_port(); break;
	}
}

/* 0x139ce */
void warm_restart(void)
{
	/* 0x1f644, 0x1f656 */
	sci0_init();
	reply_via_sci = 0;
	rx_init();
	motion_init();
	settings_apply();
	solenoid_init();
	pen_reset();
	/* 0x1054c, 0x2195c */
	enter_gpgl();
}

/* 0x137c4 */
void enter_gpgl(void)
{
	service_mode = 0;
	rx_starved = 0;
	home_carriage();
	load_cutter_offset();
	gpgl_loop();                 /* never returns */
}

/* ======================================================================== */
/* Tool solenoid and pen selection                                           */
/* ======================================================================== */

static void block_usb_rx(void)
{
	usb_rx_blocked = 1;
	USB_INT_EN &= ~(1UL << 18);
	usb_ier_shadow = USB_INT_EN | 10;
}

/* 0x17c1c: level 0 lifts, otherwise presses with D/A value `level` */
char solenoid_set(int level)
{
	int i;

	if (level == 0) {
		if (lift_delay)
			delay_ticks(lift_delay);
		block_usb_rx();
		for (i = 0; i < 5 && lift_profile[i].ticks; i++) {
			DA_SOLENOID = lift_profile[i].level;
			delay_ticks(lift_profile[i].ticks);
		}
		DA_SOLENOID = 0;
		usb_rx_blocked = 0;
		delay_ticks(lift_settle);
		pen_is_down = 0;
		pen_level = 0;
		return 0;
	}

	if (!blade_init_active && cut_force < 1) {  /* 0xffb2c2 read here as "force set explicitly" */
		level += level * media_percent[(unsigned char)media_type] / 100;
		if (level < 0)
			level = 0;
		if (level > 255)
			level = 255;
	}
	if (pen_is_down) {           /* already pressed: just change the hold level */
		pen_level = level;
		DA_SOLENOID = (uint8_t)level;
		return 1;
	}
	block_usb_rx();
	delay_ticks(1);
	for (i = 0; i < 5; i++) {
		struct kick *k = &press_profile[solenoid_profile][speed_index - 1][i];
		if (!k->ticks)
			break;
		DA_SOLENOID = k->level;
		delay_ticks(k->ticks);
	}
	usb_rx_blocked = 0;
	pen_level = level;
	DA_SOLENOID = (uint8_t)level;
	delay_ticks(press_settle);
	pen_is_down = 1;
	return 1;
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

/* 0x17fac */
void select_pen(int pen)
{
	if (pen == cur_pen || pen > 8 || pen < 0)
		return;
	motion_flush();
	motion_wait();
	pen_save_position();
	cur_pen = pen;
	/* 0xFFB4B8 = pen */
	apply_pen_params();
	pen_params_dirty = 1;
	/* 0x17992(1) empty hook; 0xFFB298/0xFFB29C = 0; 0x1054c(0) */
	pen_return_to_saved();
	solenoid_set(0);
}

/* ======================================================================== */
/* Coordinate transform stages (doubles)                                     */
/* ======================================================================== */

extern long factor_p, factor_q, factor_r;   /* & command */
extern long offset_x, offset_y;             /* ^ command */
extern long win_origin_x, win_origin_y;     /* 0xFFB57C / 0xFFB580 */
extern int  orient_flags;
extern long orient_max_x, orient_max_y;     /* 0xFFB578 / 0xFFB574 */

/* 0x18864 */
void xform_factor(double *x, double *y)
{
	*x = *x * factor_p / factor_r;
	*y = *y * factor_q / factor_r;
}

/* 0x1863e */
void xform_orientation(double *x, double *y)
{
	double t;

	if (!(orient_flags & 0x20))
		*x = orient_max_x - *x;   /* TODO: operand order of the subtraction */
	if (!(orient_flags & 0x40))
		*y = orient_max_y - *y;
	if (orient_flags != 0 && orient_flags != 0x60) {
		t = *x;
		*x = *y;
		*y = t;
	}
}

/* 0x18a12 (after 0x18930 added the device offsets) */
void xform_window_origin(double *x, double *y)
{
	*x -= win_origin_x;
	*y -= win_origin_y;
}

/* 0x18a78 */
void xform_user_offset(double *x, double *y)
{
	xform_window_origin(x, y);
	*x += offset_x;
	*y += offset_y;
}

/* ======================================================================== */
/* Motion                                                                     */
/* ======================================================================== */

#define QUEUE_LEN 8
extern volatile unsigned long mq_head, mq_tail;   /* 0xFFB7D0 / 0xFFB7D4 */
/* motor_busy_y: 0xFFBF14 / 0xFFBF58: state field */
extern long  last_x, last_y;                 /* 0xFFB600 / 0xFFB604 (requested) */
extern long  queued_x;                       /* 0xFFB608: x including backlash */
extern long  dev_x, dev_y;                   /* 0xFFB5E0 / 0xFFB5E4 */
extern int16_t last_dir;                     /* 0xFFB61C, tenths of a degree */
extern int16_t backlash_now, backlash_x;     /* 0xFFB61E / 0xFFB620 */
extern int16_t last_pen;                     /* 0xFFBBD4 */

static int queue_full(void)  { return ((mq_head + 1) & (QUEUE_LEN - 1)) == mq_tail; }
static int axes_idle(void)   { return motor_busy_x == 0 && motor_busy_y == 0; }

/* 0x19acc */
void motion_drain(void)
{
	if (queue_full() && axes_idle())
		motion_start_next();
	if (queued_x != dev_x || last_y != dev_y)
		motion_plan_segment(queued_x, last_y, 1800);
	while (mq_head != mq_tail) {
		if (axes_idle())
			motion_start_next();
		else
			delay_ticks(1);
	}
	while (!axes_idle())
		delay_ticks(1);
}

/* 0x19b9c: queue a straight move in device steps. Simplified: the speed-profile
 * selection (0xFFB628.. pen down vs 0xFFB634.. pen up) and the exact backlash
 * ramp are summarised in comments. */
int motion_line_to(long x, long y, int pen_down)
{
	long dx, dy;
	int dir, turn = 1800;

	/* idle_timer_reset() */
	if ((pen_down != 0) != last_pen) {
		motion_drain();
		last_pen = pen_down != 0;
	}
	dx = backlash_now - dev_x + x;
	dy = y - dev_y;
	dir = atan_tenths(dy, dx);
	if (dx == 0 && dy == 0)
		return 0;

	if (mq_head != mq_tail) {
		turn = dir - last_dir;
		if (turn < 0)
			turn = -turn;
		if (turn > 1800)
			turn = 3600 - turn;
		if (turn > 250) {        /* more than 25 degrees */
			if ((dx < 0 ? -dx : dx) <= 80) {
				if (axes_idle())
					motion_start_next();
			} else {
				motion_drain(); /* stop at the corner */
			}
		}
	}
	/* select pen-down / pen-up speed profile here (0xFFB60C..0xFFB614) */

	/* X backlash: ramp backlash_now towards +backlash_x when moving +X, 0 when moving -X.
	 * TODO: the ramp steps by 1 or 2 depending on |dx| and is only applied when turn <= 25 deg. */
	if (x != last_x)
		backlash_now = (x > last_x) ? backlash_x : 0;

	queued_x = backlash_now + x;
	motion_plan_segment(queued_x, y, turn);
	if (!pen_down && axes_idle())
		motion_start_next();
	last_x = x;
	last_y = y;
	last_dir = dir;
	return dir;
}

/* 0x1bc6a: X axis timer compare interrupt (vector 40). Y (0x1bcc4, vector 44) mirrors it. */
/* jog_dir_x: 0xFFBBD6 */
extern int16_t queue_loading;  /* 0xFFBBD2 */
extern struct axis axis_x, axis_y;


void timer_x_isr(void)
{
	if (jog_dir_x) {
		axis_isr_x_jog();
		return;
	}
	REG8(0xFFFF25) &= ~1;      /* clear compare flag */
	if (queue_loading == 0 && motor_busy_x) {
		axis_profile_step(&axis_x);
		if (!(REG8(0xFFFF35) & 1) && motor_busy_y)
			axis_profile_step(&axis_y);  /* service the other axis if its flag is pending */
		if (!axes_idle())
			return;
		motion_start_next();
	}
}

/* ======================================================================== */
/* Integer math helpers                                                       */
/* ======================================================================== */

/* 0x165b0: bit-by-bit integer square root, rounded */
unsigned long isqrt32(unsigned long x)
{
	unsigned long root = 0, rem = 0;
	int n = 16;

	if (x == 0)
		return 0;
	while (x < 0x40000000UL) {   /* normalise two bits at a time */
		x <<= 2;
		n--;
	}
	while (n-- > 0) {
		rem = (rem << 2) | (x >> 30);
		x <<= 2;
		root <<= 1;
		if (root * 2 + 1 <= rem) {   /* original keeps 2*root and halves at the end */
			rem -= root * 2 + 1;
			root++;
		}
	}
	if (root <= rem)
		root++;                  /* round */
	return root;
}

/* 0x1bba6: atan2(dy, dx) in tenths of a degree, -1800..1800, polynomial approximation */
int atan_tenths(long dy, long dx)
{
	int neg_y = dy < 0, neg_x = dx < 0;
	long ax = neg_x ? -dx : dx, ay = neg_y ? -dy : dy;
	long lo = ay < ax ? ay : ax, hi = ay < ax ? ax : ay;
	long t, p;
	int a;

	if (hi == 0)
		return 0;
	t = (hi / 2 + lo * 1024) / hi;              /* ratio in 1/1024 */
	p = 0x26064L - ((0x9e64L * t) >> 10);       /* atan(r) ~ r * (k1 - k2 * r), quadratic fit */
	a = (int)((p * t + 0x20000L) >> 18);        /* tenths of a degree, 0..450 */
	if (ay > ax)
		a = 900 - a;
	/* TODO: quadrant fix-up below is reconstructed from the sign tests at the end of 0x1bba6 */
	if (neg_x)
		a = 1800 - a;
	if (neg_y)
		a = -a;
	return a;
}

/* 0x164c6: direction of vector v in tenths of a degree using ROM table 0x28de2
 * (129 entries: atan(i/128) in tenths, 0..450) */
extern const int16_t atan_table[129];
int vector_angle(long *v)
{
	long ax = v[0] < 0 ? -v[0] : v[0];
	long ay = v[1] < 0 ? -v[1] : v[1];
	int a;

	if (ay < ax)
		a = atan_table[(ay << 7) / ax];
	else
		a = 900 - atan_table[(ax << 7) / ay];
	if ((v[0] ^ v[1]) < 0)
		a = -a;
	if (v[0] < 0)
		a += 1800;
	/* TODO: final normalisation to 0..3599 */
	return a;
}
