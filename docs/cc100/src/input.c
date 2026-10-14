/*
 * CC100 firmware 2.30: input layer and GP-GL parser core.
 * Hand-cleaned from the Ghidra decompilation of cc100ver230.bin (H8S, advanced mode).
 * int = 16 bit, long = 32 bit. Original addresses are given as / * 0xNNNN * /.
 * See ../notes/input.md for the prose description.
 */

#include "cc100.h"

#include <stdint.h>


/* RAM globals are declared in cc100.h */

static const uint8_t step_table[4] = { 20, 10, 5, 2 };  /* 0x28d28 */

/* hardware (external USB controller and on-chip SCI0) */
#define USB_ISR    (*(volatile uint32_t *)0xc000c0)
#define USB_IER    (*(volatile uint32_t *)0xc000c4)
#define USB_EP_FIFO (*(volatile uint8_t *)0xc000a4)
#define USB_EP_CNT  (*(volatile uint8_t *)0xc000bd)
#define USB_CTRL   (*(volatile uint8_t *)0xc00084)
#define USB_IE_BULK_OUT 0x40000
#define USB_IE_BULK_IN  0x10000
#define SCI0_SCR   (*(volatile uint8_t *)0xffff7a)
#define SCI0_TDR   (*(volatile uint8_t *)0xffff7b)
#define SCI0_SSR   (*(volatile uint8_t *)0xffff7c)
#define SCI0_RDR   (*(volatile uint8_t *)0xffff7d)
#define PORT600010 (*(volatile int *)0x600010)  /* LED / power latch */

/* ---- ring buffer -------------------------------------------------------- */

static void rx_put(uint8_t c)           /* inlined in both ISRs */
{
	rx_count++;
	*rx_wr_ptr++ = c;
	if (rx_wr_ptr >= rx_buf_base + rx_buf_size)
		rx_wr_ptr = rx_buf_base;
}

/* 0x11882 */
void rx_init(void)
{
	*(char *)0xffb2f2 = 0;
	rx_count = 0;
	rx_buf_base = workbuf_ptrs[0];      /* 0xffb0e8 */
	rx_buf_size = workbuf_sizes[0] - 4; /* 0xffb104 */
	rx_wr_ptr = rx_rd_ptr = rx_buf_base;
	tx_len = 0;
	tx_busy = 0;
	*(char *)0xffb35d = 0;
	*(int *)0xffb2f0 = 0;
}

/* 0x1181c; 0x11852 is the same with interrupts masked */
void rx_flush(void)
{
	rx_count = 0;
	rx_wr_ptr = rx_rd_ptr = rx_buf_base;
}

/* ---- USB bulk OUT: 0x11556 (via usb_isr 0x220b2, table 0x290d2) --------- */

static void immediate_reply(char c)
{
	tx_buf[0] = c;
	tx_len = 1;
	reply_putc(term1);
	if (term1 != term2)
		reply_putc(term2);
	usb_reply_send();
}

void usb_bulk_out_rx(void)
{
	int n;
	uint8_t c;

	if ((unsigned long)rx_count > rx_buf_size - 100) {
		/* buffer almost full: stop accepting, host gets NAKed */
		USB_IER &= ~USB_IE_BULK_OUT;
		usb_ier_shadow = USB_IER | 10;
		return;
	}
	for (n = USB_EP_CNT; n != 0; n--) {
		c = USB_EP_FIFO;
		if (c == ESC) {
			c = USB_EP_FIFO;
			switch (c) {
			case 0x00:                      /* ESC 00 n: remote jog keys */
				idle_timer_reset2();
				jog_keys = USB_EP_FIFO;
				n -= 3;
				continue;
			case 0x01:                      /* debugger hook, then 0x01 stored as data */
				USB_CTRL |= 0x20;
				__asm__("trapa #1");    /* vector 9 is an empty rte */
				break;
			case 0x02:                      /* ESC 02: "1" = USB online */
				immediate_reply(usb_suspended ? '0' : '1');
				n -= 2;
				continue;
			case 0x03:                      /* ESC 03: cancel job */
				cancel_request = 1;
				n -= 2;
				continue;
			case 0x04:                      /* ESC 04: ignored */
				n -= 2;
				continue;
			case 0x05:                      /* ESC 05: "0" = idle, "1" = busy */
				immediate_reply(rx_starved ? '0' : '1');
				n -= 2;
				continue;
			default:
				/* ESC x: ESC dropped, x stored. Bug: n only decremented once below. */
				break;
			}
		}
		if (!usb_suspended)
			rx_put(c);
	}
	USB_CTRL |= 0x20;                       /* re-arm OUT FIFO */
}

/* 0x11796, every timer tick */
void usb_rx_flow_check(void)
{
	/* TODO: exact meaning of the shadow test (bit 18 of 0xffc426) */
	if ((usb_ier_shadow & USB_IE_BULK_OUT) &&
	    (unsigned long)rx_count < rx_buf_size - 110 && !*(char *)0xffb2f2) {
		USB_IER |= USB_IE_BULK_OUT;
		usb_ier_shadow = USB_IER | 10;
	}
}

/* ---- SCI0 (RS-232) ------------------------------------------------------ */

/* 0x1f4a4, vector 81 */
void sci0_rxi_isr(void)
{
	uint8_t c;

	if (*(char *)0xffbf0e) {                /* alternative input path, stub (0x1f62c) */
		return;
	}
	c = SCI0_RDR;
	SCI0_SSR &= ~0x40;                      /* clear RDRF */
	switch (sci_esc_state) {
	case 0:
		if (c == ESC) {
			sci_esc_state = 1;
		} else {
			reply_via_sci = 1;
			rx_put(c);
		}
		break;
	case 1:
		if (c == 0x00)
			sci_esc_state = 2;
		else if (c == 0x02) {           /* ESC 02: cancel (USB uses ESC 03) */
			cancel_request = 1;
			sci_esc_state = 0;
		}
		/* any other byte: swallowed, state stays 1 (firmware bug) */
		break;
	default:                                /* ESC 00 n */
		jog_keys = c;
		sci_esc_state = 0;
		break;
	}
}

/* 0x1f592, vector 80 */
void sci0_eri_isr(void)
{
	if (SCI0_SSR & 0x38) {                  /* ORER | FER | PER */
		*(uint8_t *)0xffbf10 = SCI0_SSR & 0x38;
		SCI0_SSR &= ~0x38;
	}
}

/* 0x1f5c6, vector 82 */
void sci0_txi_isr(void)
{
	SCI0_TDR = tx_buf[sci_tx_idx++];
	SCI0_SSR &= ~0x80;                      /* clear TDRE */
	if (sci_tx_idx >= tx_len) {
		tx_busy = 0;
		SCI0_SCR &= ~0x80;              /* TIE off */
	}
}

/* ---- byte input --------------------------------------------------------- */

/* 0x11228: next byte, not consumed; blocks and does idle work */
unsigned int rx_peek(void)
{
	int i;

	if (rx_count == 0) {
		rx_starved = 1;
		for (i = 0; i < 100; i++) {
			delay_ticks(10);
			if (rx_count)
				goto have_data;
		}
		/* idle: finish everything that is queued */
		hook_nop();
		motion_flush();
		pen_return_to_saved();
		motion_drain();
		wait_motion_done();
		reply_via_sci = 0;
		if (demo_mode)
			demo_run();             /* 0x0d512 */
		for (i = 0; i < 200; i++) {
			delay_ticks_jog(10);
			if (rx_count)
				goto have_data;
		}
		if (!(CONFIG_BYTE & 0x20))
			tool_set(0);            /* tool up */
		if (!jog_keys)
			PORT600010 = port600010_shadow |= 0x0c;
		idle_timer_reset();
		rx_starved = 1;
		while (rx_count == 0) {
			if (!jog_keys)
				PORT600010 = port600010_shadow |= 0x0c;
			delay_ticks_jog(1);
		}
	}
have_data:
	rx_starved = 0;
	idle_timer_reset();
	return *rx_rd_ptr;
}

/* 0x11368 */
unsigned int rx_getc(void)
{
	unsigned int c = rx_peek();

	rx_count--;
	if (++rx_rd_ptr >= rx_buf_base + rx_buf_size)
		rx_rd_ptr = rx_buf_base;
	return c;
}

/* ---- character classes -------------------------------------------------- */

/* 0x11902 */
int is_terminator(unsigned int c)
{
	return c == term1 || c == term2;
}

/*
 * 0x11936: >= 0 for characters that start a command, -1 otherwise.
 * Command characters: '!'..'*', '/', ':'..'_'.
 */
int cmd_class(uint8_t c)
{
	if (c >= 0x60)
		return -1;
	if (c >= 0x3a)
		return c - 0x2f;
	if (c == '/')
		return 10;
	if (c > 0x20 && c < 0x2b)
		return c - 0x21;
	return -1;
}

/*
 * 0xcb3a: 1 if the parameter list ended (terminator, '/', ESC/FS or another
 * command letter is next), 0 if another number follows. Skips ',' and
 * control characters/space. The terminator itself is NOT consumed.
 */
int end_of_params(void)
{
	int c;

	for (;;) {
		c = rx_peek();
		if (is_terminator(c) || c == 0x80)
			return 1;
		c &= 0x7f;
		if (c > 0x2a && c < 0x3a) {     /* + , - . / 0-9 */
			if (c == '/')
				return 1;
			if (c != ',')
				return 0;
		} else if (c == ESC || c == FS || c > 0x20) {
			return 1;
		}
		rx_getc();
	}
}

/* ---- numbers ------------------------------------------------------------ */

#define D_JUNK  (-1)    /* skip it (not consumed) */
#define D_END   (-2)    /* terminator or command char (not consumed) */
#define D_POINT (-3)    /* '.' (consumed) */

/* 0x1199e */
int read_digit(void)
{
	int c;

	while ((c = rx_peek()) == 0)
		rx_getc();
	if (c >= '0' && c <= '9') {
		rx_getc();
		return c - '0';
	}
	if (is_terminator(c) || cmd_class(c) != -1)
		return D_END;
	if (c == '.') {
		rx_getc();
		return D_POINT;
	}
	return D_JUNK;
}

/*
 * 0x11a36: [junk/'-']* digits* ['.' d1 [d2 [digits*]]]
 * Magnitude is limited to 0x7fffff (+) / 0x800000 (-), overflow -> error 8.
 * The fraction digits are left in num_frac1/num_frac2 (unsigned).
 */
long read_number(void)
{
	int  neg = 0;
	long limit = 0x7fffff;
	long acc;
	int  d;

	num_frac1 = num_frac2 = 0;
	while ((d = read_digit()) == D_JUNK) {
		if (rx_getc() == '-') {
			neg = -1;
			limit = 0x800000;
		}
	}
	if (d < 0) {
		acc = 0;
	} else {
		acc = d;
		while ((d = read_digit()) >= 0)
			if (acc <= limit)
				acc = acc * 10 + d;     /* __mulsi3 */
	}
	if (d == D_POINT && (d = read_digit()) >= 0) {
		num_frac1 = d;
		if ((d = read_digit()) >= 0) {
			num_frac2 = d;
			while (read_digit() >= 0)
				;
		}
	}
	if (acc > limit) {
		set_error(8);
		acc = limit;
	}
	return neg ? -acc : acc;
}

/* 0x11b72: rounded to nearest integer (note: "-0.5" -> +1) */
long read_int(void)
{
	long v = read_number();

	if (num_frac1 >= 5)
		v += v < 0 ? -1 : 1;
	return v;
}

/*
 * 0x11bc6: coordinate in internal steps = value * steps_per_unit,
 * fraction rounded to the step grid (note: "-0.x" loses the sign).
 */
long read_coord(void)
{
	long v = read_number();
	int  f;

	switch (steps_per_unit) {
	case 10:
		v *= 10;
		f = num_frac1 + (num_frac2 >= 5);
		break;
	case 5:
		v *= 5;
		f = (num_frac1 + 1) / 2;
		break;
	case 2:
		v *= 2;
		num_frac2 = num_frac1;
		f = num_frac2 > 7 ? 2 : num_frac2 > 2;
		break;
	default:        /* 20 */
		v *= 20;
		f = num_frac1 * 2 + (num_frac2 > 7 ? 2 : num_frac2 > 2);
		break;
	}
	return v < 0 ? v - f : v + f;
}

/* 0x13ab2 */
void read_point(long *x, long *y)
{
	*x = read_coord();
	*y = read_coord();
}

/* 0x12264: step size from persistent config bits 3..2 */
char load_step_size(void)
{
	steps_per_unit = step_table[(*(uint8_t *)0xffb368 >> 2) & 3];
	return steps_per_unit;
}

/* ---- terminators, errors ------------------------------------------------ */

/* 0x131c0 ('=' is a no-op in the dispatcher, so this is the only setter) */
void init_terminators(void)
{
	term2 = ETX;
	term1 = ETX;
	/* 0x100ca: empty stub */
}

/* 0x10308. error_status is never read by any command. */
void set_error(unsigned int bits)
{
	if (bits & error_mask)
		error_status |= bits & error_mask;
}

/* 0x102d4: '"' m[,a,b]: error mask (low word of a 32-bit store at 0xffb282) */
void cmd_error_mask(void)
{
	error_mask = (int)read_int();
	read_int();
	read_int();
}

/*
 * 0xd46c: "FP" + "GRFCC1" unlocks the lowercase service commands.
 * A mismatch stops at the first wrong character, which is consumed.
 */
void cmd_FP(void)
{
	static const char key[] = "GRFCC1";
	int i;

	for (i = 0; key[i]; i++)
		if (rx_getc() != key[i])
			return;
	service_mode = 1;
}

/* ---- replies ------------------------------------------------------------ */

/* 0x113e4: nfields (and a second argument, the command char) are ignored */
void reply_begin(int nfields)
{
	(void)nfields;
	while (tx_busy)
		delay_ticks(10);
	tx_buf[0] = 0;
	tx_len = 0;
}

/* 0x1141e */
void reply_putc(char c)
{
	tx_buf[tx_len++] = c;
	tx_buf[tx_len] = 0;
}

/*
 * 0x13252: right-aligned decimal, at least `width` chars, space padded.
 * Non-zero values carry a sign char (' ' or '-'); zero is "0".
 */
void reply_number(int width, long v)
{
	int  buf[10];   /* original uses 16-bit cells */
	int *p = &buf[9];
	int *start;
	int  sign, i;

	for (i = 0; i < 10; i++)
		buf[i] = ' ';
	if (v == 0) {
		*p = '0';
	} else {
		sign = v < 0 ? '-' : ' ';
		if (v < 0)
			v = -v;
		for (i = 0; i < 10; i++) {
			*p-- = v % 10 + '0';            /* __divmodsi4 */
			v /= 10;
			if (v == 0) {
				*p = sign;
				break;
			}
		}
		if (i == 10) {          /* 10 digits, no room for sign */
			p = buf;
			goto out;
		}
	}
	start = &buf[10 - width];
	if (start < p)
		p = start;
out:
	for (; p <= &buf[9]; p++)
		reply_putc(*p);
}

/* 0x133be */
void reply_number_comma(int width, long v)
{
	reply_number(width, v);
	reply_putc(',');
}

/* 0x133ee */
void reply_number_end(int width, long v)
{
	reply_number(width, v);
	reply_putc(term1);
	if (term1 != term2)
		reply_putc(term2);
	reply_send();
}

/* 0x114ba: send on the interface the last data came from */
void reply_send(void)
{
	if (!reply_via_sci) {
		*(int *)0xffb2ee = 0;
		*(int *)0xffb2ec = tx_len;
		tx_busy = tx_len != 0;
		USB_IER |= USB_IE_BULK_IN;
		usb_ier_shadow = USB_IER | 10;
	} else {
		tx_busy = tx_len != 0;
		sci_tx_idx = 0;
		SCI0_SCR |= 0x84;               /* TIE | TE */
	}
}

/* ---- main loop: 0xcbf2 -------------------------------------------------- */

/*
 * Stand-ins for the long if/switch chains of the original (see the
 * ponytail note at the end of this file). They do not exist in ROM.
 */
void service_command(int c);
int  single_char_command(int c);
int  two_char_command(int c1c2);
int  single_char_fallback(int c);

void gpgl_loop(void)
{
	int c, c2;

	char landscape = eeprom_read_byte(0x300037);    /* stored orientation (FN) */

	if (landscape)
		set_orientation(landscape);
	*(char *)0xffb358 = 0;
	*(char *)0xffb3ae = 0;

	for (;;) {
		c = rx_getc();
		if (c == ESC || c == FS) {      /* escape pair inside the data stream */
			rx_getc();
			continue;
		}
		if (service_mode && c >= 'a' && c <= 'z') {
			service_command(c);     /* a..z, see notes/dispatch.txt */
			continue;
		}
		last_cmd = c;
		while (c == 0x80)
			c = rx_getc();
		if (c > 0x5f)
			set_error(1);
		if (cmd_class(c) < 0 || c == '#')
			continue;

		if (single_char_command(c))     /* '?','@','A','C','G','H','U','V', '!'..'_' switch */
			continue;

		c2 = rx_peek();                 /* two-letter commands */
		if (two_char_command(c << 8 | c2))      /* consumes c2 when matched */
			continue;
		if (!single_char_fallback(c))   /* '(' 'D' 'E' 'L' 'M' 'O' 'R' 'S' 'T' '^' 'B' 'F' 'N' 'Z' */
			set_error(1);
	}
}

/*
 * ponytail: the three dispatch helpers above stand for the long if/switch
 * chains in the original; the full handler map is in notes/dispatch.txt.
 */
