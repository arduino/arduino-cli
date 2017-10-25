/*
 * This file is part of the libserialport project.
 *
 * Copyright (C) 2013-2014 Martin Ling <martin-libserialport@earth.li>
 * Copyright (C) 2014 Aurelien Jacobs <aurel@gnuage.org>
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

#ifdef __APPLE__

#include "libserialport.h"
#include "libserialport_internal.h"

SP_PRIV enum sp_return get_port_details(struct sp_port *port)
{
	/* Description limited to 127 char,
	   anything longer would not be user friendly anyway */
	char description[128];
	int bus, address, vid, pid = -1;
	char manufacturer[128], product[128], serial[128];
	CFMutableDictionaryRef classes;
	io_iterator_t iter;
	io_object_t ioport, ioparent;
	CFTypeRef cf_property, cf_bus, cf_address, cf_vendor, cf_product;
	Boolean result;
	char path[PATH_MAX], class[16];

	DEBUG("Getting serial port list");
	if (!(classes = IOServiceMatching(kIOSerialBSDServiceValue)))
		RETURN_FAIL("IOServiceMatching() failed");

	if (IOServiceGetMatchingServices(kIOMasterPortDefault, classes,
	                                 &iter) != KERN_SUCCESS)
		RETURN_FAIL("IOServiceGetMatchingServices() failed");

	DEBUG("Iterating over results");
	while ((ioport = IOIteratorNext(iter))) {
		if (!(cf_property = IORegistryEntryCreateCFProperty(ioport,
		            CFSTR(kIOCalloutDeviceKey), kCFAllocatorDefault, 0))) {
			IOObjectRelease(ioport);
			continue;
		}
		result = CFStringGetCString(cf_property, path, sizeof(path),
		                            kCFStringEncodingASCII);
		CFRelease(cf_property);
		if (!result || strcmp(path, port->name)) {
			IOObjectRelease(ioport);
			continue;
		}
		DEBUG_FMT("Found port %s", path);

		IORegistryEntryGetParentEntry(ioport, kIOServicePlane, &ioparent);
		if ((cf_property=IORegistryEntrySearchCFProperty(ioparent,kIOServicePlane,
		           CFSTR("IOProviderClass"), kCFAllocatorDefault,
		           kIORegistryIterateRecursively | kIORegistryIterateParents))) {
			if (CFStringGetCString(cf_property, class, sizeof(class),
			                       kCFStringEncodingASCII) &&
			    strstr(class, "USB")) {
				DEBUG("Found USB class device");
				port->transport = SP_TRANSPORT_USB;
			}
			CFRelease(cf_property);
		}
		if ((cf_property=IORegistryEntrySearchCFProperty(ioparent,kIOServicePlane,
		           CFSTR("IOClass"), kCFAllocatorDefault,
		           kIORegistryIterateRecursively | kIORegistryIterateParents))) {
			if (CFStringGetCString(cf_property, class, sizeof(class),
			                       kCFStringEncodingASCII) &&
			    strstr(class, "USB")) {
				DEBUG("Found USB class device");
				port->transport = SP_TRANSPORT_USB;
			}
			CFRelease(cf_property);
		}
		IOObjectRelease(ioparent);

		if ((cf_property = IORegistryEntrySearchCFProperty(ioport,kIOServicePlane,
		         CFSTR("USB Interface Name"), kCFAllocatorDefault,
		         kIORegistryIterateRecursively | kIORegistryIterateParents)) ||
		    (cf_property = IORegistryEntrySearchCFProperty(ioport,kIOServicePlane,
		         CFSTR("USB Product Name"), kCFAllocatorDefault,
		         kIORegistryIterateRecursively | kIORegistryIterateParents)) ||
		    (cf_property = IORegistryEntrySearchCFProperty(ioport,kIOServicePlane,
		         CFSTR("Product Name"), kCFAllocatorDefault,
		         kIORegistryIterateRecursively | kIORegistryIterateParents)) ||
		    (cf_property = IORegistryEntryCreateCFProperty(ioport,
		         CFSTR(kIOTTYDeviceKey), kCFAllocatorDefault, 0))) {
			if (CFStringGetCString(cf_property, description, sizeof(description),
			                       kCFStringEncodingASCII)) {
				DEBUG_FMT("Found description %s", description);
				port->description = strdup(description);
			}
			CFRelease(cf_property);
		} else {
			DEBUG("No description for this device");
		}

		cf_bus = IORegistryEntrySearchCFProperty(ioport, kIOServicePlane,
		                                         CFSTR("USBBusNumber"),
		                                         kCFAllocatorDefault,
		                                         kIORegistryIterateRecursively
		                                         | kIORegistryIterateParents);
		cf_address = IORegistryEntrySearchCFProperty(ioport, kIOServicePlane,
		                                         CFSTR("USB Address"),
		                                         kCFAllocatorDefault,
		                                         kIORegistryIterateRecursively
		                                         | kIORegistryIterateParents);
		if (cf_bus && cf_address &&
		    CFNumberGetValue(cf_bus    , kCFNumberIntType, &bus) &&
		    CFNumberGetValue(cf_address, kCFNumberIntType, &address)) {
			DEBUG_FMT("Found matching USB bus:address %03d:%03d", bus, address);
			port->usb_bus = bus;
			port->usb_address = address;
		}
		if (cf_bus    )  CFRelease(cf_bus);
		if (cf_address)  CFRelease(cf_address);

		cf_vendor = IORegistryEntrySearchCFProperty(ioport, kIOServicePlane,
		                                         CFSTR("idVendor"),
		                                         kCFAllocatorDefault,
		                                         kIORegistryIterateRecursively
		                                         | kIORegistryIterateParents);
		cf_product = IORegistryEntrySearchCFProperty(ioport, kIOServicePlane,
		                                         CFSTR("idProduct"),
		                                         kCFAllocatorDefault,
		                                         kIORegistryIterateRecursively
		                                         | kIORegistryIterateParents);
		if (cf_vendor && cf_product &&
		    CFNumberGetValue(cf_vendor , kCFNumberIntType, &vid) &&
		    CFNumberGetValue(cf_product, kCFNumberIntType, &pid)) {
			DEBUG_FMT("Found matching USB vid:pid %04X:%04X", vid, pid);
			port->usb_vid = vid;
			port->usb_pid = pid;
		}
		if (cf_vendor )  CFRelease(cf_vendor);
		if (cf_product)  CFRelease(cf_product);

		if ((cf_property = IORegistryEntrySearchCFProperty(ioport,kIOServicePlane,
		         CFSTR("USB Vendor Name"), kCFAllocatorDefault,
		         kIORegistryIterateRecursively | kIORegistryIterateParents))) {
			if (CFStringGetCString(cf_property, manufacturer, sizeof(manufacturer),
			                       kCFStringEncodingASCII)) {
				DEBUG_FMT("Found manufacturer %s", manufacturer);
				port->usb_manufacturer = strdup(manufacturer);
			}
			CFRelease(cf_property);
		}

		if ((cf_property = IORegistryEntrySearchCFProperty(ioport,kIOServicePlane,
		         CFSTR("USB Product Name"), kCFAllocatorDefault,
		         kIORegistryIterateRecursively | kIORegistryIterateParents))) {
			if (CFStringGetCString(cf_property, product, sizeof(product),
			                       kCFStringEncodingASCII)) {
				DEBUG_FMT("Found product name %s", product);
				port->usb_product = strdup(product);
			}
			CFRelease(cf_property);
		}

		if ((cf_property = IORegistryEntrySearchCFProperty(ioport,kIOServicePlane,
		         CFSTR("USB Serial Number"), kCFAllocatorDefault,
		         kIORegistryIterateRecursively | kIORegistryIterateParents))) {
			if (CFStringGetCString(cf_property, serial, sizeof(serial),
			                       kCFStringEncodingASCII)) {
				DEBUG_FMT("Found serial number %s", serial);
				port->usb_serial = strdup(serial);
			}
			CFRelease(cf_property);
		}

		IOObjectRelease(ioport);
		break;
	}
	IOObjectRelease(iter);

	RETURN_OK();
}

SP_PRIV enum sp_return list_ports(struct sp_port ***list)
{
	CFMutableDictionaryRef classes;
	io_iterator_t iter;
	char path[PATH_MAX];
	io_object_t port;
	CFTypeRef cf_path;
	Boolean result;
	int ret = SP_OK;

	DEBUG("Creating matching dictionary");
	if (!(classes = IOServiceMatching(kIOSerialBSDServiceValue))) {
		SET_FAIL(ret, "IOServiceMatching() failed");
		goto out_done;
	}

	DEBUG("Getting matching services");
	if (IOServiceGetMatchingServices(kIOMasterPortDefault, classes,
	                                 &iter) != KERN_SUCCESS) {
		SET_FAIL(ret, "IOServiceGetMatchingServices() failed");
		goto out_done;
	}

	DEBUG("Iterating over results");
	while ((port = IOIteratorNext(iter))) {
		cf_path = IORegistryEntryCreateCFProperty(port,
				CFSTR(kIOCalloutDeviceKey), kCFAllocatorDefault, 0);
		if (cf_path) {
			result = CFStringGetCString(cf_path, path, sizeof(path),
			                            kCFStringEncodingASCII);
			CFRelease(cf_path);
			if (result) {
				DEBUG_FMT("Found port %s", path);
				if (!(*list = list_append(*list, path))) {
					SET_ERROR(ret, SP_ERR_MEM, "list append failed");
					IOObjectRelease(port);
					goto out;
				}
			}
		}
		IOObjectRelease(port);
	}
out:
	IOObjectRelease(iter);
out_done:

	return ret;
}

#endif
