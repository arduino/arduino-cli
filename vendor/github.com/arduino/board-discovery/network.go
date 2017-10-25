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
	"net"
	"strings"
	"time"

	"github.com/juju/errors"
	"github.com/oleksandr/bonjour"
)

// Merge updates the device with the new one, returning false if the operation
// didn't change anything
func (d *NetworkDevice) merge(dev NetworkDevice) bool {
	changed := false
	if d.Port != dev.Port {
		changed = true
		d.Port = dev.Port
	}
	if d.Address != dev.Address {
		changed = true
		d.Address = dev.Address
	}
	if d.Info != dev.Info {
		changed = true
		d.Info = dev.Info
	}
	if d.Name != dev.Name {
		changed = true
		d.Name = dev.Name
	}

	return changed
}

// IsUp checks if the device is listening on the given port
// the timeout is 1.5 seconds
// It checks up to the number of times specified. If at least one of the attempt
// is successful it returns true
func (d *NetworkDevice) isUp(times int) bool {
	for i := 0; i < times; i++ {
		if d._isUp() {
			return true
		}
	}
	return false
}

func (d *NetworkDevice) _isUp() bool {
	timeout := time.Duration(1500 * time.Millisecond)
	conn, err := net.DialTimeout("tcp", d.Address+":"+string(d.Port), timeout)
	if err != nil {
		// Check if the port 22 is open
		conn, err = net.DialTimeout("tcp", d.Address+":22", timeout)
		if err != nil {
			return false
		}
		conn.Close()
		return true
	}
	conn.Close()
	return true
}

func (m *Monitor) networkDiscover() error {
	entries, err := listEntries()
	if err != nil {
		return errors.Annotatef(err, "while listing the network ports")
	}

	for _, entry := range entries {
		m.addNetwork(&entry)

	}
	m.pruneNetwork()
	return nil
}

// listEntries returns a list of bonjour entries. It's convoluted because for
// some reason they decided they wanted to make it asynchronous. Seems like writing
// javascript, bleah.
func listEntries() ([]bonjour.ServiceEntry, error) {
	// Define some helper channels that don't have to survive the function
	finished := make(chan bool) // allows us to communicate that the reading has been completed
	errs := make(chan error)    // allows us to communicate that an error has occurred
	defer func() {
		close(finished)
		close(errs)
	}()

	// Instantiate the bonjour controller
	resolver, err := bonjour.NewResolver(nil)
	if err != nil {
		return nil, errors.Annotatef(err, "When initializing the bonjour resolver")
	}

	results := make(chan *bonjour.ServiceEntry)

	// entries is the list of entries we have to return
	entries := []bonjour.ServiceEntry{}

	// Exit if after two seconds there was no response
	go func(exitCh chan<- bool) {
		time.Sleep(4 * time.Second)
		exitCh <- true
		close(results)
	}(resolver.Exit)

	// Loop through the results
	go func(results chan *bonjour.ServiceEntry, exit chan<- bool) {
		for e := range results {
			entries = append(entries, *e)
		}
		finished <- true
	}(results, resolver.Exit)

	// Start the resolving
	err = resolver.Browse("_arduino._tcp", "", results)
	if err != nil {
		close(results)
		errs <- errors.Annotatef(err, "When browsing the services")
	}

	select {
	case <-finished:
		return entries, nil
	case err := <-errs:
		return nil, err
	}
}

func (m *Monitor) addNetwork(e *bonjour.ServiceEntry) {
	device := NetworkDevice{
		Name:    e.Instance,
		Address: e.AddrIPv4.String(),
		Info:    strings.Join(e.Text, " "),
		Port:    e.Port,
	}
	for address, dev := range m.network {
		if address == device.Address {
			changed := dev.merge(device)
			if changed {
				//m.Events <- Event{Name: "change", NetworkDevice: dev}
			}
			return
		}
	}

	m.network[device.Address] = &device
	//m.Events <- Event{Name: "add", NetworkDevice: &device}
}

func (m *Monitor) pruneNetwork() {
	toPrune := []string{}
	for address, dev := range m.network {
		if !dev.isUp(2) {
			toPrune = append(toPrune, address)
		}
	}

	for _, port := range toPrune {
		//m.Events <- Event{Name: "remove", NetworkDevice: m.network[port]}
		delete(m.network, port)
	}
}

// filter returns a new slice containing all NetworkDevice in the slice that satisfy the predicate f.
func filter(vs []NetworkDevice, f func(NetworkDevice) bool) []NetworkDevice {
	var vsf []NetworkDevice
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}
