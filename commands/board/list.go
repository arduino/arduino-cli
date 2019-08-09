/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package board

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/pkg/errors"
)

var (
	// ErrNotFound is returned when the API returns 404
	ErrNotFound = errors.New("board not found")
)

func apiByVidPid(url string) ([]*rpc.BoardListItem, error) {
	retVal := []*rpc.BoardListItem{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header = globals.HTTPClientHeader
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

// List FIXMEDOC
func List(instanceID int32) ([]*rpc.DetectedPort, error) {
	pm := commands.GetPackageManager(instanceID)
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	serialDiscovery, err := commands.NewBuiltinSerialDiscovery(pm)
	if err != nil {
		return nil, errors.Wrap(err, "unable to instance serial-discovery")
	}

	if err := serialDiscovery.Start(); err != nil {
		return nil, errors.Wrap(err, "unable to start serial-discovery")
	}
	defer serialDiscovery.Close()

	ports, err := serialDiscovery.List()
	if err != nil {
		return nil, errors.Wrap(err, "error getting port list from serial-discovery")
	}

	retVal := []*rpc.DetectedPort{}
	for _, port := range ports {
		b := []*rpc.BoardListItem{}

		// first query installed cores through the Package Manager
		for _, board := range pm.IdentifyBoard(port.IdentificationPrefs) {
			b = append(b, &rpc.BoardListItem{
				Name: board.Name(),
				FQBN: board.FQBN(),
			})
		}

		// if installed cores didn't recognize the board, try querying
		// the builder API
		if len(b) == 0 {
			url := fmt.Sprintf("https://builder.arduino.cc/v3/boards/byVidPid/%s/%s",
				port.IdentificationPrefs.Get("vid"),
				port.IdentificationPrefs.Get("pid"))
			items, err := apiByVidPid(url)
			if err == ErrNotFound {
				// the board couldn't be detected, keep going with the next port
				continue
			} else if err != nil {
				// this is bad, bail out
				return nil, errors.Wrap(err, "error getting bard info from Arduino Cloud")
			}

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

	return retVal, nil
}
