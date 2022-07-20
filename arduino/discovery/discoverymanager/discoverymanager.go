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
	"time"

	"github.com/arduino/arduino-cli/arduino/discovery"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// DiscoveryManager manages the many-to-many communication between all pluggable
// discoveries and all watchers. Each PluggableDiscovery, once started, will
// produce a sequence of "events". These events will be broadcasted to all
// listening Watcher.
// The DiscoveryManager will not start the discoveries until the Start method
// is called.
type DiscoveryManager struct {
	discoveriesMutex   sync.Mutex
	discoveries        map[string]*discovery.PluggableDiscovery // all registered PluggableDiscovery
	discoveriesRunning bool                                     // set to true once discoveries are started
	feed               chan *discovery.Event                    // all events will pass through this channel
	watchersMutex      sync.Mutex
	watchers           map[*PortWatcher]bool                  // all registered Watcher
	watchersCache      map[string]map[string]*discovery.Event // this is a cache of all active ports
}

var tr = i18n.Tr

// New creates a new DiscoveryManager
func New() *DiscoveryManager {
	return &DiscoveryManager{
		discoveries:   map[string]*discovery.PluggableDiscovery{},
		watchers:      map[*PortWatcher]bool{},
		feed:          make(chan *discovery.Event, 50),
		watchersCache: map[string]map[string]*discovery.Event{},
	}
}

// Clear resets the DiscoveryManager to its initial state
func (dm *DiscoveryManager) Clear() {
	dm.discoveriesMutex.Lock()
	defer dm.discoveriesMutex.Unlock()

	if dm.discoveriesRunning {
		for _, d := range dm.discoveries {
			d.Quit()
			logrus.Infof("Closed and removed discovery %s", d.GetID())
		}
	}
	dm.discoveries = map[string]*discovery.PluggableDiscovery{}
}

// IDs returns the list of discoveries' ids in this DiscoveryManager
func (dm *DiscoveryManager) IDs() []string {
	ids := []string{}
	dm.discoveriesMutex.Lock()
	defer dm.discoveriesMutex.Unlock()
	for id := range dm.discoveries {
		ids = append(ids, id)
	}
	return ids
}

// Start starts all the discoveries in this DiscoveryManager.
// If the discoveries are already running, this function does nothing.
func (dm *DiscoveryManager) Start() {
	dm.discoveriesMutex.Lock()
	defer dm.discoveriesMutex.Unlock()
	if dm.discoveriesRunning {
		return
	}

	go func() {
		// Send all events coming from the feed channel to all active watchers
		for ev := range dm.feed {
			dm.feedEvent(ev)
		}
	}()

	var wg sync.WaitGroup
	for _, d := range dm.discoveries {
		wg.Add(1)
		go func(d *discovery.PluggableDiscovery) {
			dm.startDiscovery(d)
			wg.Done()
		}(d)
	}
	wg.Wait()
	dm.discoveriesRunning = true
}

// Add adds a discovery to the list of managed discoveries
func (dm *DiscoveryManager) Add(d *discovery.PluggableDiscovery) error {
	dm.discoveriesMutex.Lock()
	defer dm.discoveriesMutex.Unlock()

	id := d.GetID()
	if _, has := dm.discoveries[id]; has {
		return errors.Errorf(tr("pluggable discovery already added: %s"), id)
	}
	dm.discoveries[id] = d

	if dm.discoveriesRunning {
		dm.startDiscovery(d)
	}
	return nil
}

// PortWatcher is a watcher for all discovery events (port connection/disconnection)
type PortWatcher struct {
	closeCB func()
	feed    chan *discovery.Event
}

// Feed returns the feed of events coming from the discoveries
func (pw *PortWatcher) Feed() <-chan *discovery.Event {
	return pw.feed
}

// Close closes the PortWatcher
func (pw *PortWatcher) Close() {
	pw.closeCB()
}

// Watch starts a watcher for all discovery events (port connection/disconnection).
// The watcher must be closed when it is no longer needed with the Close method.
func (dm *DiscoveryManager) Watch() (*PortWatcher, error) {
	dm.Start()

	watcher := &PortWatcher{
		feed: make(chan *discovery.Event, 10),
	}
	watcher.closeCB = func() {
		dm.watchersMutex.Lock()
		defer dm.watchersMutex.Unlock()
		delete(dm.watchers, watcher)
		close(watcher.feed)
	}
	go func() {
		dm.watchersMutex.Lock()
		// When a watcher is started, send all the current active ports first...
		for _, cache := range dm.watchersCache {
			for _, ev := range cache {
				watcher.feed <- ev
			}
		}
		// ...and after that add the watcher to the list of watchers receiving events
		dm.watchers[watcher] = true
		dm.watchersMutex.Unlock()
	}()
	return watcher, nil
}

func (dm *DiscoveryManager) startDiscovery(d *discovery.PluggableDiscovery) (discErr error) {
	defer func() {
		// If this function returns an error log it
		if discErr != nil {
			logrus.Errorf("Discovery %s failed to run: %s", d.GetID(), discErr)
		}
	}()

	if err := d.Run(); err != nil {
		return fmt.Errorf(tr("discovery %[1]s process not started: %[2]w"), d.GetID(), err)
	}
	eventCh, err := d.StartSync(5)
	if err != nil {
		return fmt.Errorf("%s: %s", tr("starting discovery %s", d.GetID()), err)
	}

	go func() {
		// Transfer all incoming events from this discovery to the feed channel
		for ev := range eventCh {
			dm.feed <- ev
		}
	}()
	return nil
}

func (dm *DiscoveryManager) feedEvent(ev *discovery.Event) {
	dm.watchersMutex.Lock()
	defer dm.watchersMutex.Unlock()

	if ev.Type == "stop" {
		// Remove all the cached events for the terminating discovery
		delete(dm.watchersCache, ev.DiscoveryID)
		return
	}

	// Send the event to all watchers
	for watcher := range dm.watchers {
		select {
		case watcher.feed <- ev:
			// OK
		case <-time.After(time.Millisecond * 500):
			// If the watcher is not able to process event fast enough
			// remove the watcher from the list of watchers
			logrus.Info("Watcher is not able to process events fast enough, removing it from the list of watchers")
			delete(dm.watchers, watcher)
		}
	}

	// Cache the event for the discovery
	cache := dm.watchersCache[ev.DiscoveryID]
	if cache == nil {
		cache = map[string]*discovery.Event{}
		dm.watchersCache[ev.DiscoveryID] = cache
	}
	eventID := ev.Port.Address + "|" + ev.Port.Protocol
	switch ev.Type {
	case "add":
		cache[eventID] = ev
	case "remove":
		delete(cache, eventID)
	default:
		logrus.Errorf("Unhandled event from discovery: %s", ev.Type)
	}
}

// List return the current list of ports detected from all discoveries
func (dm *DiscoveryManager) List() []*discovery.Port {
	dm.Start()

	res := []*discovery.Port{}
	dm.watchersMutex.Lock()
	defer dm.watchersMutex.Unlock()
	for _, cache := range dm.watchersCache {
		for _, ev := range cache {
			res = append(res, ev.Port)
		}
	}
	return res
}
