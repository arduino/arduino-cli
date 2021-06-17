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
	serialDiscoveryVersion = semver.ParseRelaxed("1.3.0-rc1")
	serialDiscoveryFlavors = []*cores.Flavor{
		{
			OS: "i686-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_32bit.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_32bit.tar.gz", serialDiscoveryVersion),
				Size:            1633143,
				Checksum:        "SHA-256:2fb17882018f3eefeaf933673cbc42cea83ce739503880ccc7f9cf521de0e513",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-pc-linux-gnu",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_64bit.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_64bit.tar.gz", serialDiscoveryVersion),
				Size:            1688362,
				Checksum:        "SHA-256:e0e55ea9c5e05f12af5d89dc3a69d63e12211f54122b4bf45a7cab9f0a6f89e5",
				CachePath:       "tools",
			},
		},
		{
			OS: "i686-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Windows_32bit.zip", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Windows_32bit.zip", serialDiscoveryVersion),
				Size:            1742668,
				Checksum:        "SHA-256:4acfe521d6fc3b29643ab69ced246d7dd20637772fc79fc3e509829c18290d90",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-mingw32",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Windows_64bit.zip", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Windows_64bit.zip", serialDiscoveryVersion),
				Size:            1709333,
				Checksum:        "SHA-256:82b2edea04f7c97b98cbb04de95ec48be95de64fa5f196d730dc824d7558b952",
				CachePath:       "tools",
			},
		},
		{
			OS: "x86_64-apple-darwin",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_macOS_64bit.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_macOS_64bit.tar.gz", serialDiscoveryVersion),
				Size:            964596,
				Checksum:        "SHA-256:ec4be0f5c1ed6af3f31bb01ed6a5433274a76a1dc7cb68d39813b2b0475d7337",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_ARMv6.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_ARMv6.tar.gz", serialDiscoveryVersion),
				Size:            1570847,
				Checksum:        "SHA-256:9341e2541ad41ee2cdaad1e8d851254c8bce63c937cdafd57db7d1439d8ced59",
				CachePath:       "tools",
			},
		},
		{
			OS: "arm64-linux-gnueabihf",
			Resource: &resources.DownloadResource{
				ArchiveFileName: fmt.Sprintf("serial-discovery_v%s_Linux_ARM64.tar.bz2", serialDiscoveryVersion),
				URL:             fmt.Sprintf("https://downloads.arduino.cc/discovery/serial-discovery/serial-discovery_v%s_Linux_ARM64.tar.gz", serialDiscoveryVersion),
				Size:            1580108,
				Checksum:        "SHA-256:1da38f94be8db69bbe26d6a95692b665f6bc9bf89aa62b58d4e4cfb0f7fd8733",
				CachePath:       "tools",
			},
		},
	}
)

// BoardPort is a generic port descriptor
type BoardPort struct {
	Address       string          `json:"address"`
	Label         string          `json:"label"`
	Properties    *properties.Map `json:"properties"`
	Protocol      string          `json:"protocol"`
	ProtocolLabel string          `json:"protocolLabel"`
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

	if err = disc.Hello(); err != nil {
		return nil, fmt.Errorf("starting discovery: %v", err)
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
