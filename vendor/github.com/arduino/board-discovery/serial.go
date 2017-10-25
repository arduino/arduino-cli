/*
 * This file is part of board-discovery.
 *
 * board-discovery is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package discovery

import (
	"fmt"
	"time"

	"github.com/facchinm/go-serial-native"
	"github.com/juju/errors"
)

// Merge updates the device with the new one, returning false if the operation
// didn't change anything
func (d *SerialDevice) merge(dev SerialDevice) bool {
	changed := false
	if d.Port != dev.Port {
		changed = true
		d.Port = dev.Port
	}
	if d.SerialNumber != dev.SerialNumber {
		changed = true
		d.SerialNumber = dev.SerialNumber
	}
	if d.ProductID != dev.ProductID {
		changed = true
		d.ProductID = dev.ProductID
	}
	if d.VendorID != dev.VendorID {
		changed = true
		d.VendorID = dev.VendorID
	}

	if d.Serial != dev.Serial {
		d.Serial = dev.Serial
	}
	return changed
}

func (m *Monitor) serialDiscover() error {
	ports, err := serial.ListPorts()
	if err != nil {
		return errors.Annotatef(err, "while listing the serial ports")
	}

	for _, port := range ports {
		m.addSerial(port)

	}
	m.pruneSerial(ports)

	time.Sleep(m.Interval)
	return nil
}

func (m *Monitor) addSerial(port *serial.Info) {
	vid, pid, _ := port.USBVIDPID()
	if vid == 0 || pid == 0 {
		return
	}

	device := SerialDevice{
		Port:         port.Name(),
		SerialNumber: port.USBSerialNumber(),
		ProductID:    fmt.Sprintf("0x%04X", pid),
		VendorID:     fmt.Sprintf("0x%04X", vid),
		Serial:       port,
	}
	for port, dev := range m.serial {
		if port == device.Port {
			changed := dev.merge(device)
			if changed {
				//m.Events <- Event{Name: "change", SerialDevice: dev}
			}
			return
		}
	}

	m.serial[device.Port] = &device
	//m.Events <- Event{Name: "add", SerialDevice: &device}
}

func (m *Monitor) pruneSerial(ports []*serial.Info) {
	toPrune := []string{}
	for port := range m.serial {
		found := false
		for _, p := range ports {
			if port == p.Name() {
				found = true
			}
		}
		if !found {
			toPrune = append(toPrune, port)
		}
	}

	for _, port := range toPrune {
		//m.Events <- Event{Name: "remove", SerialDevice: m.serial[port]}
		delete(m.serial, port)
	}
}
