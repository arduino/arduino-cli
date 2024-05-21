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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/arduino/arduino-cli/internal/i18n"
	discovery "github.com/arduino/pluggable-discovery-protocol-handler/v2"
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
	discoveries        map[string]*discovery.Client // all registered PluggableDiscovery
	discoveriesRunning bool                         // set to true once discoveries are started
	feed               chan *discovery.Event        // all events will pass through this channel
	watchersMutex      sync.Mutex
	watchers           map[*PortWatcher]bool                  // all registered Watcher
	watchersCache      map[string]map[string]*discovery.Event // this is a cache of all active ports
	userAgent          string
}

// New creates a new DiscoveryManager
func New(userAgent string) *DiscoveryManager {
	return &DiscoveryManager{
		discoveries:   map[string]*discovery.Client{},
		watchers:      map[*PortWatcher]bool{},
		feed:          make(chan *discovery.Event, 50),
		watchersCache: map[string]map[string]*discovery.Event{},
		userAgent:     userAgent,
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
	dm.discoveries = map[string]*discovery.Client{}
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
func (dm *DiscoveryManager) Start() []error {
	dm.discoveriesMutex.Lock()
	defer dm.discoveriesMutex.Unlock()
	if dm.discoveriesRunning {
		return nil
	}

	go func() {
		// Send all events coming from the feed channel to all active watchers
		for ev := range dm.feed {
			dm.feedEvent(ev)
		}
	}()

	errs := []error{}
	var errsLock sync.Mutex

	var wg sync.WaitGroup
	for _, d := range dm.discoveries {
		wg.Add(1)
		go func(d *discovery.Client) {
			if err := dm.startDiscovery(d); err != nil {
				errsLock.Lock()
				errs = append(errs, err)
				errsLock.Unlock()
			}
			wg.Done()
		}(d)
	}
	wg.Wait()
	dm.discoveriesRunning = true

	return errs
}

// Add adds a discovery to the list of managed discoveries
func (dm *DiscoveryManager) Add(id string, args ...string) error {
	d := discovery.NewClient(id, args...)
	d.SetLogger(logrus.WithField("discovery", id))
	d.SetUserAgent(dm.userAgent)
	return dm.add(d)
}

func (dm *DiscoveryManager) add(d *discovery.Client) error {
	dm.discoveriesMutex.Lock()
	defer dm.discoveriesMutex.Unlock()

	id := d.GetID()
	if _, has := dm.discoveries[id]; has {
		return errors.New(i18n.Tr("pluggable discovery already added: %s", id))
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
		watcher.feed = nil
	}
	go func() {
		dm.watchersMutex.Lock()
		defer dm.watchersMutex.Unlock()

		// Check if the watcher is still alive (it could have been closed before the goroutine started...)
		if watcher.feed == nil {
			return
		}

		// When a watcher is started, send all the current active ports first...
		for _, cache := range dm.watchersCache {
			for _, ev := range cache {
				watcher.feed <- ev
			}
		}
		// ...and after that add the watcher to the list of watchers receiving events
		dm.watchers[watcher] = true
	}()
	return watcher, nil
}

func (dm *DiscoveryManager) startDiscovery(d *discovery.Client) (discErr error) {
	defer func() {
		// If this function returns an error log it
		if discErr != nil {
			logrus.Errorf("Discovery %s failed to run: %s", d.GetID(), discErr)
		}
	}()

	if err := d.Run(); err != nil {
		return fmt.Errorf("%s: %w", i18n.Tr("discovery %[1]s process not started", d.GetID()), err)
	}
	eventCh, err := d.StartSync(5)
	if err != nil {
		return fmt.Errorf("%s: %s", i18n.Tr("starting discovery %s", d.GetID()), err)
	}

	go func(d *discovery.Client) {
		// Transfer all incoming events from this discovery to the feed channel
		for ev := range eventCh {
			dm.feed <- ev
		}
		logrus.Infof("Discovery event channel closed %s. Exiting goroutine.", d.GetID())
	}(d)
	return nil
}

func (dm *DiscoveryManager) feedEvent(ev *discovery.Event) {
	dm.watchersMutex.Lock()
	defer dm.watchersMutex.Unlock()

	sendToAllWatchers := func(ev *discovery.Event) {
		// Send the event to all watchers
		for watcher := range dm.watchers {
			select {
			case watcher.feed <- ev:
				// OK
			case <-time.After(time.Millisecond * 500):
				// If the watcher is not able to process event fast enough
				// remove the watcher from the list of watchers
				logrus.Error("Watcher is not able to process events fast enough, removing it from the list of watchers")
				delete(dm.watchers, watcher)
			}
		}
	}

	if ev.Type == "stop" {
		// Send remove events for all the cached ports of the terminating discovery
		cache := dm.watchersCache[ev.DiscoveryID]
		for _, addEv := range cache {
			removeEv := &discovery.Event{
				Type: "remove",
				Port: &discovery.Port{
					Address:       addEv.Port.Address,
					AddressLabel:  addEv.Port.AddressLabel,
					Protocol:      addEv.Port.Protocol,
					ProtocolLabel: addEv.Port.ProtocolLabel},
				DiscoveryID: addEv.DiscoveryID,
			}
			sendToAllWatchers(removeEv)
		}

		// Remove the cache for the terminating discovery
		delete(dm.watchersCache, ev.DiscoveryID)
		return
	}

	sendToAllWatchers(ev)

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

// AddAllDiscoveriesFrom transfers discoveries from src to the receiver
func (dm *DiscoveryManager) AddAllDiscoveriesFrom(src *DiscoveryManager) {
	for _, d := range src.discoveries {
		dm.add(d)
	}
}
