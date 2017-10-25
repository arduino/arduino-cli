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

/**
 * @mainpage libserialport API
 *
 * Introduction
 * ============
 *
 * libserialport is a minimal library written in C that is intended to take
 * care of the OS-specific details when writing software that uses serial ports.
 *
 * By writing your serial code to use libserialport, you enable it to work
 * transparently on any platform supported by the library.
 *
 * The operations that are supported are:
 *
 * - @ref Enumeration (obtaining a list of serial ports on the system)
 * - @ref Ports
 * - @ref Configuration (baud rate, parity, etc.)
 * - @ref Signals (modem control lines, breaks, etc.)
 * - @ref Data
 * - @ref Waiting
 * - @ref Errors
 *
 * libserialport is an open source project released under the LGPL3+ license.
 *
 * API principles
 * ==============
 *
 * The API is simple, and designed to be a minimal wrapper around the serial
 * port support in each OS.
 *
 * Most functions take a pointer to a struct sp_port, which represents a serial
 * port. These structures are always allocated and freed by the library, using
 * the functions in the @ref Enumeration "Enumeration" section.
 *
 * Most functions have return type @ref sp_return and can return only four
 * possible error values:
 *
 * - @ref SP_ERR_ARG means that a function was called with invalid
 *   arguments. This implies a bug in the caller. The arguments passed would
 *   be invalid regardless of the underlying OS or serial device involved.
 *
 * - @ref SP_ERR_FAIL means that the OS reported a failure. The error code or
 *   message provided by the OS can be obtained by calling sp_last_error_code()
 *   or sp_last_error_message().
 *
 * - @ref SP_ERR_SUPP indicates that there is no support for the requested
 *   operation in the current OS, driver or device. No error message is
 *   available from the OS in this case. There is either no way to request
 *   the operation in the first place, or libserialport does not know how to
 *   do so in the current version.
 *
 * - @ref SP_ERR_MEM indicates that a memory allocation failed.
 *
 * All of these error values are negative.
 *
 * Calls that succeed return @ref SP_OK, which is equal to zero. Some functions
 * declared @ref sp_return can also return a positive value for a successful
 * numeric result, e.g. sp_blocking_read() or sp_blocking_write().
 */

#ifndef LIBSERIALPORT_LIBSERIALPORT_H
#define LIBSERIALPORT_LIBSERIALPORT_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stddef.h>
#ifdef _WIN32
#include <windows.h>
#endif

/** Return values. */
enum sp_return {
	/** Operation completed successfully. */
	SP_OK = 0,
	/** Invalid arguments were passed to the function. */
	SP_ERR_ARG = -1,
	/** A system error occured while executing the operation. */
	SP_ERR_FAIL = -2,
	/** A memory allocation failed while executing the operation. */
	SP_ERR_MEM = -3,
	/** The requested operation is not supported by this system or device. */
	SP_ERR_SUPP = -4
};

/** Port access modes. */
enum sp_mode {
	/** Open port for read access. */
	SP_MODE_READ = 1,
	/** Open port for write access. */
	SP_MODE_WRITE = 2,
	/** Open port for read and write access. */
	SP_MODE_READ_WRITE = 3
};

/** Port events. */
enum sp_event {
	/* Data received and ready to read. */
	SP_EVENT_RX_READY = 1,
	/* Ready to transmit new data. */
	SP_EVENT_TX_READY = 2,
	/* Error occured. */
	SP_EVENT_ERROR = 4
};

/** Buffer selection. */
enum sp_buffer {
	/** Input buffer. */
	SP_BUF_INPUT = 1,
	/** Output buffer. */
	SP_BUF_OUTPUT = 2,
	/** Both buffers. */
	SP_BUF_BOTH = 3
};

/** Parity settings. */
enum sp_parity {
	/** Special value to indicate setting should be left alone. */
	SP_PARITY_INVALID = -1,
	/** No parity. */
	SP_PARITY_NONE = 0,
	/** Odd parity. */
	SP_PARITY_ODD = 1,
	/** Even parity. */
	SP_PARITY_EVEN = 2,
	/** Mark parity. */
	SP_PARITY_MARK = 3,
	/** Space parity. */
	SP_PARITY_SPACE = 4
};

/** RTS pin behaviour. */
enum sp_rts {
	/** Special value to indicate setting should be left alone. */
	SP_RTS_INVALID = -1,
	/** RTS off. */
	SP_RTS_OFF = 0,
	/** RTS on. */
	SP_RTS_ON = 1,
	/** RTS used for flow control. */
	SP_RTS_FLOW_CONTROL = 2
};

/** CTS pin behaviour. */
enum sp_cts {
	/** Special value to indicate setting should be left alone. */
	SP_CTS_INVALID = -1,
	/** CTS ignored. */
	SP_CTS_IGNORE = 0,
	/** CTS used for flow control. */
	SP_CTS_FLOW_CONTROL = 1
};

/** DTR pin behaviour. */
enum sp_dtr {
	/** Special value to indicate setting should be left alone. */
	SP_DTR_INVALID = -1,
	/** DTR off. */
	SP_DTR_OFF = 0,
	/** DTR on. */
	SP_DTR_ON = 1,
	/** DTR used for flow control. */
	SP_DTR_FLOW_CONTROL = 2
};

/** DSR pin behaviour. */
enum sp_dsr {
	/** Special value to indicate setting should be left alone. */
	SP_DSR_INVALID = -1,
	/** DSR ignored. */
	SP_DSR_IGNORE = 0,
	/** DSR used for flow control. */
	SP_DSR_FLOW_CONTROL = 1
};

/** XON/XOFF flow control behaviour. */
enum sp_xonxoff {
	/** Special value to indicate setting should be left alone. */
	SP_XONXOFF_INVALID = -1,
	/** XON/XOFF disabled. */
	SP_XONXOFF_DISABLED = 0,
	/** XON/XOFF enabled for input only. */
	SP_XONXOFF_IN = 1,
	/** XON/XOFF enabled for output only. */
	SP_XONXOFF_OUT = 2,
	/** XON/XOFF enabled for input and output. */
	SP_XONXOFF_INOUT = 3
};

/** Standard flow control combinations. */
enum sp_flowcontrol {
	/** No flow control. */
	SP_FLOWCONTROL_NONE = 0,
	/** Software flow control using XON/XOFF characters. */
	SP_FLOWCONTROL_XONXOFF = 1,
	/** Hardware flow control using RTS/CTS signals. */
	SP_FLOWCONTROL_RTSCTS = 2,
	/** Hardware flow control using DTR/DSR signals. */
	SP_FLOWCONTROL_DTRDSR = 3
};

/** Input signals. */
enum sp_signal {
	/** Clear to send. */
	SP_SIG_CTS = 1,
	/** Data set ready. */
	SP_SIG_DSR = 2,
	/** Data carrier detect. */
	SP_SIG_DCD = 4,
	/** Ring indicator. */
	SP_SIG_RI = 8
};

/** Transport types. */
enum sp_transport {
	/** Native platform serial port. */
	SP_TRANSPORT_NATIVE,
	/** USB serial port adapter. */
	SP_TRANSPORT_USB,
	/** Bluetooth serial port adapter. */
	SP_TRANSPORT_BLUETOOTH
};

/**
 * @struct sp_port
 * An opaque structure representing a serial port.
 */
struct sp_port;

/**
 * @struct sp_port_config
 * An opaque structure representing the configuration for a serial port.
 */
struct sp_port_config;

/**
 * @struct sp_event_set
 * A set of handles to wait on for events.
 */
struct sp_event_set {
	/** Array of OS-specific handles. */
	void *handles;
	/** Array of bitmasks indicating which events apply for each handle. */
	enum sp_event *masks;
	/** Number of handles. */
	unsigned int count;
};

/**
@defgroup Enumeration Port enumeration
@{
*/

/**
 * Obtain a pointer to a new sp_port structure representing the named port.
 *
 * The user should allocate a variable of type "struct sp_port *" and pass a
 * pointer to this to receive the result.
 *
 * The result should be freed after use by calling sp_free_port().
 *
 * If any error is returned, the variable pointed to by port_ptr will be set
 * to NULL. Otherwise, it will be set to point to the newly allocated port.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_get_port_by_name(const char *portname, struct sp_port **port_ptr);

/**
 * Free a port structure obtained from sp_get_port_by_name() or sp_copy_port().
 *
 * @since 0.1.0
 */
void sp_free_port(struct sp_port *port);

/**
 * List the serial ports available on the system.
 *
 * The result obtained is an array of pointers to sp_port structures,
 * terminated by a NULL. The user should allocate a variable of type
 * "struct sp_port **" and pass a pointer to this to receive the result.
 *
 * The result should be freed after use by calling sp_free_port_list().
 * If a port from the list is to be used after freeing the list, it must be
 * copied first using sp_copy_port().
 *
 * If any error is returned, the variable pointed to by list_ptr will be set
 * to NULL. Otherwise, it will be set to point to the newly allocated array.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_list_ports(struct sp_port ***list_ptr);

/**
 * Make a new copy of a sp_port structure.
 *
 * The user should allocate a variable of type "struct sp_port *" and pass a
 * pointer to this to receive the result.
 *
 * The copy should be freed after use by calling sp_free_port().
 *
 * If any error is returned, the variable pointed to by copy_ptr will be set
 * to NULL. Otherwise, it will be set to point to the newly allocated copy.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_copy_port(const struct sp_port *port, struct sp_port **copy_ptr);

/**
 * Free a port list obtained from sp_list_ports().
 *
 * This will also free all the sp_port structures referred to from the list;
 * any that are to be retained must be copied first using sp_copy_port().
 *
 * @since 0.1.0
 */
void sp_free_port_list(struct sp_port **ports);

/**
 * @}
 * @defgroup Ports Opening, closing and querying ports
 * @{
 */

/**
 * Open the specified serial port.
 *
 * @param port Pointer to port structure.
 * @param flags Flags to use when opening the serial port.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_open(struct sp_port *port, enum sp_mode flags);

/**
 * Close the specified serial port.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_close(struct sp_port *port);

/**
 * Get the name of a port.
 *
 * The name returned is whatever is normally used to refer to a port on the
 * current operating system; e.g. for Windows it will usually be a "COMn"
 * device name, and for Unix it will be a device path beginning with "/dev/".
 *
 * @param port Pointer to port structure.
 *
 * @return The port name, or NULL if an invalid port is passed. The name
 * string is part of the port structure and may not be used after the
 * port structure has been freed.
 *
 * @since 0.1.0
 */
char *sp_get_port_name(const struct sp_port *port);

/**
 * Get a description for a port, to present to end user.
 *
 * @param port Pointer to port structure.
 *
 * @return The port description, or NULL if an invalid port is passed.
 * The description string is part of the port structure and may not be used
 * after the port structure has been freed.
 *
 * @since 0.2.0
 */
char *sp_get_port_description(struct sp_port *port);

/**
 * Get the transport type used by a port.
 *
 * @param port Pointer to port structure.
 *
 * @return The port transport type.
 *
 * @since 0.2.0
 */
enum sp_transport sp_get_port_transport(struct sp_port *port);

/**
 * Get the USB bus number and address on bus of a USB serial adapter port.
 *
 * @param port Pointer to port structure.
 * @param usb_bus Pointer to variable to store USB bus.
 * @param usb_address Pointer to variable to store USB address
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.2.0
 */
enum sp_return sp_get_port_usb_bus_address(const struct sp_port *port,
                                           int *usb_bus, int *usb_address);

/**
 * Get the USB Vendor ID and Product ID of a USB serial adapter port.
 *
 * @param port Pointer to port structure.
 * @param usb_vid Pointer to variable to store USB VID.
 * @param usb_pid Pointer to variable to store USB PID
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.2.0
 */
enum sp_return sp_get_port_usb_vid_pid(const struct sp_port *port, int *usb_vid, int *usb_pid);

/**
 * Get the USB manufacturer string of a USB serial adapter port.
 *
 * @param port Pointer to port structure.
 *
 * @return The port manufacturer string, or NULL if an invalid port is passed.
 * The manufacturer string is part of the port structure and may not be used
 * after the port structure has been freed.
 *
 * @since 0.2.0
 */
char *sp_get_port_usb_manufacturer(const struct sp_port *port);

/**
 * Get the USB product string of a USB serial adapter port.
 *
 * @param port Pointer to port structure.
 *
 * @return The port product string, or NULL if an invalid port is passed.
 * The product string is part of the port structure and may not be used
 * after the port structure has been freed.
 *
 * @since 0.2.0
 */
char *sp_get_port_usb_product(const struct sp_port *port);

/**
 * Get the USB serial number string of a USB serial adapter port.
 *
 * @param port Pointer to port structure.
 *
 * @return The port serial number, or NULL if an invalid port is passed.
 * The serial number string is part of the port structure and may not be used
 * after the port structure has been freed.
 *
 * @since 0.2.0
 */
char *sp_get_port_usb_serial(const struct sp_port *port);

/**
 * Get the MAC address of a Bluetooth serial adapter port.
 *
 * @param port Pointer to port structure.
 *
 * @return The port MAC address, or NULL if an invalid port is passed.
 * The MAC address string is part of the port structure and may not be used
 * after the port structure has been freed.
 *
 * @since 0.2.0
 */
char *sp_get_port_bluetooth_address(const struct sp_port *port);

/**
 * Get the operating system handle for a port.
 *
 * The type of the handle depends on the operating system. On Unix based
 * systems, the handle is a file descriptor of type "int". On Windows, the
 * handle is of type "HANDLE". The user should allocate a variable of the
 * appropriate type and pass a pointer to this to receive the result.
 *
 * To obtain a valid handle, the port must first be opened by calling
 * sp_open() using the same port structure.
 *
 * After the port is closed or the port structure freed, the handle may
 * no longer be valid.
 *
 * @warning This feature is provided so that programs may make use of
 *          OS-specific functionality where desired. Doing so obviously
 *          comes at a cost in portability. It also cannot be guaranteed
 *          that direct usage of the OS handle will not conflict with the
 *          library's own usage of the port. Be careful.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_get_port_handle(const struct sp_port *port, void *result_ptr);

/**
 * @}
 * @defgroup Configuration Setting port parameters
 * @{
 */

/**
 * Allocate a port configuration structure.
 *
 * The user should allocate a variable of type "struct sp_config *" and pass a
 * pointer to this to receive the result. The variable will be updated to
 * point to the new configuration structure. The structure is opaque and must
 * be accessed via the functions provided.
 *
 * All parameters in the structure will be initialised to special values which
 * are ignored by sp_set_config().
 *
 * The structure should be freed after use by calling sp_free_config().
 *
 * @param config_ptr Pointer to variable to receive result.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_new_config(struct sp_port_config **config_ptr);

/**
 * Free a port configuration structure.
 *
 * @param config Pointer to configuration structure.
 *
 * @since 0.1.0
 */
void sp_free_config(struct sp_port_config *config);

/**
 * Get the current configuration of the specified serial port.
 *
 * The user should allocate a configuration structure using sp_new_config()
 * and pass this as the config parameter. The configuration structure will
 * be updated with the port configuration.
 *
 * Any parameters that are configured with settings not recognised or
 * supported by libserialport, will be set to special values that are
 * ignored by sp_set_config().
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_get_config(struct sp_port *port, struct sp_port_config *config);

/**
 * Set the configuration for the specified serial port.
 *
 * For each parameter in the configuration, there is a special value (usually
 * -1, but see the documentation for each field). These values will be ignored
 * and the corresponding setting left unchanged on the port.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_config(struct sp_port *port, const struct sp_port_config *config);

/**
 * Set the baud rate for the specified serial port.
 *
 * @param port Pointer to port structure.
 * @param baudrate Baud rate in bits per second.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_baudrate(struct sp_port *port, int baudrate);

/**
 * Get the baud rate from a port configuration.
 *
 * The user should allocate a variable of type int and pass a pointer to this
 * to receive the result.
 *
 * @param config Pointer to configuration structure.
 * @param baudrate_ptr Pointer to variable to store result.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_get_config_baudrate(const struct sp_port_config *config, int *baudrate_ptr);

/**
 * Set the baud rate in a port configuration.
 *
 * @param config Pointer to configuration structure.
 * @param baudrate Baud rate in bits per second, or -1 to retain current setting.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_config_baudrate(struct sp_port_config *config, int baudrate);

/**
 * Set the data bits for the specified serial port.
 *
 * @param port Pointer to port structure.
 * @param bits Number of data bits.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_bits(struct sp_port *port, int bits);

/**
 * Get the data bits from a port configuration.
 *
 * The user should allocate a variable of type int and pass a pointer to this
 * to receive the result.
 *
 * @param config Pointer to configuration structure.
 * @param bits_ptr Pointer to variable to store result.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_get_config_bits(const struct sp_port_config *config, int *bits_ptr);

/**
 * Set the data bits in a port configuration.
 *
 * @param config Pointer to configuration structure.
 * @param bits Number of data bits, or -1 to retain current setting.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_config_bits(struct sp_port_config *config, int bits);

/**
 * Set the parity setting for the specified serial port.
 *
 * @param port Pointer to port structure.
 * @param parity Parity setting.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_parity(struct sp_port *port, enum sp_parity parity);

/**
 * Get the parity setting from a port configuration.
 *
 * The user should allocate a variable of type enum sp_parity and pass a pointer to this
 * to receive the result.
 *
 * @param config Pointer to configuration structure.
 * @param parity_ptr Pointer to variable to store result.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_get_config_parity(const struct sp_port_config *config, enum sp_parity *parity_ptr);

/**
 * Set the parity setting in a port configuration.
 *
 * @param config Pointer to configuration structure.
 * @param parity Parity setting, or -1 to retain current setting.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_config_parity(struct sp_port_config *config, enum sp_parity parity);

/**
 * Set the stop bits for the specified serial port.
 *
 * @param port Pointer to port structure.
 * @param stopbits Number of stop bits.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_stopbits(struct sp_port *port, int stopbits);

/**
 * Get the stop bits from a port configuration.
 *
 * The user should allocate a variable of type int and pass a pointer to this
 * to receive the result.
 *
 * @param config Pointer to configuration structure.
 * @param stopbits_ptr Pointer to variable to store result.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_get_config_stopbits(const struct sp_port_config *config, int *stopbits_ptr);

/**
 * Set the stop bits in a port configuration.
 *
 * @param config Pointer to configuration structure.
 * @param stopbits Number of stop bits, or -1 to retain current setting.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_config_stopbits(struct sp_port_config *config, int stopbits);

/**
 * Set the RTS pin behaviour for the specified serial port.
 *
 * @param port Pointer to port structure.
 * @param rts RTS pin mode.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_rts(struct sp_port *port, enum sp_rts rts);

/**
 * Get the RTS pin behaviour from a port configuration.
 *
 * The user should allocate a variable of type enum sp_rts and pass a pointer to this
 * to receive the result.
 *
 * @param config Pointer to configuration structure.
 * @param rts_ptr Pointer to variable to store result.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_get_config_rts(const struct sp_port_config *config, enum sp_rts *rts_ptr);

/**
 * Set the RTS pin behaviour in a port configuration.
 *
 * @param config Pointer to configuration structure.
 * @param rts RTS pin mode, or -1 to retain current setting.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_config_rts(struct sp_port_config *config, enum sp_rts rts);

/**
 * Set the CTS pin behaviour for the specified serial port.
 *
 * @param port Pointer to port structure.
 * @param cts CTS pin mode.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_cts(struct sp_port *port, enum sp_cts cts);

/**
 * Get the CTS pin behaviour from a port configuration.
 *
 * The user should allocate a variable of type enum sp_cts and pass a pointer to this
 * to receive the result.
 *
 * @param config Pointer to configuration structure.
 * @param cts_ptr Pointer to variable to store result.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_get_config_cts(const struct sp_port_config *config, enum sp_cts *cts_ptr);

/**
 * Set the CTS pin behaviour in a port configuration.
 *
 * @param config Pointer to configuration structure.
 * @param cts CTS pin mode, or -1 to retain current setting.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_config_cts(struct sp_port_config *config, enum sp_cts cts);

/**
 * Set the DTR pin behaviour for the specified serial port.
 *
 * @param port Pointer to port structure.
 * @param dtr DTR pin mode.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_dtr(struct sp_port *port, enum sp_dtr dtr);

/**
 * Get the DTR pin behaviour from a port configuration.
 *
 * The user should allocate a variable of type enum sp_dtr and pass a pointer to this
 * to receive the result.
 *
 * @param config Pointer to configuration structure.
 * @param dtr_ptr Pointer to variable to store result.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_get_config_dtr(const struct sp_port_config *config, enum sp_dtr *dtr_ptr);

/**
 * Set the DTR pin behaviour in a port configuration.
 *
 * @param config Pointer to configuration structure.
 * @param dtr DTR pin mode, or -1 to retain current setting.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_config_dtr(struct sp_port_config *config, enum sp_dtr dtr);

/**
 * Set the DSR pin behaviour for the specified serial port.
 *
 * @param port Pointer to port structure.
 * @param dsr DSR pin mode.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_dsr(struct sp_port *port, enum sp_dsr dsr);

/**
 * Get the DSR pin behaviour from a port configuration.
 *
 * The user should allocate a variable of type enum sp_dsr and pass a pointer to this
 * to receive the result.
 *
 * @param config Pointer to configuration structure.
 * @param dsr_ptr Pointer to variable to store result.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_get_config_dsr(const struct sp_port_config *config, enum sp_dsr *dsr_ptr);

/**
 * Set the DSR pin behaviour in a port configuration.
 *
 * @param config Pointer to configuration structure.
 * @param dsr DSR pin mode, or -1 to retain current setting.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_config_dsr(struct sp_port_config *config, enum sp_dsr dsr);

/**
 * Set the XON/XOFF configuration for the specified serial port.
 *
 * @param port Pointer to port structure.
 * @param xon_xoff XON/XOFF mode.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_xon_xoff(struct sp_port *port, enum sp_xonxoff xon_xoff);

/**
 * Get the XON/XOFF configuration from a port configuration.
 *
 * The user should allocate a variable of type enum sp_xonxoff and pass a pointer to this
 * to receive the result.
 *
 * @param config Pointer to configuration structure.
 * @param xon_xoff_ptr Pointer to variable to store result.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_get_config_xon_xoff(const struct sp_port_config *config, enum sp_xonxoff *xon_xoff_ptr);

/**
 * Set the XON/XOFF configuration in a port configuration.
 *
 * @param config Pointer to configuration structure.
 * @param xon_xoff XON/XOFF mode, or -1 to retain current setting.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_config_xon_xoff(struct sp_port_config *config, enum sp_xonxoff xon_xoff);

/**
 * Set the flow control type in a port configuration.
 *
 * This function is a wrapper that sets the RTS, CTS, DTR, DSR and
 * XON/XOFF settings as necessary for the specified flow control
 * type. For more fine-grained control of these settings, use their
 * individual configuration functions.
 *
 * @param config Pointer to configuration structure.
 * @param flowcontrol Flow control setting to use.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_config_flowcontrol(struct sp_port_config *config, enum sp_flowcontrol flowcontrol);

/**
 * Set the flow control type for the specified serial port.
 *
 * This function is a wrapper that sets the RTS, CTS, DTR, DSR and
 * XON/XOFF settings as necessary for the specified flow control
 * type. For more fine-grained control of these settings, use their
 * individual configuration functions.
 *
 * @param port Pointer to port structure.
 * @param flowcontrol Flow control setting to use.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_set_flowcontrol(struct sp_port *port, enum sp_flowcontrol flowcontrol);

/**
 * @}
 * @defgroup Data Reading, writing, and flushing data
 * @{
*/

/**
 * Read bytes from the specified serial port, blocking until complete.
 *
 * @warning If your program runs on Unix, defines its own signal handlers, and
 *          needs to abort blocking reads when these are called, then you
 *          should not use this function. It repeats system calls that return
 *          with EINTR. To be able to abort a read from a signal handler, you
 *          should implement your own blocking read using sp_nonblocking_read()
 *          together with a blocking method that makes sense for your program.
 *          E.g. you can obtain the file descriptor for an open port using
 *          sp_get_port_handle() and use this to call select() or pselect(),
 *          with appropriate arrangements to return if a signal is received.
 *
 * @param port Pointer to port structure.
 * @param buf Buffer in which to store the bytes read.
 * @param count Requested number of bytes to read.
 * @param timeout Timeout in milliseconds, or zero to wait indefinitely.
 *
 * @return The number of bytes read on success, or a negative error code. If
 *         the number of bytes returned is less than that requested, the
 *         timeout was reached before the requested number of bytes was
 *         available. If timeout is zero, the function will always return
 *         either the requested number of bytes or a negative error code.
 *
 * @since 0.1.0
 */
enum sp_return sp_blocking_read(struct sp_port *port, void *buf, size_t count, unsigned int timeout);

/**
 * Read bytes from the specified serial port, without blocking.
 *
 * @param port Pointer to port structure.
 * @param buf Buffer in which to store the bytes read.
 * @param count Maximum number of bytes to read.
 *
 * @return The number of bytes read on success, or a negative error code. The
 *         number of bytes returned may be any number from zero to the maximum
 *         that was requested.
 *
 * @since 0.1.0
 */
enum sp_return sp_nonblocking_read(struct sp_port *port, void *buf, size_t count);

/**
 * Write bytes to the specified serial port, blocking until complete.
 *
 * Note that this function only ensures that the accepted bytes have been
 * written to the OS; they may be held in driver or hardware buffers and not
 * yet physically transmitted. To check whether all written bytes have actually
 * been transmitted, use the sp_output_waiting() function. To wait until all
 * written bytes have actually been transmitted, use the sp_drain() function.
 *
 * @warning If your program runs on Unix, defines its own signal handlers, and
 *          needs to abort blocking writes when these are called, then you
 *          should not use this function. It repeats system calls that return
 *          with EINTR. To be able to abort a write from a signal handler, you
 *          should implement your own blocking write using sp_nonblocking_write()
 *          together with a blocking method that makes sense for your program.
 *          E.g. you can obtain the file descriptor for an open port using
 *          sp_get_port_handle() and use this to call select() or pselect(),
 *          with appropriate arrangements to return if a signal is received.
 *
 * @param port Pointer to port structure.
 * @param buf Buffer containing the bytes to write.
 * @param count Requested number of bytes to write.
 * @param timeout Timeout in milliseconds, or zero to wait indefinitely.
 *
 * @return The number of bytes written on success, or a negative error code.
 *         If the number of bytes returned is less than that requested, the
 *         timeout was reached before the requested number of bytes was
 *         written. If timeout is zero, the function will always return
 *         either the requested number of bytes or a negative error code. In
 *         the event of an error there is no way to determine how many bytes
 *         were sent before the error occured.
 *
 * @since 0.1.0
 */
enum sp_return sp_blocking_write(struct sp_port *port, const void *buf, size_t count, unsigned int timeout);

/**
 * Write bytes to the specified serial port, without blocking.
 *
 * Note that this function only ensures that the accepted bytes have been
 * written to the OS; they may be held in driver or hardware buffers and not
 * yet physically transmitted. To check whether all written bytes have actually
 * been transmitted, use the sp_output_waiting() function. To wait until all
 * written bytes have actually been transmitted, use the sp_drain() function.
 *
 * @param port Pointer to port structure.
 * @param buf Buffer containing the bytes to write.
 * @param count Maximum number of bytes to write.
 *
 * @return The number of bytes written on success, or a negative error code.
 *         The number of bytes returned may be any number from zero to the
 *         maximum that was requested.
 *
 * @since 0.1.0
 */
enum sp_return sp_nonblocking_write(struct sp_port *port, const void *buf, size_t count);

/**
 * Gets the number of bytes waiting in the input buffer.
 *
 * @param port Pointer to port structure.
 *
 * @return Number of bytes waiting on success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_input_waiting(struct sp_port *port);

/**
 * Gets the number of bytes waiting in the output buffer.
 *
 * @param port Pointer to port structure.
 *
 * @return Number of bytes waiting on success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_output_waiting(struct sp_port *port);

/**
 * Flush serial port buffers. Data in the selected buffer(s) is discarded.
 *
 * @param port Pointer to port structure.
 * @param buffers Which buffer(s) to flush.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_flush(struct sp_port *port, enum sp_buffer buffers);

/**
 * Wait for buffered data to be transmitted.
 *
 * @warning If your program runs on Unix, defines its own signal handlers, and
 *          needs to abort draining the output buffer when when these are
 *          called, then you should not use this function. It repeats system
 *          calls that return with EINTR. To be able to abort a drain from a
 *          signal handler, you would need to implement your own blocking
 *          drain by polling the result of sp_output_waiting().
 *
 * @param port Pointer to port structure.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_drain(struct sp_port *port);

/**
 * @}
 * @defgroup Waiting Waiting for events
 * @{
 */

/**
 * Allocate storage for a set of events.
 *
 * The user should allocate a variable of type struct sp_event_set *,
 * then pass a pointer to this variable to receive the result.
 *
 * The result should be freed after use by calling sp_free_event_set().
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_new_event_set(struct sp_event_set **result_ptr);

/**
 * Add events to a struct sp_event_set for a given port.
 *
 * The port must first be opened by calling sp_open() using the same port
 * structure.
 *
 * After the port is closed or the port structure freed, the results may
 * no longer be valid.
 *
 * @param event_set Event set to update.
 * @param port Pointer to port structure.
 * @param mask Bitmask of events to be waited for.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_add_port_events(struct sp_event_set *event_set,
	const struct sp_port *port, enum sp_event mask);

/**
 * Wait for any of a set of events to occur.
 *
 * @param event_set Event set to wait on.
 * @param timeout Timeout in milliseconds, or zero to wait indefinitely.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_wait(struct sp_event_set *event_set, unsigned int timeout);

/**
 * Free a structure allocated by sp_new_event_set().
 *
 * @since 0.1.0
 */
void sp_free_event_set(struct sp_event_set *event_set);

/**
 * @}
 * @defgroup Signals Port signalling operations
 * @{
 */

/**
 * Gets the status of the control signals for the specified port.
 *
 * The user should allocate a variable of type "enum sp_signal" and pass a
 * pointer to this variable to receive the result. The result is a bitmask
 * in which individual signals can be checked by bitwise OR with values of
 * the sp_signal enum.
 *
 * @param port Pointer to port structure.
 * @param signal_mask Pointer to variable to receive result.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_get_signals(struct sp_port *port, enum sp_signal *signal_mask);

/**
 * Put the port transmit line into the break state.
 *
 * @param port Pointer to port structure.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_start_break(struct sp_port *port);

/**
 * Take the port transmit line out of the break state.
 *
 * @param port Pointer to port structure.
 *
 * @return SP_OK upon success, a negative error code otherwise.
 *
 * @since 0.1.0
 */
enum sp_return sp_end_break(struct sp_port *port);

/**
 * @}
 * @defgroup Errors Obtaining error information
 * @{
*/

/**
 * Get the error code for a failed operation.
 *
 * In order to obtain the correct result, this function should be called
 * straight after the failure, before executing any other system operations.
 *
 * @return The system's numeric code for the error that caused the last
 *         operation to fail.
 *
 * @since 0.1.0
 */
int sp_last_error_code(void);

/**
 * Get the error message for a failed operation.
 *
 * In order to obtain the correct result, this function should be called
 * straight after the failure, before executing other system operations.
 *
 * @return The system's message for the error that caused the last
 *         operation to fail. This string may be allocated by the function,
 *         and should be freed after use by calling sp_free_error_message().
 *
 * @since 0.1.0
 */
char *sp_last_error_message(void);

/**
 * Free an error message returned by sp_last_error_message().
 *
 * @since 0.1.0
 */
void sp_free_error_message(char *message);

/**
 * Set the handler function for library debugging messages.
 *
 * Debugging messages are generated by the library during each operation,
 * to help in diagnosing problems. The handler will be called for each
 * message. The handler can be set to NULL to ignore all debug messages.
 *
 * The handler function should accept a format string and variable length
 * argument list, in the same manner as e.g. printf().
 *
 * The default handler is sp_default_debug_handler().
 *
 * @since 0.1.0
 */
void sp_set_debug_handler(void (*handler)(const char *format, ...));

/**
 * Default handler function for library debugging messages.
 *
 * This function prints debug messages to the standard error stream if the
 * environment variable LIBSERIALPORT_DEBUG is set. Otherwise, they are
 * ignored.
 *
 * @since 0.1.0
 */
void sp_default_debug_handler(const char *format, ...);

/** @} */

/**
 * @defgroup Versions Version number querying functions, definitions, and macros
 *
 * This set of API calls returns two different version numbers related
 * to libserialport. The "package version" is the release version number of the
 * libserialport tarball in the usual "major.minor.micro" format, e.g. "0.1.0".
 *
 * The "library version" is independent of that; it is the libtool version
 * number in the "current:revision:age" format, e.g. "2:0:0".
 * See http://www.gnu.org/software/libtool/manual/libtool.html#Libtool-versioning for details.
 *
 * Both version numbers (and/or individual components of them) can be
 * retrieved via the API calls at runtime, and/or they can be checked at
 * compile/preprocessor time using the respective macros.
 *
 * @{
 */

/*
 * Package version macros (can be used for conditional compilation).
 */

/** The libserialport package 'major' version number. */
#define SP_PACKAGE_VERSION_MAJOR 0

/** The libserialport package 'minor' version number. */
#define SP_PACKAGE_VERSION_MINOR 2

/** The libserialport package 'micro' version number. */
#define SP_PACKAGE_VERSION_MICRO 0

/** The libserialport package version ("major.minor.micro") as string. */
#define SP_PACKAGE_VERSION_STRING "0.2.0"

/*
 * Library/libtool version macros (can be used for conditional compilation).
 */

/** The libserialport libtool 'current' version number. */
#define SP_LIB_VERSION_CURRENT 0

/** The libserialport libtool 'revision' version number. */
#define SP_LIB_VERSION_REVISION 0

/** The libserialport libtool 'age' version number. */
#define SP_LIB_VERSION_AGE 0

/** The libserialport libtool version ("current:revision:age") as string. */
#define SP_LIB_VERSION_STRING "0:0:0"

/**
 * Get the major libserialport package version number.
 *
 * @return The major package version number.
 *
 * @since 0.1.0
 */
int sp_get_major_package_version(void);

/**
 * Get the minor libserialport package version number.
 *
 * @return The minor package version number.
 *
 * @since 0.1.0
 */
int sp_get_minor_package_version(void);

/**
 * Get the micro libserialport package version number.
 *
 * @return The micro package version number.
 *
 * @since 0.1.0
 */
int sp_get_micro_package_version(void);

/**
 * Get the libserialport package version number as a string.
 *
 * @return The package version number string. The returned string is
 *         static and thus should NOT be free'd by the caller.
 *
 * @since 0.1.0
 */
const char *sp_get_package_version_string(void);

/**
 * Get the "current" part of the libserialport library version number.
 *
 * @return The "current" library version number.
 *
 * @since 0.1.0
 */
int sp_get_current_lib_version(void);

/**
 * Get the "revision" part of the libserialport library version number.
 *
 * @return The "revision" library version number.
 *
 * @since 0.1.0
 */
int sp_get_revision_lib_version(void);

/**
 * Get the "age" part of the libserialport library version number.
 *
 * @return The "age" library version number.
 *
 * @since 0.1.0
 */
int sp_get_age_lib_version(void);

/**
 * Get the libserialport library version number as a string.
 *
 * @return The library version number string. The returned string is
 *         static and thus should NOT be free'd by the caller.
 *
 * @since 0.1.0
 */
const char *sp_get_lib_version_string(void);

/** @} */

#ifdef __cplusplus
}
#endif

#endif
