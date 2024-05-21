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

// Package monitor provides a client for Pluggable Monitors.
// Documentation is available here:
// https://arduino.github.io/arduino-cli/latest/pluggable-monitor-specification/
package monitor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/arduino-cli/version"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// PluggableMonitor is a tool that communicates with a board through a communication port.
type PluggableMonitor struct {
	id                   string
	processArgs          []string
	process              *paths.Process
	outgoingCommandsPipe io.Writer
	incomingMessagesChan <-chan *monitorMessage
	supportedProtocol    string
	log                  *logrus.Entry

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

// New create and connect to the given pluggable monitor
func New(id string, args ...string) *PluggableMonitor {
	return &PluggableMonitor{
		id:          id,
		processArgs: args,
		log:         logrus.WithField("monitor", id),
	}
}

// GetID returns the identifier for this monitor
func (mon *PluggableMonitor) GetID() string {
	return mon.id
}

func (mon *PluggableMonitor) String() string {
	return mon.id
}

func jsonDecodeLoop(in io.Reader, outChan chan<- *monitorMessage, log *logrus.Entry, lastError *error) {
	decoder := json.NewDecoder(in)

	for {
		var msg monitorMessage
		if err := decoder.Decode(&msg); err != nil {
			*lastError = err
			close(outChan)
			log.Errorf("stopped decode loop: %s", err)
			return
		}
		log.WithField("event_type", msg.EventType).
			WithField("message", msg.Message).
			WithField("error", msg.Error).
			Infof("received message")
		if msg.EventType == "port_closed" {
			log.Infof("monitor port has been closed externally")
		} else {
			outChan <- &msg
		}
	}
}

func (mon *PluggableMonitor) waitMessage(timeout time.Duration, expectedEvt string) (*monitorMessage, error) {
	mon.log.WithField("expected", expectedEvt).Debugf("waiting for event")
	var msg *monitorMessage
	select {
	case m, ok := <-mon.incomingMessagesChan:
		if !ok {
			// channel has been closed
			return nil, mon.incomingMessagesError
		}
		msg = m
	case <-time.After(timeout):
		return nil, errors.New(i18n.Tr("timeout waiting for message"))
	}
	if expectedEvt == "" {
		// No message processing required for this call
		return msg, nil
	}
	if msg.EventType != expectedEvt {
		return msg, errors.New(i18n.Tr("communication out of sync, expected '%[1]s', received '%[2]s'", expectedEvt, msg.EventType))
	}
	if msg.Error {
		return msg, errors.New(i18n.Tr("command '%[1]s' failed: %[2]s", expectedEvt, msg.Message))
	}
	if strings.ToUpper(msg.Message) != "OK" {
		return msg, errors.New(i18n.Tr("communication out of sync, expected '%[1]s', received '%[2]s'", "OK", msg.Message))
	}
	return msg, nil
}

func (mon *PluggableMonitor) sendCommand(command string) error {
	mon.log.WithField("command", strings.TrimSpace(command)).Infof("sending command")
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
	mon.log.Infof("Starting monitor process")
	proc, err := paths.NewProcess(nil, mon.processArgs...)
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
	go jsonDecodeLoop(stdout, messageChan, mon.log, &mon.incomingMessagesError)

	mon.log.Infof("Monitor process started successfully!")
	return nil
}

func (mon *PluggableMonitor) killProcess() {
	mon.log.Infof("Killing monitor process")
	if err := mon.process.Kill(); err != nil {
		mon.log.WithError(err).Error("Sent kill signal")
	}
	if err := mon.process.Wait(); err != nil {
		mon.log.WithError(err).Error("Waiting for process end")
	}
	mon.log.Infof("Monitor process killed")
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
		// fails the monitor is in an unusable state, we kill the process to avoid
		// further issues down the line.
		if err != nil {
			mon.killProcess()
		}
	}()

	if err = mon.sendCommand("HELLO 1 \"arduino-cli " + version.VersionInfo.VersionString + "\"\n"); err != nil {
		return err
	}
	if msg, err := mon.waitMessage(time.Second*10, "hello"); err != nil {
		return err
	} else if msg.ProtocolVersion > 1 {
		return errors.New(i18n.Tr("protocol version not supported: requested %[1]d, got %[2]d", 1, msg.ProtocolVersion))
	}
	return nil
}

// Describe returns a description of the Port and the configuration parameters.
func (mon *PluggableMonitor) Describe() (*PortDescriptor, error) {
	if err := mon.sendCommand("DESCRIBE\n"); err != nil {
		return nil, err
	}
	msg, err := mon.waitMessage(time.Second*10, "describe")
	if err != nil {
		return nil, err
	}
	mon.supportedProtocol = msg.PortDescription.Protocol
	return msg.PortDescription, nil
}

// Configure sets a port configuration parameter.
func (mon *PluggableMonitor) Configure(param, value string) error {
	if err := mon.sendCommand(fmt.Sprintf("CONFIGURE %s %s\n", param, value)); err != nil {
		return err
	}
	_, err := mon.waitMessage(time.Second*10, "configure")
	return err
}

// Open connects to the given Port. A communication channel is opened
func (mon *PluggableMonitor) Open(portAddress, portProtocol string) (io.ReadWriter, error) {
	if portProtocol != mon.supportedProtocol {
		return nil, fmt.Errorf("invalid monitor protocol '%s': only '%s' is accepted", portProtocol, mon.supportedProtocol)
	}

	tcpListener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return nil, err
	}
	defer tcpListener.Close()
	tcpListenerPort := tcpListener.Addr().(*net.TCPAddr).Port

	if err := mon.sendCommand(fmt.Sprintf("OPEN 127.0.0.1:%d %s\n", tcpListenerPort, portAddress)); err != nil {
		return nil, err
	}
	if _, err := mon.waitMessage(time.Second*10, "open"); err != nil {
		return nil, err
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
	_, err := mon.waitMessage(time.Millisecond*5000, "close")
	return err
}

// Quit terminates the monitor. No more commands can be accepted by the monitor.
func (mon *PluggableMonitor) Quit() error {
	defer mon.killProcess() // ensure that killProcess is called in any case...

	if err := mon.sendCommand("QUIT\n"); err != nil {
		return err
	}
	if _, err := mon.waitMessage(time.Millisecond*250, "quit"); err != nil {
		return err
	}
	return nil
}
