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
	"fmt"
	"sync"

	"github.com/arduino/arduino-cli/arduino/discovery"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// DiscoveryManager is required to handle multiple pluggable-discovery that
// may be shared across platforms
type DiscoveryManager struct {
	discoveries map[string]*discovery.PluggableDiscovery
}

var tr = i18n.Tr

// New creates a new DiscoveryManager
func New() *DiscoveryManager {
	return &DiscoveryManager{
		discoveries: map[string]*discovery.PluggableDiscovery{},
	}
}

// Clear resets the DiscoveryManager to its initial state
func (dm *DiscoveryManager) Clear() {
	dm.QuitAll()
	dm.discoveries = map[string]*discovery.PluggableDiscovery{}
}

// IDs returns the list of discoveries' ids in this DiscoveryManager
func (dm *DiscoveryManager) IDs() []string {
	ids := []string{}
	for id := range dm.discoveries {
		ids = append(ids, id)
	}
	return ids
}

// Add adds a discovery to the list of managed discoveries
func (dm *DiscoveryManager) Add(disc *discovery.PluggableDiscovery) error {
	id := disc.GetID()
	if _, has := dm.discoveries[id]; has {
		return errors.Errorf(tr("pluggable discovery already added: %s"), id)
	}
	dm.discoveries[id] = disc
	return nil
}

// remove quits and deletes the discovery with specified id
// from the discoveries managed by this DiscoveryManager
func (dm *DiscoveryManager) remove(id string) {
	dm.discoveries[id].Quit()
	delete(dm.discoveries, id)
	logrus.Infof("Closed and removed discovery %s", id)
}

// parallelize runs function f concurrently for each discovery.
// Returns a list of errors returned by each call of f.
func (dm *DiscoveryManager) parallelize(f func(d *discovery.PluggableDiscovery) error) []error {
	var wg sync.WaitGroup
	errChan := make(chan error)
	for _, d := range dm.discoveries {
		wg.Add(1)
		go func(d *discovery.PluggableDiscovery) {
			defer wg.Done()
			if err := f(d); err != nil {
				errChan <- err
			}
		}(d)
	}

	// Wait in a goroutine to collect eventual errors running a discovery.
	// When all goroutines that are calling discoveries are done close the errors chan.
	go func() {
		wg.Wait()
		close(errChan)
	}()

	errs := []error{}
	for err := range errChan {
		errs = append(errs, err)
	}
	return errs
}

// RunAll the discoveries for this DiscoveryManager,
// returns an error for each discovery failing to run
func (dm *DiscoveryManager) RunAll() []error {
	return dm.parallelize(func(d *discovery.PluggableDiscovery) error {
		if d.State() != discovery.Dead {
			// This discovery is already alive, nothing to do
			return nil
		}

		if err := d.Run(); err != nil {
			dm.remove(d.GetID())
			return fmt.Errorf(tr("discovery %[1]s process not started: %[2]w"), d.GetID(), err)
		}
		return nil
	})
}

// StartAll the discoveries for this DiscoveryManager,
// returns an error for each discovery failing to start
func (dm *DiscoveryManager) StartAll() []error {
	return dm.parallelize(func(d *discovery.PluggableDiscovery) error {
		state := d.State()
		if state != discovery.Idling {
			// Already started
			return nil
		}
		if err := d.Start(); err != nil {
			dm.remove(d.GetID())
			return fmt.Errorf(tr("starting discovery %[1]s: %[2]w"), d.GetID(), err)
		}
		return nil
	})
}

// StartSyncAll the discoveries for this DiscoveryManager,
// returns an error for each discovery failing to start syncing
func (dm *DiscoveryManager) StartSyncAll() (<-chan *discovery.Event, []error) {
	eventSink := make(chan *discovery.Event, 5)
	var wg sync.WaitGroup
	errs := dm.parallelize(func(d *discovery.PluggableDiscovery) error {
		state := d.State()
		if state != discovery.Idling || state == discovery.Syncing {
			// Already syncing
			return nil
		}

		eventCh, err := d.StartSync(5)
		if err != nil {
			dm.remove(d.GetID())
			return fmt.Errorf(tr("start syncing discovery %[1]s: %[2]w"), d.GetID(), err)
		}

		wg.Add(1)
		go func() {
			for ev := range eventCh {
				eventSink <- ev
			}
			wg.Done()
		}()
		return nil
	})
	go func() {
		wg.Wait()
		eventSink <- &discovery.Event{Type: "quit"}
		close(eventSink)
	}()
	return eventSink, errs
}

// StopAll the discoveries for this DiscoveryManager,
// returns an error for each discovery failing to stop
func (dm *DiscoveryManager) StopAll() []error {
	return dm.parallelize(func(d *discovery.PluggableDiscovery) error {
		state := d.State()
		if state != discovery.Syncing && state != discovery.Running {
			// Not running nor syncing, nothing to stop
			return nil
		}

		if err := d.Stop(); err != nil {
			dm.remove(d.GetID())
			return fmt.Errorf(tr("stopping discovery %[1]s: %[2]w"), d.GetID(), err)
		}
		return nil
	})
}

// QuitAll quits all the discoveries managed by this DiscoveryManager.
// Returns an error for each discovery that fails quitting
func (dm *DiscoveryManager) QuitAll() []error {
	errs := dm.parallelize(func(d *discovery.PluggableDiscovery) error {
		if d.State() == discovery.Dead {
			// Stop! Stop! It's already dead!
			return nil
		}

		d.Quit()
		return nil
	})
	return errs
}

// List returns a list of available ports detected from all discoveries
// and a list of errors for those discoveries that returned one.
func (dm *DiscoveryManager) List() ([]*discovery.Port, []error) {
	var wg sync.WaitGroup
	// Use this struct to avoid the need of two separate
	// channels for ports and errors.
	type listMsg struct {
		Err  error
		Port *discovery.Port
	}
	msgChan := make(chan listMsg)
	for _, d := range dm.discoveries {
		wg.Add(1)
		go func(d *discovery.PluggableDiscovery) {
			defer wg.Done()
			if d.State() != discovery.Running {
				// Discovery is not running, it won't return anything
				return
			}
			ports, err := d.List()
			if err != nil {
				msgChan <- listMsg{Err: fmt.Errorf(tr("listing ports from discovery %[1]s: %[2]w"), d.GetID(), err)}
			}
			for _, p := range ports {
				msgChan <- listMsg{Port: p}
			}
		}(d)
	}

	go func() {
		// Close the channel only after all goroutines are done
		wg.Wait()
		close(msgChan)
	}()

	ports := []*discovery.Port{}
	errs := []error{}
	for msg := range msgChan {
		if msg.Err != nil {
			errs = append(errs, msg.Err)
		} else {
			ports = append(ports, msg.Port)
		}
	}
	return ports, errs
}

// ListCachedPorts return the current list of ports detected from all discoveries
func (dm *DiscoveryManager) ListCachedPorts() []*discovery.Port {
	res := []*discovery.Port{}
	for _, d := range dm.discoveries {
		if d.State() != discovery.Syncing {
			// Discovery is not syncing
			continue
		}
		res = append(res, d.ListCachedPorts()...)
	}
	return res
}
