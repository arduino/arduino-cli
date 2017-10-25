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

#ifdef __linux__

#ifndef LIBSERIALPORT_LINUX_TERMIOS_H
#define LIBSERIALPORT_LINUX_TERMIOS_H

#include <stdlib.h>

SP_PRIV unsigned long get_termios_get_ioctl(void);
SP_PRIV unsigned long get_termios_set_ioctl(void);
SP_PRIV size_t get_termios_size(void);
SP_PRIV int get_termios_speed(void *data);
SP_PRIV void set_termios_speed(void *data, int speed);
SP_PRIV size_t get_termiox_size(void);
SP_PRIV int get_termiox_flow(void *data, int *rts, int *cts, int *dtr, int *dsr);
SP_PRIV void set_termiox_flow(void *data, int rts, int cts, int dtr, int dsr);

#endif

#endif
