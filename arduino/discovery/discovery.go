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

package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	properties "github.com/arduino/go-properties-orderedmap"

	"github.com/arduino/arduino-cli/executils"
)

// Discovery is an instance of a discovery tool
type Discovery struct {
	in      io.WriteCloser
	out     io.ReadCloser
	outJSON *json.Decoder
	cmd     *exec.Cmd
}

// BoardPort is a generic port descriptor
type BoardPort struct {
	Address             string          `json:"address"`
	Label               string          `json:"label"`
	Prefs               *properties.Map `json:"prefs"`
	IdentificationPrefs *properties.Map `json:"identificationPrefs"`
	Protocol            string          `json:"protocol"`
	ProtocolLabel       string          `json:"protocolLabel"`
}

type eventJSON struct {
	EventType string       `json:"eventType,required"`
	Ports     []*BoardPort `json:"ports"`
}

// NewFromCommandLine creates a new Discovery object
func NewFromCommandLine(args ...string) (*Discovery, error) {
	cmd, err := executils.Command(args)
	if err != nil {
		return nil, fmt.Errorf("creating discovery process: %s", err)
	}
	disc := &Discovery{}
	disc.cmd = cmd
	return disc, nil
}

// Start starts the specified discovery
func (d *Discovery) Start() error {
	if in, err := d.cmd.StdinPipe(); err == nil {
		d.in = in
	} else {
		return fmt.Errorf("creating stdin pipe for discovery: %s", err)
	}
	if out, err := d.cmd.StdoutPipe(); err == nil {
		d.out = out
		d.outJSON = json.NewDecoder(d.out)
	} else {
		return fmt.Errorf("creating stdout pipe for discovery: %s", err)
	}
	if err := d.cmd.Start(); err != nil {
		return fmt.Errorf("starting discovery process: %s", err)
	}
	return nil
}

// List retrieve the port list from this discovery
func (d *Discovery) List() ([]*BoardPort, error) {
	if _, err := d.in.Write([]byte("LIST\n")); err != nil {
		return nil, fmt.Errorf("sending LIST command to discovery: %s", err)
	}
	var event eventJSON
	if err := d.outJSON.Decode(&event); err != nil {
		return nil, fmt.Errorf("decoding LIST command: %s", err)
	}
	return event.Ports, nil
}

// Close stops the Discovery and free the resources
func (d *Discovery) Close() error {
	// TODO: Send QUIT for safe close or terminate process after a small timeout
	if err := d.in.Close(); err != nil {
		return fmt.Errorf("closing stdin pipe: %s", err)
	}
	if err := d.out.Close(); err != nil {
		return fmt.Errorf("closing stdout pipe: %s", err)
	}
	if d.cmd != nil {
		d.cmd.Process.Kill()
	}
	return nil
}
