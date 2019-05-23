//
// This file is part of arduino-cli.
//
// Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.
//

package commands

import (
	"sync"

	"github.com/arduino/arduino-cli/arduino/discovery"
)

type sharedDiscovery struct {
	discovery      *discovery.Discovery
	referenceCount int
}

// this map contains all the running pluggable-discoveries instances
var sharedDiscoveries = map[string]*sharedDiscovery{}
var sharedDiscoveriesMutex sync.Mutex

// StartSharedDiscovery starts a discovery or returns the instance of an already
// started shared discovery.
func StartSharedDiscovery(disc *discovery.Discovery) (*discovery.Discovery, error) {
	sharedDiscoveriesMutex.Lock()
	defer sharedDiscoveriesMutex.Unlock()

	instance, started := sharedDiscoveries[disc.ID]
	if started {
		instance.referenceCount++
		return instance.discovery, nil
	}
	sharedDiscoveries[disc.ID] = &sharedDiscovery{
		discovery:      disc,
		referenceCount: 1,
	}
	err := disc.Start()
	return disc, err
}

// StopSharedDiscovery will dispose an instance of a shared discovery if it is
// no more needed.
func StopSharedDiscovery(disc *discovery.Discovery) error {
	sharedDiscoveriesMutex.Lock()
	defer sharedDiscoveriesMutex.Unlock()

	instance, started := sharedDiscoveries[disc.ID]
	if started {
		instance.referenceCount--

		if instance.referenceCount == 0 {
			delete(sharedDiscoveries, disc.ID)
			return instance.discovery.Close()
		}
	}
	return nil
}
