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
	"sort"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/discovery"
	"github.com/arduino/arduino-cli/arduino/httpclient"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/pkg/errors"
	"github.com/segmentio/stats/v4"
	"github.com/sirupsen/logrus"
)

type boardNotFoundError struct{}

func (e *boardNotFoundError) Error() string {
	return tr("board not found")
}

var (
	// ErrNotFound is returned when the API returns 404
	ErrNotFound = &boardNotFoundError{}
	vidPidURL   = "https://builder.arduino.cc/v3/boards/byVidPid"
	validVidPid = regexp.MustCompile(`0[xX][a-fA-F\d]{4}`)
)

func apiByVidPid(vid, pid string) ([]*rpc.BoardListItem, error) {
	// ensure vid and pid are valid before hitting the API
	if !validVidPid.MatchString(vid) {
		return nil, errors.Errorf(tr("Invalid vid value: '%s'"), vid)
	}
	if !validVidPid.MatchString(pid) {
		return nil, errors.Errorf(tr("Invalid pid value: '%s'"), pid)
	}

	url := fmt.Sprintf("%s/%s/%s", vidPidURL, vid, pid)
	retVal := []*rpc.BoardListItem{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")

	// TODO: use proxy if set

	httpClient, err := httpclient.New()

	if err != nil {
		return nil, errors.Wrap(err, tr("failed to initialize http client"))
	}

	if res, err := httpClient.Do(req); err == nil {
		if res.StatusCode >= 400 {
			if res.StatusCode == 404 {
				return nil, ErrNotFound
			}
			return nil, errors.Errorf(tr("the server responded with status %s"), res.Status)
		}

		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()

		var dat map[string]interface{}
		err = json.Unmarshal(body, &dat)
		if err != nil {
			return nil, errors.Wrap(err, tr("error processing response from server"))
		}

		name, nameFound := dat["name"].(string)
		fqbn, fbqnFound := dat["fqbn"].(string)

		if !nameFound || !fbqnFound {
			return nil, errors.New(tr("wrong format in server response"))
		}

		retVal = append(retVal, &rpc.BoardListItem{
			Name: name,
			Fqbn: fqbn,
		})
	} else {
		return nil, errors.Wrap(err, tr("error querying Arduino Cloud Api"))
	}

	return retVal, nil
}

func identifyViaCloudAPI(port *discovery.Port) ([]*rpc.BoardListItem, error) {
	// If the port is not USB do not try identification via cloud
	id := port.Properties
	if !id.ContainsKey("vid") || !id.ContainsKey("pid") {
		return nil, ErrNotFound
	}

	logrus.Debug("Querying builder API for board identification...")
	return apiByVidPid(id.Get("vid"), id.Get("pid"))
}

// identify returns a list of boards checking first the installed platforms or the Cloud API
func identify(pm *packagemanager.PackageManager, port *discovery.Port) ([]*rpc.BoardListItem, error) {
	boards := []*rpc.BoardListItem{}

	// first query installed cores through the Package Manager
	logrus.Debug("Querying installed cores for board identification...")
	for _, board := range pm.IdentifyBoard(port.Properties) {
		// We need the Platform maintaner for sorting so we set it here
		platform := &rpc.Platform{
			Maintainer: board.PlatformRelease.Platform.Package.Maintainer,
		}
		boards = append(boards, &rpc.BoardListItem{
			Name:     board.Name(),
			Fqbn:     board.FQBN(),
			Platform: platform,
		})
	}

	// if installed cores didn't recognize the board, try querying
	// the builder API if the board is a USB device port
	if len(boards) == 0 {
		items, err := identifyViaCloudAPI(port)
		if errors.Is(err, ErrNotFound) {
			// the board couldn't be detected, print a warning
			logrus.Debug("Board not recognized")
		} else if err != nil {
			// this is bad, bail out
			return nil, &arduino.UnavailableError{Message: tr("Error getting board info from Arduino Cloud")}
		}

		// add a DetectedPort entry in any case: the `Boards` field will
		// be empty but the port will be shown anyways (useful for 3rd party
		// boards)
		boards = items
	}

	// Sort by FQBN alphabetically
	sort.Slice(boards, func(i, j int) bool {
		return strings.ToLower(boards[i].Fqbn) < strings.ToLower(boards[j].Fqbn)
	})

	// Put Arduino boards before others in case there are non Arduino boards with identical VID:PID combination
	sort.SliceStable(boards, func(i, j int) bool {
		if boards[i].Platform.Maintainer == "Arduino" && boards[j].Platform.Maintainer != "Arduino" {
			return true
		}
		return false
	})

	// We need the Board's Platform only for sorting but it shouldn't be present in the output
	for _, board := range boards {
		board.Platform = nil
	}

	return boards, nil
}

// List returns a list of boards found by the loaded discoveries.
// In case of errors partial results from discoveries that didn't fail
// are returned.
func List(req *rpc.BoardListRequest) (r []*rpc.DetectedPort, e error) {
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

	pm := commands.GetPackageManager(req.GetInstance().Id)
	if pm == nil {
		return nil, &arduino.InvalidInstanceError{}
	}

	dm := pm.DiscoveryManager()
	if errs := dm.RunAll(); len(errs) > 0 {
		return nil, &arduino.UnavailableError{Message: tr("Error starting board discoveries"), Cause: fmt.Errorf("%v", errs)}
	}
	if errs := dm.StartAll(); len(errs) > 0 {
		return nil, &arduino.UnavailableError{Message: tr("Error starting board discoveries"), Cause: fmt.Errorf("%v", errs)}
	}
	defer func() {
		if errs := dm.StopAll(); len(errs) > 0 {
			logrus.Error(errs)
		}
	}()
	time.Sleep(time.Duration(req.GetTimeout()) * time.Millisecond)

	retVal := []*rpc.DetectedPort{}
	ports, errs := pm.DiscoveryManager().List()
	for _, port := range ports {
		boards, err := identify(pm, port)
		if err != nil {
			return nil, err
		}

		// boards slice can be empty at this point if neither the cores nor the
		// API managed to recognize the connected board
		b := &rpc.DetectedPort{
			Port:           port.ToRPC(),
			MatchingBoards: boards,
		}
		retVal = append(retVal, b)
	}
	if len(errs) > 0 {
		return retVal, &arduino.UnavailableError{Message: tr("Error getting board list"), Cause: fmt.Errorf("%v", errs)}
	}
	return retVal, nil
}

// Watch returns a channel that receives boards connection and disconnection events.
// The discovery process can be interrupted by sending a message to the interrupt channel.
func Watch(instanceID int32, interrupt <-chan bool) (<-chan *rpc.BoardListWatchResponse, error) {
	pm := commands.GetPackageManager(instanceID)
	dm := pm.DiscoveryManager()

	runErrs := dm.RunAll()
	if len(runErrs) == len(dm.IDs()) {
		// All discoveries failed to run, we can't do anything
		return nil, &arduino.UnavailableError{Message: tr("Error starting board discoveries"), Cause: fmt.Errorf("%v", runErrs)}
	}

	eventsChan, errs := dm.StartSyncAll()
	if len(runErrs) > 0 {
		errs = append(runErrs, errs...)
	}

	outChan := make(chan *rpc.BoardListWatchResponse)

	go func() {
		defer close(outChan)
		for _, err := range errs {
			outChan <- &rpc.BoardListWatchResponse{
				EventType: "error",
				Error:     err.Error(),
			}
		}
		for {
			select {
			case event := <-eventsChan:
				if event.Type == "quit" {
					// The discovery manager has closed its event channel because it's
					// quitting all the discovery processes that are running, this
					// means that the events channel we're listening from won't receive any
					// more events.
					// Handling this case is necessary when the board watcher is running and
					// the instance being used is reinitialized since that quits all the
					// discovery processes and reset the discovery manager. That would leave
					// this goroutine listening forever on a "dead" channel and might even
					// cause panics.
					// This message avoid all this issues.
					// It will be the client's task restarting the board watcher if necessary,
					// this host won't attempt restarting it.
					outChan <- &rpc.BoardListWatchResponse{
						EventType: event.Type,
					}
					return
				}

				port := &rpc.DetectedPort{
					Port: event.Port.ToRPC(),
				}

				boardsError := ""
				if event.Type == "add" {
					boards, err := identify(pm, event.Port)
					if err != nil {
						boardsError = err.Error()
					}
					port.MatchingBoards = boards
				}
				outChan <- &rpc.BoardListWatchResponse{
					EventType: event.Type,
					Port:      port,
					Error:     boardsError,
				}
			case <-interrupt:
				for _, err := range dm.StopAll() {
					// Discoveries that return errors have their process
					// closed and are removed from the list of discoveries
					// in the manager
					outChan <- &rpc.BoardListWatchResponse{
						EventType: "error",
						Error:     tr("stopping discoveries: %s", err),
					}
				}
				return
			}
		}
	}()

	return outChan, nil
}
