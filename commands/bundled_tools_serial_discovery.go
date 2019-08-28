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
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/resources"
	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
	semver "go.bug.st/relaxed-semver"
)

var (
	sdVersion = semver.ParseRelaxed("1.0.0")
	flavors   = []*cores.Flavor{
		{
			OS: "i686-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery-linux32-v%s.tar.bz2", sdVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/tools/serial-discovery-linux32-v%s.tar.bz2", sdVersion),
				Size:            1469113,
				Checksum:        "SHA-256:35d96977844ad8d5ca9363e1ae5794450e5f7cf3d29ce7fdfe656b59e7fff725",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery-linux64-v%s.tar.bz2", sdVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/tools/serial-discovery-linux64-v%s.tar.bz2", sdVersion),
				Size:            1503971,
				Checksum:        "SHA-256:1a870d4d823ea6ebec403f63b10a1dbc9c623a6efea5cfa9141fa20045b731e2",
				CachePath:       "tools",
			},
		},
		{
			OS: "i686-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery-windows-v%s.zip", sdVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/tools/serial-discovery-windows-v%s.zip", sdVersion),
				Size:            1512379,
				Checksum:        "SHA-256:b956128ab27a3a883c938d17cad640ba396876472f2ed25d8e661f12f5d0f584",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-apple-darwin",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery-macosx-v%s.tar.bz2", sdVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/tools/serial-discovery-macosx-v%s.tar.bz2", sdVersion),
				Size:            746132,
				Checksum:        "SHA-256:fcff1b972b70a73cd738facc6d99174d8323293b60c12149c8f6f3084fb2170e",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery-linuxarm-v%s.tar.bz2", sdVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/tools/serial-discovery-linuxarm-v%s.tar.bz2", sdVersion),
				Size:            1395174,
				Checksum:        "SHA-256:f196765caa62d38475208c27b3b516e61427d5d3a8ddc6e863acb4e4a3984701",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm64-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery-linuxarm64-v%s.tar.bz2", sdVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/tools/serial-discovery-linuxarm64-v%s.tar.bz2", sdVersion),
				Size:            1402706,
				Checksum:        "SHA-256:c87010ed670254c06ac7abbc4daf7446e4e17f1945a75fc2602dd5930835dd25",
				CachePath:       "tools",
			},
		},
	}
)

// SerialDiscovery is an instance of a discovery tool
type SerialDiscovery struct {
	sync.Mutex
	ID      string
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

// NewBuiltinSerialDiscovery returns a wrapper to control the serial-discovery program
func NewBuiltinSerialDiscovery(pm *packagemanager.PackageManager) (*SerialDiscovery, error) {
	t, err := getBuiltinSerialDiscoveryTool(pm)
	if err != nil {
		return nil, err
	}

	if !t.IsInstalled() {
		return nil, fmt.Errorf("missing serial-discovery tool")
	}

	cmdArgs := []string{
		t.InstallDir.Join("serial-discovery").String(),
	}

	cmd, err := executils.Command(cmdArgs)
	if err != nil {
		return nil, errors.Wrap(err, "creating discovery process")
	}

	return &SerialDiscovery{
		ID:  strings.Join(cmdArgs, " "),
		cmd: cmd,
	}, nil
}

// Start starts the specified discovery
func (d *SerialDiscovery) start() error {
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
func (d *SerialDiscovery) List() ([]*BoardPort, error) {
	// ensure the connection to the discoverer is unique to avoid messing up
	// the messages exchanged
	d.Lock()
	defer d.Unlock()

	if err := d.start(); err != nil {
		return nil, fmt.Errorf("discovery hasn't started: %v", err)
	}

	if _, err := d.in.Write([]byte("LIST\n")); err != nil {
		return nil, fmt.Errorf("sending LIST command to discovery: %s", err)
	}
	var event eventJSON
	done := make(chan bool)
	timeout := false
	go func() {
		select {
		case <-done:
		case <-time.After(2000 * time.Millisecond):
			timeout = true
			d.close()
		}
	}()
	if err := d.outJSON.Decode(&event); err != nil {
		if timeout {
			return nil, fmt.Errorf("decoding LIST command: timeout")
		}
		return nil, fmt.Errorf("decoding LIST command: %s", err)
	}
	done <- true
	return event.Ports, d.close()
}

// Close stops the Discovery and free the resources
func (d *SerialDiscovery) close() error {
	_, _ = d.in.Write([]byte("QUIT\n"))
	_ = d.in.Close()
	_ = d.out.Close()
	timer := time.AfterFunc(time.Second, func() {
		_ = d.cmd.Process.Kill()
	})
	err := d.cmd.Wait()
	_ = timer.Stop()
	return err
}

func getBuiltinSerialDiscoveryTool(pm *packagemanager.PackageManager) (*cores.ToolRelease, error) {
	builtinPackage := pm.Packages.GetOrCreatePackage("builtin")
	ctagsTool := builtinPackage.GetOrCreateTool("serial-discovery")
	ctagsRel := ctagsTool.GetOrCreateRelease(sdVersion)
	ctagsRel.Flavors = flavors
	return pm.Package("builtin").Tool("serial-discovery").Release(sdVersion).Get()
}
