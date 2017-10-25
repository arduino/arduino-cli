/*
 * This file is part of the libserialport project.
 *
 * Copyright (C) 2013 Martin Ling <martin-libserialport@earth.li>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

/*
 * At the time of writing, glibc does not support the Linux kernel interfaces
 * for setting non-standard baud rates and flow control. We therefore have to
 * prepare the correct ioctls ourselves, for which we need the declarations in
 * linux/termios.h.
 *
 * We can't include linux/termios.h in serialport.c however, because its
 * contents conflict with the termios.h provided by glibc. So this file exists
 * to isolate the bits of code which use the kernel termios declarations.
 *
 * The details vary by architecture. Some architectures have c_ispeed/c_ospeed
 * in struct termios, accessed with TCSETS/TCGETS. Others have these fields in
 * struct termios2, accessed with TCSETS2/TCGETS2. Some architectures have the
 * TCSETX/TCGETX ioctls used with struct termiox, others do not.
 */

#ifdef __linux__

#include <stdlib.h>
#include <linux/termios.h>
#include "linux_termios.h"

SP_PRIV unsigned long get_termios_get_ioctl(void)
{
#ifdef HAVE_TERMIOS2
	return TCGETS2;
#else
	return TCGETS;
#endif
}

SP_PRIV unsigned long get_termios_set_ioctl(void)
{
#ifdef HAVE_TERMIOS2
	return TCSETS2;
#else
	return TCSETS;
#endif
}

SP_PRIV size_t get_termios_size(void)
{
#ifdef HAVE_TERMIOS2
	return sizeof(struct termios2);
#else
	return sizeof(struct termios);
#endif
}

#if (defined(HAVE_TERMIOS_SPEED) || defined(HAVE_TERMIOS2_SPEED)) && defined(HAVE_BOTHER)
SP_PRIV int get_termios_speed(void *data)
{
#ifdef HAVE_TERMIOS2
	struct termios2 *term = (struct termios2 *) data;
#else
	struct termios *term = (struct termios *) data;
#endif
	if (term->c_ispeed != term->c_ospeed)
		return -1;
	else
		return term->c_ispeed;
}

SP_PRIV void set_termios_speed(void *data, int speed)
{
#ifdef HAVE_TERMIOS2
	struct termios2 *term = (struct termios2 *) data;
#else
	struct termios *term = (struct termios *) data;
#endif
	term->c_cflag &= ~CBAUD;
	term->c_cflag |= BOTHER;
	term->c_ispeed = term->c_ospeed = speed;
}
#endif

#ifdef HAVE_TERMIOX
SP_PRIV size_t get_termiox_size(void)
{
	return sizeof(struct termiox);
}

SP_PRIV int get_termiox_flow(void *data, int *rts, int *cts, int *dtr, int *dsr)
{
	struct termiox *termx = (struct termiox *) data;
	int flags = 0;

	*rts = (termx->x_cflag & RTSXOFF);
	*cts = (termx->x_cflag & CTSXON);
	*dtr = (termx->x_cflag & DTRXOFF);
	*dsr = (termx->x_cflag & DSRXON);

	return flags;
}

SP_PRIV void set_termiox_flow(void *data, int rts, int cts, int dtr, int dsr)
{
	struct termiox *termx = (struct termiox *) data;

	termx->x_cflag &= ~(RTSXOFF | CTSXON | DTRXOFF | DSRXON);

	if (rts)
		termx->x_cflag |= RTSXOFF;
	if (cts)
		termx->x_cflag |= CTSXON;
	if (dtr)
		termx->x_cflag |= DTRXOFF;
	if (dsr)
		termx->x_cflag |= DSRXON;
}
#endif

#endif
