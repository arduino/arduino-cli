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
	serialDiscoveryVersion = semver.ParseRelaxed("1.2.1")
	serialDiscoveryFlavors = []*cores.Flavor{
		{
			OS: "i686-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_32bit.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_32bit.tar.gz", serialDiscoveryVersion),
				Size:            1623562,
				Checksum:        "SHA-256:624996c2483cd66dc318e9559b9e25754180514a794acd390b4c0de58742d335",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_64bit.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_64bit.tar.gz", serialDiscoveryVersion),
				Size:            1679556,
				Checksum:        "SHA-256:fca394dde79838cdd4cbeaa4a249237c95869c7d39ec5c778ecdc76d227679ef",
				CachePath:       "tools",
			},
		},
		{
			OS: "i686-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Windows_32bit.zip", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Windows_32bit.zip", serialDiscoveryVersion),
				Size:            1735085,
				Checksum:        "SHA-256:c2d8e92e790862ee3374810121a588c9e8c6e6ff8100112912c05312e04e7570",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Windows_64bit.zip", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Windows_64bit.zip", serialDiscoveryVersion),
				Size:            1700849,
				Checksum:        "SHA-256:4da007d89bb5134e7c44a70d23b026470ee0466912db16254a2a6f431cb7d9a4",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-apple-darwin",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_macOS_64bit.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_macOS_64bit.tar.gz", serialDiscoveryVersion),
				Size:            872913,
				Checksum:        "SHA-256:c3dfb2d20b1fe839ddfba3567377200d335a93eff19b0a0f553db76d5a7e2dbd",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_ARMv6.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_ARMv6.tar.gz", serialDiscoveryVersion),
				Size:            1560586,
				Checksum:        "SHA-256:dd1687748c59ba94631f63a1f8ebee7ec21f5f40c01f5b0510f02c72c71d2d66",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm64-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_ARM64.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_ARM64.tar.gz", serialDiscoveryVersion),
				Size:            1573778,
				Checksum:        "SHA-256:3f44b5932dcfc2f01a72536d19080e45976d87c0652fcfbe75a87451327faa1f",
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
