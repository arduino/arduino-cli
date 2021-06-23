// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package discoverymanager

import (
	"github.com/arduino/arduino-cli/arduino/discovery"
	"github.com/pkg/errors"
)

// DiscoveryManager is required to handle multiple pluggable-discovery that
// may be shared across platforms
type DiscoveryManager struct {
	discoveries map[string]*discovery.PluggableDiscovery
}

// New creates a new DiscoveriesManager
func New() *DiscoveryManager {
	return &DiscoveryManager{
		discoveries: map[string]*discovery.PluggableDiscovery{},
	}
}

// Add adds a discovery to the list of managed discoveries
func (dm *DiscoveryManager) Add(disc *discovery.PluggableDiscovery) error {
	id := disc.GetID()
	if _, has := dm.discoveries[id]; has {
		return errors.Errorf("pluggable discovery already added: %s", id)
	}
	dm.discoveries[id] = disc
	return nil
}

// StartAll the discoveries for this DiscoveryManager,
// returns the first error it meets or nil
func (dm *DiscoveryManager) StartAll() error {
	for _, d := range dm.discoveries {
		err := d.Start()
		if err != nil {
			return err
		}
		err = d.StartSync()
		if err != nil {
			return err
		}
	}
	return nil
}

// StopAll the discoveries for this DiscoveryManager,
// returns the first error it meets or nil
func (dm *DiscoveryManager) StopAll() error {
	for _, d := range dm.discoveries {
		err := d.Stop()
		if err != nil {
			return err
		}
	}
	return nil
}

// ListPorts return the current list of ports detected from all discoveries
func (dm *DiscoveryManager) ListPorts() []*discovery.Port {
	// c := make(chan []*discovery.Port, len(dm.discoveries))

	// var wg sync.WaitGroup
	// for _, d := range dm.discoveries {
	// 	wg.Add(1)
	// 	d := d
	// 	go func() {
	// 		c <- d.ListSync()
	// 		wg.Done()
	// 	}()
	// }
	// wg.Wait()
	// // Close the channel only after all the goroutines are finished
	// close(c)

	// ports := []*discovery.Port{}
	// for p := range c {
	// 	ports = append(ports, p...)
	// }

	// return ports
	res := []*discovery.Port{}
	for _, disc := range dm.discoveries {
		res = append(res, disc.ListSync()...)
	}
	return res
}
