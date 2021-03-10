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

package board

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/httpclient"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/pkg/errors"
	"github.com/segmentio/stats/v4"
	"github.com/sirupsen/logrus"
)

var (
	// ErrNotFound is returned when the API returns 404
	ErrNotFound = errors.New("board not found")
	m           sync.Mutex
	vidPidURL   = "https://builder.arduino.cc/v3/boards/byVidPid"
	validVidPid = regexp.MustCompile(`0[xX][a-fA-F\d]{4}`)
)

func apiByVidPid(vid, pid string) ([]*rpc.BoardListItem, error) {
	// ensure vid and pid are valid before hitting the API
	if !validVidPid.MatchString(vid) {
		return nil, errors.Errorf("Invalid vid value: '%s'", vid)
	}
	if !validVidPid.MatchString(pid) {
		return nil, errors.Errorf("Invalid pid value: '%s'", pid)
	}

	url := fmt.Sprintf("%s/%s/%s", vidPidURL, vid, pid)
	retVal := []*rpc.BoardListItem{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")

	// TODO: use proxy if set

	httpClient, err := httpclient.New()

	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize http client")
	}

	if res, err := httpClient.Do(req); err == nil {
		if res.StatusCode >= 400 {
			if res.StatusCode == 404 {
				return nil, ErrNotFound
			}
			return nil, errors.Errorf("the server responded with status %s", res.Status)
		}

		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()

		var dat map[string]interface{}
		err = json.Unmarshal(body, &dat)
		if err != nil {
			return nil, errors.Wrap(err, "error processing response from server")
		}

		name, nameFound := dat["name"].(string)
		fqbn, fbqnFound := dat["fqbn"].(string)

		if !nameFound || !fbqnFound {
			return nil, errors.New("wrong format in server response")
		}

		retVal = append(retVal, &rpc.BoardListItem{
			Name: name,
			FQBN: fqbn,
			VID:  vid,
			PID:  pid,
		})
	} else {
		return nil, errors.Wrap(err, "error querying Arduino Cloud Api")
	}

	return retVal, nil
}

func identifyViaCloudAPI(port *commands.BoardPort) ([]*rpc.BoardListItem, error) {
	// If the port is not USB do not try identification via cloud
	id := port.IdentificationPrefs
	if !id.ContainsKey("vid") || !id.ContainsKey("pid") {
		return nil, ErrNotFound
	}

	logrus.Debug("Querying builder API for board identification...")
	return apiByVidPid(id.Get("vid"), id.Get("pid"))
}

// identify returns a list of boards checking first the installed platforms or the Cloud API
func identify(pm *packagemanager.PackageManager, port *commands.BoardPort) ([]*rpc.BoardListItem, error) {
	boards := []*rpc.BoardListItem{}

	// first query installed cores through the Package Manager
	logrus.Debug("Querying installed cores for board identification...")
	for _, board := range pm.IdentifyBoard(port.IdentificationPrefs) {
		boards = append(boards, &rpc.BoardListItem{
			Name: board.Name(),
			FQBN: board.FQBN(),
			VID:  port.IdentificationPrefs.Get("vid"),
			PID:  port.IdentificationPrefs.Get("pid"),
		})
	}

	// if installed cores didn't recognize the board, try querying
	// the builder API if the board is a USB device port
	if len(boards) == 0 {
		items, err := identifyViaCloudAPI(port)
		if err == ErrNotFound {
			// the board couldn't be detected, print a warning
			logrus.Debug("Board not recognized")
		} else if err != nil {
			// this is bad, bail out
			return nil, errors.Wrap(err, "error getting board info from Arduino Cloud")
		}

		// add a DetectedPort entry in any case: the `Boards` field will
		// be empty but the port will be shown anyways (useful for 3rd party
		// boards)
		boards = items
	}
	return boards, nil
}

// List FIXMEDOC
func List(instanceID int32) (r []*rpc.DetectedPort, e error) {
	m.Lock()
	defer m.Unlock()

	tags := map[string]string{}
	// Use defer func() to evaluate tags map when function returns
	// and set success flag inspecting the error named return parameter
	defer func() {
		tags["success"] = "true"
		if e != nil {
			tags["success"] = "false"
		}
		stats.Incr("compile", stats.M(tags)...)
	}()

	pm := commands.GetPackageManager(instanceID)
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	ports, err := commands.ListBoards(pm)
	if err != nil {
		return nil, errors.Wrap(err, "error getting port list from serial-discovery")
	}

	retVal := []*rpc.DetectedPort{}
	for _, port := range ports {
		boards, err := identify(pm, port)
		if err != nil {
			return nil, err
		}

		// boards slice can be empty at this point if neither the cores nor the
		// API managed to recognize the connected board
		p := &rpc.DetectedPort{
			Address:       port.Address,
			Protocol:      port.Protocol,
			ProtocolLabel: port.ProtocolLabel,
			Boards:        boards,
			SerialNumber:  port.Prefs.Get("serialNumber"),
		}
		retVal = append(retVal, p)
	}

	return retVal, nil
}

// Watch returns a channel that receives boards connection and disconnection events.
// The discovery process can be interrupted by sending a message to the interrupt channel.
func Watch(instanceID int32, interrupt <-chan bool) (<-chan *rpc.BoardListWatchResp, error) {
	pm := commands.GetPackageManager(instanceID)
	eventsChan, err := commands.WatchListBoards(pm)
	if err != nil {
		return nil, err
	}

	outChan := make(chan *rpc.BoardListWatchResp)
	go func() {
		for {
			select {
			case event := <-eventsChan:
				boards := []*rpc.BoardListItem{}
				boardsError := ""
				if event.Type == "add" {
					boards, err = identify(pm, &commands.BoardPort{
						Address:             event.Port.Address,
						Label:               event.Port.AddressLabel,
						Prefs:               event.Port.Properties,
						IdentificationPrefs: event.Port.IdentificationProperties,
						Protocol:            event.Port.Protocol,
						ProtocolLabel:       event.Port.ProtocolLabel,
					})
					if err != nil {
						boardsError = err.Error()
					}
				}

				serialNumber := ""
				if props := event.Port.Properties; props != nil {
					serialNumber = props.Get("serialNumber")
				}

				outChan <- &rpc.BoardListWatchResp{
					EventType: event.Type,
					Port: &rpc.DetectedPort{
						Address:       event.Port.Address,
						Protocol:      event.Port.Protocol,
						ProtocolLabel: event.Port.ProtocolLabel,
						Boards:        boards,
						SerialNumber:  serialNumber,
					},
					Error: boardsError,
				}
			case <-interrupt:
				break
			}
		}
	}()

	return outChan, nil
}
