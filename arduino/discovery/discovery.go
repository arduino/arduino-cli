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
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/version"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// To work correctly a Pluggable Discovery must respect the state machine specifed on the documentation:
// https://arduino.github.io/arduino-cli/latest/pluggable-discovery-specification/#state-machine
// States a PluggableDiscovery can be in
const (
	Alive int = iota
	Idling
	Running
	Syncing
	Dead
)

// PluggableDiscovery is a tool that detects communication ports to interact
// with the boards.
type PluggableDiscovery struct {
	id                   string
	processArgs          []string
	process              *executils.Process
	outgoingCommandsPipe io.Writer
	incomingMessagesChan <-chan *discoveryMessage

	// All the following fields are guarded by statusMutex
	statusMutex           sync.Mutex
	incomingMessagesError error
	state                 int
	eventChan             chan<- *Event
}

type discoveryMessage struct {
	EventType       string  `json:"eventType"`
	Message         string  `json:"message"`
	Error           bool    `json:"error"`
	ProtocolVersion int     `json:"protocolVersion"` // Used in HELLO command
	Ports           []*Port `json:"ports"`           // Used in LIST command
	Port            *Port   `json:"port"`            // Used in add and remove events
}

func (msg discoveryMessage) String() string {
	s := fmt.Sprintf("type: %s", msg.EventType)
	if msg.Message != "" {
		s = tr("%[1]s, message: %[2]s", s, msg.Message)
	}
	if msg.ProtocolVersion != 0 {
		s = tr("%[1]s, protocol version: %[2]d", s, msg.ProtocolVersion)
	}
	if len(msg.Ports) > 0 {
		s = tr("%[1]s, ports: %[2]s", s, msg.Ports)
	}
	if msg.Port != nil {
		s = tr("%[1]s, port: %[2]s", s, msg.Port)
	}
	return s
}

// Port containts metadata about a port to connect to a board.
type Port struct {
	Address       string          `json:"address"`
	AddressLabel  string          `json:"label"`
	Protocol      string          `json:"protocol"`
	ProtocolLabel string          `json:"protocolLabel"`
	HardwareID    string          `json:"hardwareId,omitempty"`
	Properties    *properties.Map `json:"properties"`
}

var tr = i18n.Tr

// ToRPC converts Port into rpc.Port
func (p *Port) ToRPC() *rpc.Port {
	props := p.Properties
	if props == nil {
		props = properties.NewMap()
	}
	return &rpc.Port{
		Address:       p.Address,
		Label:         p.AddressLabel,
		Protocol:      p.Protocol,
		ProtocolLabel: p.ProtocolLabel,
		HardwareId:    p.HardwareID,
		Properties:    props.AsMap(),
	}
}

// PortFromRPCPort converts an *rpc.Port to a *Port
func PortFromRPCPort(o *rpc.Port) (p *Port) {
	if o == nil {
		return nil
	}
	res := &Port{
		Address:       o.Address,
		AddressLabel:  o.Label,
		Protocol:      o.Protocol,
		ProtocolLabel: o.ProtocolLabel,
		HardwareID:    o.HardwareId,
	}
	if o.Properties != nil {
		res.Properties = properties.NewFromHashmap(o.Properties)
	}
	return res
}

func (p *Port) String() string {
	if p == nil {
		return "none"
	}
	return p.Address
}

// Clone creates a copy of this Port
func (p *Port) Clone() *Port {
	if p == nil {
		return nil
	}
	var res Port = *p
	if p.Properties != nil {
		res.Properties = p.Properties.Clone()
	}
	return &res
}

// Event is a pluggable discovery event
type Event struct {
	Type        string
	Port        *Port
	DiscoveryID string
}

// New create and connect to the given pluggable discovery
func New(id string, args ...string) *PluggableDiscovery {
	return &PluggableDiscovery{
		id:          id,
		processArgs: args,
		state:       Dead,
	}
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
		disc.state = Dead
		disc.incomingMessagesError = err
		disc.statusMutex.Unlock()
		close(outChan)
		logrus.Errorf("stopped discovery %s decode loop: %v", disc.id, err)
	}

	for {
		var msg discoveryMessage
		if err := decoder.Decode(&msg); err == io.EOF {
			// This is fine, we exit gracefully
			disc.statusMutex.Lock()
			disc.state = Dead
			disc.incomingMessagesError = err
			disc.statusMutex.Unlock()
			close(outChan)
			return
		} else if err != nil {
			closeAndReportError(err)
			return
		}
		logrus.Infof("from discovery %s received message %s", disc.id, msg)
		if msg.EventType == "add" {
			if msg.Port == nil {
				closeAndReportError(errors.New(tr("invalid 'add' message: missing port")))
				return
			}
			disc.statusMutex.Lock()
			if disc.eventChan != nil {
				disc.eventChan <- &Event{"add", msg.Port, disc.GetID()}
			}
			disc.statusMutex.Unlock()
		} else if msg.EventType == "remove" {
			if msg.Port == nil {
				closeAndReportError(errors.New(tr("invalid 'remove' message: missing port")))
				return
			}
			disc.statusMutex.Lock()
			if disc.eventChan != nil {
				disc.eventChan <- &Event{"remove", msg.Port, disc.GetID()}
			}
			disc.statusMutex.Unlock()
		} else {
			outChan <- &msg
		}
	}
}

// State returns the current state of this PluggableDiscovery
func (disc *PluggableDiscovery) State() int {
	disc.statusMutex.Lock()
	defer disc.statusMutex.Unlock()
	return disc.state
}

func (disc *PluggableDiscovery) waitMessage(timeout time.Duration) (*discoveryMessage, error) {
	select {
	case msg := <-disc.incomingMessagesChan:
		if msg == nil {
			return nil, disc.incomingMessagesError
		}
		return msg, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf(tr("timeout waiting for message from %s"), disc.id)
	}
}

func (disc *PluggableDiscovery) sendCommand(command string) error {
	logrus.Infof("sending command %s to discovery %s", strings.TrimSpace(command), disc)
	data := []byte(command)
	for {
		n, err := disc.outgoingCommandsPipe.Write(data)
		if err != nil {
			return err
		}
		if n == len(data) {
			return nil
		}
		data = data[n:]
	}
}

func (disc *PluggableDiscovery) runProcess() error {
	logrus.Infof("starting discovery %s process", disc.id)
	proc, err := executils.NewProcess(nil, disc.processArgs...)
	if err != nil {
		return err
	}
	stdout, err := proc.StdoutPipe()
	if err != nil {
		return err
	}
	stdin, err := proc.StdinPipe()
	if err != nil {
		return err
	}
	disc.outgoingCommandsPipe = stdin

	messageChan := make(chan *discoveryMessage)
	disc.incomingMessagesChan = messageChan
	go disc.jsonDecodeLoop(stdout, messageChan)

	if err := proc.Start(); err != nil {
		return err
	}

	disc.statusMutex.Lock()
	defer disc.statusMutex.Unlock()
	disc.process = proc
	disc.state = Alive
	logrus.Infof("started discovery %s process", disc.id)
	return nil
}

func (disc *PluggableDiscovery) killProcess() error {
	logrus.Infof("killing discovery %s process", disc.id)
	if disc.process != nil {
		if err := disc.process.Kill(); err != nil {
			return err
		}
		if err := disc.process.Wait(); err != nil {
			return err
		}
	}
	disc.statusMutex.Lock()
	defer disc.statusMutex.Unlock()
	disc.stopSync()
	disc.state = Dead
	logrus.Infof("killed discovery %s process", disc.id)
	return nil
}

// Run starts the discovery executable process and sends the HELLO command to the discovery to agree on the
// pluggable discovery protocol. This must be the first command to run in the communication with the discovery.
// If the process is started but the HELLO command fails the process is killed.
func (disc *PluggableDiscovery) Run() (err error) {
	if err = disc.runProcess(); err != nil {
		return err
	}

	defer func() {
		// If the discovery process is started successfully but the HELLO handshake
		// fails the discovery is an unusable state, we kill the process to avoid
		// further issues down the line.
		if err == nil {
			return
		}
		if err := disc.killProcess(); err != nil {
			// Log failure to kill the process, ideally that should never happen
			// but it's best to know it if it does
			logrus.Errorf("Killing discovery %s after unsuccessful start: %s", disc.id, err)
		}
	}()

	if err = disc.sendCommand("HELLO 1 \"arduino-cli " + version.VersionInfo.VersionString + "\"\n"); err != nil {
		return err
	}
	if msg, err := disc.waitMessage(time.Second * 10); err != nil {
		return fmt.Errorf(tr("calling %[1]s: %[2]w"), "HELLO", err)
	} else if msg.EventType != "hello" {
		return errors.Errorf(tr("communication out of sync, expected '%[1]s', received '%[2]s'"), "hello", msg.EventType)
	} else if msg.Error {
		return errors.Errorf(tr("command failed: %s"), msg.Message)
	} else if strings.ToUpper(msg.Message) != "OK" {
		return errors.Errorf(tr("communication out of sync, expected '%[1]s', received '%[2]s'"), "OK", msg.Message)
	} else if msg.ProtocolVersion > 1 {
		return errors.Errorf(tr("protocol version not supported: requested 1, got %d"), msg.ProtocolVersion)
	}
	disc.statusMutex.Lock()
	defer disc.statusMutex.Unlock()
	disc.state = Idling
	return nil
}

// Start initializes and start the discovery internal subroutines. This command must be
// called before List or StartSync.
func (disc *PluggableDiscovery) Start() error {
	if err := disc.sendCommand("START\n"); err != nil {
		return err
	}
	if msg, err := disc.waitMessage(time.Second * 10); err != nil {
		return fmt.Errorf(tr("calling %[1]s: %[2]w"), "START", err)
	} else if msg.EventType != "start" {
		return errors.Errorf(tr("communication out of sync, expected '%[1]s', received '%[2]s'"), "start", msg.EventType)
	} else if msg.Error {
		return errors.Errorf(tr("command failed: %s"), msg.Message)
	} else if strings.ToUpper(msg.Message) != "OK" {
		return errors.Errorf(tr("communication out of sync, expected '%[1]s', received '%[2]s'"), "OK", msg.Message)
	}
	disc.statusMutex.Lock()
	defer disc.statusMutex.Unlock()
	disc.state = Running
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
		return fmt.Errorf(tr("calling %[1]s: %[2]w"), "STOP", err)
	} else if msg.EventType != "stop" {
		return errors.Errorf(tr("communication out of sync, expected '%[1]s', received '%[2]s'"), "stop", msg.EventType)
	} else if msg.Error {
		return errors.Errorf(tr("command failed: %s"), msg.Message)
	} else if strings.ToUpper(msg.Message) != "OK" {
		return errors.Errorf(tr("communication out of sync, expected '%[1]s', received '%[2]s'"), "OK", msg.Message)
	}
	disc.statusMutex.Lock()
	defer disc.statusMutex.Unlock()
	disc.stopSync()
	disc.state = Idling
	return nil
}

func (disc *PluggableDiscovery) stopSync() {
	if disc.eventChan != nil {
		disc.eventChan <- &Event{"stop", nil, disc.GetID()}
		close(disc.eventChan)
		disc.eventChan = nil
	}
}

// Quit terminates the discovery. No more commands can be accepted by the discovery.
func (disc *PluggableDiscovery) Quit() {
	_ = disc.sendCommand("QUIT\n")
	if _, err := disc.waitMessage(time.Second * 5); err != nil {
		logrus.Errorf("Quitting discovery %s: %s", disc.id, err)
	}
	disc.stopSync()
	disc.killProcess()
}

// List executes an enumeration of the ports and returns a list of the available
// ports at the moment of the call.
func (disc *PluggableDiscovery) List() ([]*Port, error) {
	if err := disc.sendCommand("LIST\n"); err != nil {
		return nil, err
	}
	if msg, err := disc.waitMessage(time.Second * 10); err != nil {
		return nil, fmt.Errorf(tr("calling %[1]s: %[2]w"), "LIST", err)
	} else if msg.EventType != "list" {
		return nil, errors.Errorf(tr("communication out of sync, expected '%[1]s', received '%[2]s'"), "list", msg.EventType)
	} else if msg.Error {
		return nil, errors.Errorf(tr("command failed: %s"), msg.Message)
	} else {
		return msg.Ports, nil
	}
}

// StartSync puts the discovery in "events" mode: the discovery will send "add"
// and "remove" events each time a new port is detected or removed respectively.
// After calling StartSync an initial burst of "add" events may be generated to
// report all the ports available at the moment of the start.
// It also creates a channel used to receive events from the pluggable discovery.
// The event channel must be consumed as quickly as possible since it may block the
// discovery if it becomes full. The channel size is configurable.
func (disc *PluggableDiscovery) StartSync(size int) (<-chan *Event, error) {
	disc.statusMutex.Lock()
	defer disc.statusMutex.Unlock()

	if err := disc.sendCommand("START_SYNC\n"); err != nil {
		return nil, err
	}

	if msg, err := disc.waitMessage(time.Second * 10); err != nil {
		return nil, fmt.Errorf(tr("calling %[1]s: %[2]w"), "START_SYNC", err)
	} else if msg.EventType != "start_sync" {
		return nil, errors.Errorf(tr("communication out of sync, expected '%[1]s', received '%[2]s'"), "start_sync", msg.EventType)
	} else if msg.Error {
		return nil, errors.Errorf(tr("command failed: %s"), msg.Message)
	} else if strings.ToUpper(msg.Message) != "OK" {
		return nil, errors.Errorf(tr("communication out of sync, expected '%[1]s', received '%[2]s'"), "OK", msg.Message)
	}

	disc.state = Syncing
	// In case there is already an existing event channel in use we close it before creating a new one.
	disc.stopSync()
	c := make(chan *Event, size)
	disc.eventChan = c
	return c, nil
}
