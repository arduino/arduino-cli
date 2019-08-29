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
	mutex     = sync.Mutex{}
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

// ListBoards foo
func ListBoards(pm *packagemanager.PackageManager) ([]*BoardPort, error) {
	// ensure the connection to the discoverer is unique to avoid messing up
	// the messages exchanged
	mutex.Lock()
	defer mutex.Unlock()

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
	args := []string{t.InstallDir.Join("serial-discovery").String()}
	cmd, err := executils.Command(args)
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
		cmd.Process.Kill()
	})
	cmd.Wait()

	return retVal, finalError
}

func getBuiltinSerialDiscoveryTool(pm *packagemanager.PackageManager) (*cores.ToolRelease, error) {
	builtinPackage := pm.Packages.GetOrCreatePackage("builtin")
	ctagsTool := builtinPackage.GetOrCreateTool("serial-discovery")
	ctagsRel := ctagsTool.GetOrCreateRelease(sdVersion)
	ctagsRel.Flavors = flavors
	return pm.Package("builtin").Tool("serial-discovery").Release(sdVersion).Get()
}
