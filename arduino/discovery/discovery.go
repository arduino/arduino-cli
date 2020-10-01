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

package discovery

import (
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

// PluggableDiscovery is a tool that detects communication ports to interact
// with the boards.
type PluggableDiscovery struct {
	id                   string
	args                 []string
	process              *executils.Process
	outgoingCommandsPipe io.Writer
	incomingMessagesChan <-chan *discoveryMessage

	// All the following fields are guarded by statusMutex
	statusMutex           sync.Mutex
	incomingMessagesError error
	alive                 bool
	eventsMode            bool
	eventChan             chan<- *Event
	cachedPorts           map[string]*Port
}

type discoveryMessage struct {
	EventType string  `json:"eventType"`
	Message   string  `json:"message"`
	Ports     []*Port `json:"ports"`
	Port      *Port   `json:"port"`
}

// Port containts metadata about a port to connect to a board.
type Port struct {
	Address                  string          `json:"address"`
	AddressLabel             string          `json:"label"`
	Protocol                 string          `json:"protocol"`
	ProtocolLabel            string          `json:"protocolLabel"`
	Properties               *properties.Map `json:"prefs"`
	IdentificationProperties *properties.Map `json:"identificationPrefs"`
}

func (p *Port) String() string {
	if p == nil {
		return "none"
	}
	return p.Address
}

// Event is a pluggable discovery event
type Event struct {
	Type string
	Port *Port
}

// New create and connect to the given pluggable discovery
func New(id string, args ...string) (*PluggableDiscovery, error) {
	proc, err := executils.NewProcess(args...)
	if err != nil {
		return nil, err
	}
	stdout, err := proc.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdin, err := proc.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err := proc.Start(); err != nil {
		return nil, err
	}
	messageChan := make(chan *discoveryMessage)
	disc := &PluggableDiscovery{
		id:                   id,
		process:              proc,
		incomingMessagesChan: messageChan,
		outgoingCommandsPipe: stdin,
		alive:                true,
	}
	go disc.jsonDecodeLoop(stdout, messageChan)
	return disc, nil
}

// GetID returns the identifier for this discovery
func (disc *PluggableDiscovery) GetID() string {
	return disc.id
}

func (disc *PluggableDiscovery) String() string {
	return disc.id
}

func (disc *PluggableDiscovery) jsonDecodeLoop(in io.Reader, outChan chan<- *discoveryMessage) {
	decoder := json.NewDecoder(in)
	closeAndReportError := func(err error) {
		disc.statusMutex.Lock()
		disc.alive = false
		disc.incomingMessagesError = err
		disc.statusMutex.Unlock()
		close(outChan)
	}
	for {
		var msg discoveryMessage
		if err := decoder.Decode(&msg); err != nil {
			closeAndReportError(err)
			return
		}

		if msg.EventType == "add" {
			if msg.Port == nil {
				closeAndReportError(errors.New("invalid 'add' message: missing port"))
				return
			}
			disc.statusMutex.Lock()
			disc.cachedPorts[msg.Port.Address] = msg.Port
			if disc.eventChan != nil {
				disc.eventChan <- &Event{"add", msg.Port}
			}
			disc.statusMutex.Unlock()
		} else if msg.EventType == "remove" {
			if msg.Port == nil {
				closeAndReportError(errors.New("invalid 'remove' message: missing port"))
				return
			}
			disc.statusMutex.Lock()
			delete(disc.cachedPorts, msg.Port.Address)
			if disc.eventChan != nil {
				disc.eventChan <- &Event{"remove", msg.Port}
			}
			disc.statusMutex.Unlock()
		} else {
			outChan <- &msg
		}
	}
}

// IsAlive return true if the discovery process is running and so is able to receive commands
// and produce events.
func (disc *PluggableDiscovery) IsAlive() bool {
	disc.statusMutex.Lock()
	defer disc.statusMutex.Unlock()
	return disc.alive
}

// IsEventMode return true if the discovery is in "events" mode
func (disc *PluggableDiscovery) IsEventMode() bool {
	disc.statusMutex.Lock()
	defer disc.statusMutex.Unlock()
	return disc.eventsMode
}

func (disc *PluggableDiscovery) waitMessage(timeout time.Duration) (*discoveryMessage, error) {
	select {
	case msg := <-disc.incomingMessagesChan:
		if msg == nil {
			// channel has been closed
			disc.statusMutex.Lock()
			defer disc.statusMutex.Unlock()
			return nil, disc.incomingMessagesError
		}
		return msg, nil
	case <-time.After(timeout):
		return nil, errors.New("timeout")
	}
}

func (disc *PluggableDiscovery) sendCommand(command string) error {
	if n, err := disc.outgoingCommandsPipe.Write([]byte(command)); err != nil {
		return err
	} else if n < len(command) {
		return disc.sendCommand(command[n:])
	} else {
		return nil
	}
}

// Start initializes and start the discovery internal subroutines. This command must be
// called before List or StartSync.
func (disc *PluggableDiscovery) Start() error {
	if err := disc.sendCommand("START\n"); err != nil {
		return err
	}
	if msg, err := disc.waitMessage(time.Second * 10); err != nil {
		return err
	} else if msg.EventType != "start" {
		return errors.Errorf("communication out of sync, expected 'start', received '%s'", msg.EventType)
	} else if msg.Message != "OK" {
		return errors.Errorf("command failed: %s", msg.Message)
	}
	return nil
}

// Stop stops the discovery internal subroutines and possibly free the internally
// used resources. This command should be called if the client wants to pause the
// discovery for a while.
func (disc *PluggableDiscovery) Stop() error {
	if err := disc.sendCommand("STOP\n"); err != nil {
		return err
	}
	if msg, err := disc.waitMessage(time.Second * 10); err != nil {
		return err
	} else if msg.EventType != "stop" {
		return errors.Errorf("communication out of sync, expected 'stop', received '%s'", msg.EventType)
	} else if msg.Message != "OK" {
		return errors.Errorf("command failed: %s", msg.Message)
	}
	return nil
}

// Quit terminates the discovery. No more commands can be accepted by the discovery.
func (disc *PluggableDiscovery) Quit() error {
	if err := disc.sendCommand("QUIT\n"); err != nil {
		return err
	}
	if msg, err := disc.waitMessage(time.Second * 10); err != nil {
		return err
	} else if msg.EventType != "quit" {
		return errors.Errorf("communication out of sync, expected 'quit', received '%s'", msg.EventType)
	} else if msg.Message != "OK" {
		return errors.Errorf("command failed: %s", msg.Message)
	}
	return nil
}

// List executes an enumeration of the ports and returns a list of the available
// ports at the moment of the call.
func (disc *PluggableDiscovery) List() ([]*Port, error) {
	if err := disc.sendCommand("LIST\n"); err != nil {
		return nil, err
	}
	if msg, err := disc.waitMessage(time.Second * 10); err != nil {
		return nil, err
	} else if msg.EventType != "list" {
		return nil, errors.Errorf("communication out of sync, expected 'list', received '%s'", msg.EventType)
	} else {
		return msg.Ports, nil
	}
}

// EventChannel creates a channel used to receive events from the pluggable discovery.
// The event channel must be consumed as quickly as possible since it may block the
// discovery if it becomes full. The channel size is configurable.
func (disc *PluggableDiscovery) EventChannel(size int) <-chan *Event {
	c := make(chan *Event, size)
	disc.statusMutex.Lock()
	defer disc.statusMutex.Unlock()
	disc.eventChan = c
	return c
}

// StartSync puts the discovery in "events" mode: the discovery will send "add"
// and "remove" events each time a new port is detected or removed respectively.
// After calling StartSync an initial burst of "add" events may be generated to
// report all the ports available at the moment of the start.
func (disc *PluggableDiscovery) StartSync() error {
	disc.statusMutex.Lock()
	defer disc.statusMutex.Unlock()

	if disc.eventsMode {
		return errors.New("already in events mode")
	}
	if err := disc.sendCommand("START_SYNC\n"); err != nil {
		return err
	}

	// START_SYNC does not give any response

	disc.eventsMode = true
	disc.cachedPorts = map[string]*Port{}
	return nil
}

// ListSync returns a list of the available ports. The list is a cache of all the
// add/remove events happened from the StartSync call and it will not consume any
// resource from the underliying discovery.
func (disc *PluggableDiscovery) ListSync() []*Port {
	disc.statusMutex.Lock()
	defer disc.statusMutex.Unlock()
	res := []*Port{}
	for _, port := range disc.cachedPorts {
		res = append(res, port)
	}
	return res
}
