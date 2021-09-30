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

package monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// To work correctly a Pluggable Monitor must respect the state machine specifed on the documentation:
// https://arduino.github.io/arduino-cli/latest/pluggable-monitor-specification/#state-machine
// States a PluggableMonitor can be in
const (
	Alive int = iota
	Idle
	Opened
	Dead
)

// PluggableMonitor is a tool that communicates with a board through a communication port.
type PluggableMonitor struct {
	id                   string
	processArgs          []string
	process              *executils.Process
	outgoingCommandsPipe io.Writer
	incomingMessagesChan <-chan *monitorMessage
	supportedProtocol    string

	// All the following fields are guarded by statusMutex
	incomingMessagesError error
}

type monitorMessage struct {
	EventType       string          `json:"eventType"`
	Message         string          `json:"message"`
	Error           bool            `json:"error"`
	ProtocolVersion int             `json:"protocolVersion"` // Used in HELLO command
	PortDescription *PortDescriptor `json:"port_description,omitempty"`
}

// PortDescriptor is a struct to describe the characteristic of a port
type PortDescriptor struct {
	Protocol                string                              `json:"protocol,omitempty"`
	ConfigurationParameters map[string]*PortParameterDescriptor `json:"configuration_parameters,omitempty"`
}

// PortParameterDescriptor contains characteristics for every parameter
type PortParameterDescriptor struct {
	Label    string   `json:"label,omitempty"`
	Type     string   `json:"type,omitempty"`
	Values   []string `json:"value,omitempty"`
	Selected string   `json:"selected,omitempty"`
}

func (msg monitorMessage) String() string {
	s := fmt.Sprintf("type: %s", msg.EventType)
	if msg.Message != "" {
		s = fmt.Sprintf("%[1]s, message: %[2]s", s, msg.Message)
	}
	if msg.ProtocolVersion != 0 {
		s = fmt.Sprintf("%[1]s, protocol version: %[2]d", s, msg.ProtocolVersion)
	}
	if msg.PortDescription != nil {
		s = fmt.Sprintf("%s, port descriptor: protocol %s, %d parameters",
			s, msg.PortDescription.Protocol, len(msg.PortDescription.ConfigurationParameters))
	}
	return s
}

var tr = i18n.Tr

// New create and connect to the given pluggable monitor
func New(id string, args ...string) *PluggableMonitor {
	return &PluggableMonitor{
		id:          id,
		processArgs: args,
	}
}

// GetID returns the identifier for this monitor
func (mon *PluggableMonitor) GetID() string {
	return mon.id
}

func (mon *PluggableMonitor) String() string {
	return mon.id
}

func (mon *PluggableMonitor) jsonDecodeLoop(in io.Reader, outChan chan<- *monitorMessage) {
	decoder := json.NewDecoder(in)

	for {
		var msg monitorMessage
		if err := decoder.Decode(&msg); err != nil {
			mon.incomingMessagesError = err
			close(outChan)
			logrus.Errorf("stopped monitor %s decode loop", mon.id)
			return
		}
		logrus.Infof("from monitor %s received message %s", mon.id, msg)
		if msg.EventType == "port_closed" {
			// port has been closed externally...
		} else {
			outChan <- &msg
		}
	}
}

func (mon *PluggableMonitor) waitMessage(timeout time.Duration) (*monitorMessage, error) {
	select {
	case msg := <-mon.incomingMessagesChan:
		if msg == nil {
			// channel has been closed
			return nil, mon.incomingMessagesError
		}
		return msg, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf(tr("timeout waiting for message from monitor %s"), mon.id)
	}
}

func (mon *PluggableMonitor) sendCommand(command string) error {
	logrus.Infof("sending command %s to monitor %s", strings.TrimSpace(command), mon)
	data := []byte(command)
	for {
		n, err := mon.outgoingCommandsPipe.Write(data)
		if err != nil {
			return err
		}
		if n == len(data) {
			return nil
		}
		data = data[n:]
	}
}

func (mon *PluggableMonitor) runProcess() error {
	logrus.Infof("starting monitor %s process", mon.id)
	proc, err := executils.NewProcess(mon.processArgs...)
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
	mon.outgoingCommandsPipe = stdin
	mon.process = proc

	if err := mon.process.Start(); err != nil {
		return err
	}

	messageChan := make(chan *monitorMessage)
	mon.incomingMessagesChan = messageChan
	go mon.jsonDecodeLoop(stdout, messageChan)

	logrus.Infof("started monitor %s process", mon.id)
	return nil
}

func (mon *PluggableMonitor) killProcess() error {
	logrus.Infof("killing monitor %s process", mon.id)
	if err := mon.process.Kill(); err != nil {
		return err
	}
	if err := mon.process.Wait(); err != nil {
		return err
	}
	logrus.Infof("killed monitor %s process", mon.id)
	return nil
}

// Run starts the monitor executable process and sends the HELLO command to the monitor to agree on the
// pluggable monitor protocol. This must be the first command to run in the communication with the monitor.
// If the process is started but the HELLO command fails the process is killed.
func (mon *PluggableMonitor) Run() (err error) {
	if err = mon.runProcess(); err != nil {
		return err
	}

	defer func() {
		// If the monitor process is started successfully but the HELLO handshake
		// fails the monitor is an unusable state, we kill the process to avoid
		// further issues down the line.
		if err == nil {
			return
		}
		if killErr := mon.killProcess(); killErr != nil {
			// Log failure to kill the process, ideally that should never happen
			// but it's best to know it if it does
			logrus.Errorf("Killing monitor %s after unsuccessful start: %s", mon.id, killErr)
		}
	}()

	if err = mon.sendCommand("HELLO 1 \"arduino-cli " + globals.VersionInfo.VersionString + "\"\n"); err != nil {
		return err
	}
	if msg, err := mon.waitMessage(time.Second * 10); err != nil {
		return fmt.Errorf(tr("calling %[1]s: %[2]w"), "HELLO", err)
	} else if msg.EventType != "hello" {
		return errors.Errorf(tr("communication out of sync, expected 'hello', received '%s'"), msg.EventType)
	} else if msg.Message != "OK" || msg.Error {
		return errors.Errorf(tr("command failed: %s"), msg.Message)
	} else if msg.ProtocolVersion > 1 {
		return errors.Errorf(tr("protocol version not supported: requested 1, got %d"), msg.ProtocolVersion)
	}
	return nil
}

// Describe returns a description of the Port and the configuration parameters.
func (mon *PluggableMonitor) Describe() (*PortDescriptor, error) {
	if err := mon.sendCommand("DESCRIBE\n"); err != nil {
		return nil, err
	}
	if msg, err := mon.waitMessage(time.Second * 10); err != nil {
		return nil, fmt.Errorf("calling %s: %w", "", err)
	} else if msg.EventType != "describe" {
		return nil, errors.Errorf(tr("communication out of sync, expected 'describe', received '%s'"), msg.EventType)
	} else if msg.Message != "OK" || msg.Error {
		return nil, errors.Errorf(tr("command failed: %s"), msg.Message)
	} else {
		mon.supportedProtocol = msg.PortDescription.Protocol
		return msg.PortDescription, nil
	}
}

// Configure sets a port configuration parameter.
func (mon *PluggableMonitor) Configure(param, value string) error {
	if err := mon.sendCommand(fmt.Sprintf("CONFIGURE %s %s\n", param, value)); err != nil {
		return err
	}
	if msg, err := mon.waitMessage(time.Second * 10); err != nil {
		return fmt.Errorf("calling %s: %w", "", err)
	} else if msg.EventType != "configure" {
		return errors.Errorf(tr("communication out of sync, expected 'configure', received '%s'"), msg.EventType)
	} else if msg.Message != "OK" || msg.Error {
		return errors.Errorf(tr("configure failed: %s"), msg.Message)
	} else {
		return nil
	}
}

// Open connects to the given Port. A communication channel is opened
func (mon *PluggableMonitor) Open(port *rpc.Port) (io.ReadWriter, error) {
	if port.Protocol != mon.supportedProtocol {
		return nil, fmt.Errorf("invalid monitor protocol '%s': only '%s' is accepted", port.Protocol, mon.supportedProtocol)
	}

	tcpListener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return nil, err
	}
	defer tcpListener.Close()
	tcpListenerPort := tcpListener.Addr().(*net.TCPAddr).Port

	if err := mon.sendCommand(fmt.Sprintf("OPEN 127.0.0.1:%d %s\n", tcpListenerPort, port.Address)); err != nil {
		return nil, err
	}
	if msg, err := mon.waitMessage(time.Second * 10); err != nil {
		return nil, fmt.Errorf("calling %s: %w", "", err)
	} else if msg.EventType != "open" {
		return nil, errors.Errorf(tr("communication out of sync, expected 'open', received '%s'"), msg.EventType)
	} else if msg.Message != "OK" || msg.Error {
		return nil, errors.Errorf(tr("open failed: %s"), msg.Message)
	}

	conn, err := tcpListener.Accept()
	if err != nil {
		return nil, err // TODO
	}
	return conn, nil
}

// Close the communication port with the board.
func (mon *PluggableMonitor) Close() error {
	if err := mon.sendCommand("CLOSE\n"); err != nil {
		return err
	}
	if msg, err := mon.waitMessage(time.Second * 10); err != nil {
		return fmt.Errorf("calling %s: %w", "", err)
	} else if msg.EventType != "close" {
		return errors.Errorf(tr("communication out of sync, expected 'close', received '%s'"), msg.EventType)
	} else if msg.Message != "OK" || msg.Error {
		return fmt.Errorf(tr("command failed: %s"), msg.Message)
	}
	return nil
}

// Quit terminates the monitor. No more commands can be accepted by the monitor.
func (mon *PluggableMonitor) Quit() error {
	if err := mon.sendCommand("QUIT\n"); err != nil {
		return err
	}
	if msg, err := mon.waitMessage(time.Second * 10); err != nil {
		return fmt.Errorf(tr("calling %[1]s: %[2]w"), "QUIT", err)
	} else if msg.EventType != "quit" {
		return errors.Errorf(tr("communication out of sync, expected 'quit', received '%s'"), msg.EventType)
	} else if msg.Message != "OK" || msg.Error {
		return errors.Errorf(tr("command failed: %s"), msg.Message)
	}
	mon.killProcess()
	return nil
}
