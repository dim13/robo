/*
 * CC100 firmware 2.30 -- reset, main startup, the application entry table,
 * the alternative boot_wait_key entry and two orphan routines.
 * See ../notes/boot.md. Comments give the original ROM/RAM addresses.
 */
#include "cc100.h"

#define REG8(a)  (*(volatile unsigned char *)(a))
#define REG32(a) (*(volatile unsigned long *)(a))

#define PORT_KEY        REG8(0xffffbe)   /* bit 0 = panel/power key, 1 = pressed */
#define PORT_LED        REG8(0xffff09)
#define PORT_FE14       REG8(0xfffe14) /* 0 while waiting for the key, 4 when running */
#define JUMPERS         REG8(0xffffb2) /* bits 2..5: factory test selection */
#define MSTPCRB         REG8(0xfffde9) /* module stop control B */
#define USB_IER         REG32(0xc000c4)  /* USB controller interrupt enable */
#define USB_REG86       REG8(0xc00086)


int boot_key_ticks;                             /* 0xffb398 */
int boot_key_seen;                              /* 0xffb39a */

/*
 * 0x4000: application entry table. Fixed interface for a boot block in
 * 0x0000-0x3fff. Entries 0..9 are the same targets, in the same order,
 * that the vector wrappers at 0x400-0x719 call directly; entry 10 is only
 * reachable through this table. Nothing in 2.30 reads it.
 */
void (*const app_entry_table[11])(void) = {
	(void (*)(void))main, /* reset */
	timer_x_isr,    /* vec 40  */
	timer_y_isr,    /* vec 44  */
	usb_isr,        /* vec 104 */
	power_key_isr,  /* vec 18  */
	timer_tick_isr, /* vec 64  */
	sci0_eri_isr,   /* vec 80  */
	sci0_rxi_isr,   /* vec 81  */
	sci0_txi_isr,   /* vec 82  */
	sci0_tei_isr,   /* vec 83  */
	boot_wait_key,  /* -- */
};

/*
 * 0x0400 reset (assembly):
 *	mov.l #0xffefb0,sp ; push.l er6 ; mov.l sp,er6 ; orc #0x80,ccr
 *	jsr main ; sleep
 */

/* 0x13940 (never returns: power_off hands over to the key ISR) */
int main(void)
{
	save_return_addr();      /* caller's return address -> 0xffefba, never read */
	hw_init_bus_ports(0);
	_INITSCT();
	nop_1f660();
	record_stack_ptr();      /* SP -> 0xffbf9c, reported by service 'y' */
	tick_timer_init();
	eeprom_check_format();
	workbuf_init();
	rx_init();
	usb_tx_idle();
	sci0_init();
	usb_suspended = 0;       /* really: soft power-off flag */
	REG8(0xffb396) = 0;
	cur_pen = 1;
	motion_init();
	settings_apply();
	solenoid_init();
	pen_reset();
	select_pen_params();
	sensor_init();
	if (JUMPERS & 0x3c)
		factory_test_select();
	power_off(0);            /* boot into soft-off; power key starts gpgl_loop */
	return 0;                /* not reached */
}

/* 0x137ec: alternative entry, reachable only via app_entry_table[10] */
void boot_wait_key(void)
{
	save_return_addr();
	hw_init_bus_ports(1);
	PORT_LED = 0;
	_INITSCT();
	nop_1f660();
	record_stack_ptr();
	tick_timer_init();
	eeprom_check_format();

	/* wait until the key has been released for 100 ms */
	for (boot_key_ticks = 0; boot_key_ticks < 100; boot_key_ticks++) {
		delay_ticks(1);
		if (PORT_KEY & 1)
			boot_key_ticks = -1;    /* pressed: restart the count */
	}
	/* then watch it for 1 s */
	boot_key_seen = 0;
	for (boot_key_ticks = 0; boot_key_ticks < 1000; boot_key_ticks++) {
		delay_ticks(1);
		if (PORT_KEY & 1)
			boot_key_seen = 1;
	}

	PORT_FE14 = 4;
	workbuf_init();
	rx_init();
	usb_tx_remaining = 0;
	/* USB power-up, same sequence as power_key_isr */
	USB_IER &= ~1UL;
	usb_ier_shadow = USB_IER | 0x0a;
	usb_cfg_pending = 1;
	MSTPCRB &= 0xfe;
	USB_REG86 = 0x36;
	sci0_init();
	usb_suspended = 0;
	REG8(0xffb396) = 0;
	motion_init();
	settings_apply();
	solenoid_init();
	pen_reset();
	select_pen_params();
	sensor_init();
	home_carriage();
	if (boot_key_seen)
		demo_run();          /* self test + ROM GP-GL job */
	else
		test_motion_loop();  /* endless */
}

/* 0x1af28: orphan, superseded by the inline clears in motion_init (0x1af64) */
void motion_clear_counters(void)
{
	motor_pos_x = 0;
	plan_pos_x = 0;
	last_target_x = 0;
	last_target_x_bl = 0;
	motor_pos_y = 0;
	plan_pos_y = 0;
	last_target_y = 0;
}

/*
 * 0x1bb22: orphan fast integer sqrt(a*a + b*b) for a, b >= 0.
 * M*(1 + 0.0576 t + 0.3564 t^2), t = min/max: exact at t=0, 1.414*M at t=1.
 * The firmware uses hypot_l (0x10180) instead.
 */
long hypot_approx(long a, long b)
{
	long hi = a, lo = b, t, q;

	if (a < b) {
		hi = b;
		lo = a;
	}
	if (hi == 0)
		return 0;
	t = __divmodsi4(lo << 8, hi);              /* 0..256 */
	q = (long)((unsigned long)(t * (t * 365 + 0x3b00)) >> 16);
	return (hi * (q + 1024)) >> 10;
}
