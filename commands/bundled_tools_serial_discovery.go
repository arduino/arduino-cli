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

package commands

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/discovery"
	"github.com/arduino/arduino-cli/arduino/resources"
	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
	semver "go.bug.st/relaxed-semver"
)

var (
	serialDiscoveryVersion = semver.ParseRelaxed("1.2.0")
	serialDiscoveryFlavors = []*cores.Flavor{
		{
			OS: "i686-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_32bit.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_32bit.tar.gz", serialDiscoveryVersion),
				Size:            1623560,
				Checksum:        "SHA-256:be545a01d60b52e2986bf6d042afee1ea910eade726ae46c04abae97ce8eb7ed",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_64bit.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_64bit.tar.gz", serialDiscoveryVersion),
				Size:            1679625,
				Checksum:        "SHA-256:215b0044dc8b09de5755299956ef72b7f998bc6cccc7a601d16c0cefb72f431f",
				CachePath:       "tools",
			},
		},
		{
			OS: "i686-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Windows_32bit.zip", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Windows_32bit.zip", serialDiscoveryVersion),
				Size:            1733449,
				Checksum:        "SHA-256:e5c1ee925ca68063e3cef91918d2831e9996e2faa987a7b9e86e715642ea7dc5",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Windows_64bit.zip", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Windows_64bit.zip", serialDiscoveryVersion),
				Size:            1698881,
				Checksum:        "SHA-256:82201b5d63e85cb815e3fa7696ae8c1a0ef958bb04c5661fe1a53362de2c7642",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-apple-darwin",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_macOS_64bit.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_macOS_64bit.tar.gz", serialDiscoveryVersion),
				Size:            869169,
				Checksum:        "SHA-256:5685cc3ece60e5eb058d10db8748201d088467d7b2df266fecb56d2cd12da6ad",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_ARMv6.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_ARMv6.tar.gz", serialDiscoveryVersion),
				Size:            1560574,
				Checksum:        "SHA-256:3e0f200cc7edce72df0f9fa1cb3fe0695466b7206a06f6c521ef5dbc3649287b",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm64-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_ARM64.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_ARM64.tar.gz", serialDiscoveryVersion),
				Size:            1573777,
				Checksum:        "SHA-256:585bc4e2043edc6824443b36d9dd561ae19a5c6686b72a25db014215a5e07f36",
				CachePath:       "tools",
			},
		},
	}
)

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

var listBoardMutex sync.Mutex

// ListBoards foo
func ListBoards(pm *packagemanager.PackageManager) ([]*BoardPort, error) {
	// ensure the connection to the discoverer is unique to avoid messing up
	// the messages exchanged
	listBoardMutex.Lock()
	defer listBoardMutex.Unlock()

	// get the bundled tool
	t, err := getBuiltinSerialDiscoveryTool(pm)
	if err != nil {
		return nil, err
	}

	// determine if it's installed
	if !t.IsInstalled() {
		return nil, fmt.Errorf("missing serial-discovery tool")
	}

	// build the command to be executed
	cmd, err := executils.NewProcessFromPath(t.InstallDir.Join("serial-discovery"))
	if err != nil {
		return nil, errors.Wrap(err, "creating discovery process")
	}

	// attach in/out pipes to the process
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdin pipe for discovery: %s", err)
	}

	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdout pipe for discovery: %s", err)
	}
	outJSON := json.NewDecoder(out)

	// start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting discovery process: %s", err)
	}

	// send the LIST command
	if _, err := in.Write([]byte("LIST\n")); err != nil {
		return nil, fmt.Errorf("sending LIST command to discovery: %s", err)
	}

	// read the response from the pipe
	decodeResult := make(chan error)
	var event eventJSON
	go func() {
		decodeResult <- outJSON.Decode(&event)
	}()

	var finalError error
	var retVal []*BoardPort

	// wait for the response
	select {
	case err := <-decodeResult:
		if err == nil {
			retVal = event.Ports
		} else {
			finalError = err
		}
	case <-time.After(10 * time.Second):
		finalError = fmt.Errorf("decoding LIST command: timeout")
	}

	// tell the process to quit
	in.Write([]byte("QUIT\n"))
	in.Close()
	out.Close()
	// kill the process if it takes too long to quit
	time.AfterFunc(time.Second, func() {
		cmd.Kill()
	})
	cmd.Wait()

	return retVal, finalError
}

// WatchListBoards returns a channel that receives events from the bundled discovery tool
func WatchListBoards(pm *packagemanager.PackageManager) (<-chan *discovery.Event, error) {
	t, err := getBuiltinSerialDiscoveryTool(pm)
	if err != nil {
		return nil, err
	}

	if !t.IsInstalled() {
		return nil, fmt.Errorf("missing serial-discovery tool")
	}

	disc, err := discovery.New("serial-discovery", t.InstallDir.Join(t.Tool.Name).String())
	if err != nil {
		return nil, err
	}

	if err = disc.Start(); err != nil {
		return nil, fmt.Errorf("starting discovery: %v", err)
	}

	if err = disc.StartSync(); err != nil {
		return nil, fmt.Errorf("starting sync: %v", err)
	}

	return disc.EventChannel(10), nil
}

func getBuiltinSerialDiscoveryTool(pm *packagemanager.PackageManager) (*cores.ToolRelease, error) {
	builtinPackage := pm.Packages.GetOrCreatePackage("builtin")
	serialDiscoveryTool := builtinPackage.GetOrCreateTool("serial-discovery")
	serialDiscoveryToolRel := serialDiscoveryTool.GetOrCreateRelease(serialDiscoveryVersion)
	serialDiscoveryToolRel.Flavors = serialDiscoveryFlavors
	return pm.Package("builtin").Tool("serial-discovery").Release(serialDiscoveryVersion).Get()
}
