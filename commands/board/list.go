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
	"github.com/segmentio/stats/v4"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"

	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/pkg/errors"
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
	req.Header = globals.NewHTTPClientHeader()
	req.Header.Set("Content-Type", "application/json")

	if res, err := http.DefaultClient.Do(req); err == nil {
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

// List FIXMEDOC
func List(instanceID int32) ([]*rpc.DetectedPort, error) {
	m.Lock()
	defer m.Unlock()

	pm := commands.GetPackageManager(instanceID)
	if pm == nil {
		stats.Incr("board.list", stats.Tag{"success", "false"})
		return nil, errors.New("invalid instance")
	}

	ports, err := commands.ListBoards(pm)
	if err != nil {
		stats.Incr("board.list", stats.Tag{"success", "false"})
		return nil, errors.Wrap(err, "error getting port list from serial-discovery")
	}

	retVal := []*rpc.DetectedPort{}
	for _, port := range ports {
		b := []*rpc.BoardListItem{}

		// first query installed cores through the Package Manager
		logrus.Debug("Querying installed cores for board identification...")
		for _, board := range pm.IdentifyBoard(port.IdentificationPrefs) {
			b = append(b, &rpc.BoardListItem{
				Name: board.Name(),
				FQBN: board.FQBN(),
			})
		}

		// if installed cores didn't recognize the board, try querying
		// the builder API if the board is a USB device port
		if len(b) == 0 {
			items, err := identifyViaCloudAPI(port)
			if err == ErrNotFound {
				// the board couldn't be detected, print a warning
				logrus.Debug("Board not recognized")
			} else if err != nil {
				// this is bad, bail out
				stats.Incr("board.list", stats.Tag{"success", "false"})
				return nil, errors.Wrap(err, "error getting board info from Arduino Cloud")
			}

			// add a DetectedPort entry in any case: the `Boards` field will
			// be empty but the port will be shown anyways (useful for 3rd party
			// boards)
			b = items
		}

		// boards slice can be empty at this point if neither the cores nor the
		// API managed to recognize the connected board
		p := &rpc.DetectedPort{
			Address:       port.Address,
			Protocol:      port.Protocol,
			ProtocolLabel: port.ProtocolLabel,
			Boards:        b,
		}
		retVal = append(retVal, p)
	}

	stats.Incr("board.list", stats.Tag{"success", "true"})
	return retVal, nil
}
