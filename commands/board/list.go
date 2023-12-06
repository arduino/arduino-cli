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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/internal/arduino/discovery"
	"github.com/arduino/arduino-cli/internal/arduino/httpclient"
	"github.com/arduino/arduino-cli/internal/inventory"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	vidPidURL   = "https://builder.arduino.cc/v3/boards/byVidPid"
	validVidPid = regexp.MustCompile(`0[xX][a-fA-F\d]{4}`)
)

func cachedAPIByVidPid(vid, pid string) ([]*rpc.BoardListItem, error) {
	var resp []*rpc.BoardListItem

	cacheKey := fmt.Sprintf("cache.builder-api.v3/boards/byvid/pid/%s/%s", vid, pid)
	if cachedResp := inventory.Store.GetString(cacheKey + ".data"); cachedResp != "" {
		ts := inventory.Store.GetTime(cacheKey + ".ts")
		if time.Since(ts) < time.Hour*24 {
			// Use cached response
			if err := json.Unmarshal([]byte(cachedResp), &resp); err == nil {
				return resp, nil
			}
		}
	}

	resp, err := apiByVidPid(vid, pid) // Perform API requrest

	if err == nil {
		if cachedResp, err := json.Marshal(resp); err == nil {
			inventory.Store.Set(cacheKey+".data", string(cachedResp))
			inventory.Store.Set(cacheKey+".ts", time.Now())
			inventory.WriteStore()
		}
	}
	return resp, err
}

func apiByVidPid(vid, pid string) ([]*rpc.BoardListItem, error) {
	// ensure vid and pid are valid before hitting the API
	if !validVidPid.MatchString(vid) {
		return nil, errors.Errorf(tr("Invalid vid value: '%s'"), vid)
	}
	if !validVidPid.MatchString(pid) {
		return nil, errors.Errorf(tr("Invalid pid value: '%s'"), pid)
	}

	url := fmt.Sprintf("%s/%s/%s", vidPidURL, vid, pid)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")

	// TODO: use proxy if set

	httpClient, err := httpclient.New()

	if err != nil {
		return nil, errors.Wrap(err, tr("failed to initialize http client"))
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, tr("error querying Arduino Cloud Api"))
	}
	if res.StatusCode == 404 {
		// This is not an error, it just means that the board is not recognized
		return nil, nil
	}
	if res.StatusCode >= 400 {
		return nil, errors.Errorf(tr("the server responded with status %s"), res.Status)
	}

	resp, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := res.Body.Close(); err != nil {
		return nil, err
	}

	var dat map[string]interface{}
	if err := json.Unmarshal(resp, &dat); err != nil {
		return nil, errors.Wrap(err, tr("error processing response from server"))
	}
	name, nameFound := dat["name"].(string)
	fqbn, fbqnFound := dat["fqbn"].(string)
	if !nameFound || !fbqnFound {
		return nil, errors.New(tr("wrong format in server response"))
	}

	return []*rpc.BoardListItem{
		{
			Name: name,
			Fqbn: fqbn,
		},
	}, nil
}

func identifyViaCloudAPI(props *properties.Map) ([]*rpc.BoardListItem, error) {
	// If the port is not USB do not try identification via cloud
	if !props.ContainsKey("vid") || !props.ContainsKey("pid") {
		return nil, nil
	}

	logrus.Debug("Querying builder API for board identification...")
	return cachedAPIByVidPid(props.Get("vid"), props.Get("pid"))
}

// identify returns a list of boards checking first the installed platforms or the Cloud API
func identify(pme *packagemanager.Explorer, port *discovery.Port) ([]*rpc.BoardListItem, error) {
	boards := []*rpc.BoardListItem{}
	if port.Properties == nil {
		return boards, nil
	}

	// first query installed cores through the Package Manager
	logrus.Debug("Querying installed cores for board identification...")
	for _, board := range pme.IdentifyBoard(port.Properties) {
		fqbn, err := cores.ParseFQBN(board.FQBN())
		if err != nil {
			return nil, &arduino.InvalidFQBNError{Cause: err}
		}
		fqbn.Configs = board.IdentifyBoardConfiguration(port.Properties)

		// We need the Platform maintaner for sorting so we set it here
		platform := &rpc.Platform{
			Metadata: &rpc.PlatformMetadata{
				Maintainer: board.PlatformRelease.Platform.Package.Maintainer,
			},
		}
		boards = append(boards, &rpc.BoardListItem{
			Name:     board.Name(),
			Fqbn:     fqbn.String(),
			IsHidden: board.IsHidden(),
			Platform: platform,
		})
	}

	// if installed cores didn't recognize the board, try querying
	// the builder API if the board is a USB device port
	if len(boards) == 0 {
		items, err := identifyViaCloudAPI(port.Properties)
		if err != nil {
			// this is bad, but keep going
			logrus.WithError(err).Debug("Error querying builder API")
		}
		boards = items
	}

	// Sort by FQBN alphabetically
	sort.Slice(boards, func(i, j int) bool {
		return strings.ToLower(boards[i].GetFqbn()) < strings.ToLower(boards[j].GetFqbn())
	})

	// Put Arduino boards before others in case there are non Arduino boards with identical VID:PID combination
	sort.SliceStable(boards, func(i, j int) bool {
		if boards[i].GetPlatform().GetMetadata().GetMaintainer() == "Arduino" && boards[j].GetPlatform().GetMetadata().GetMaintainer() != "Arduino" {
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
func List(req *rpc.BoardListRequest) (r []*rpc.DetectedPort, discoveryStartErrors []error, e error) {
	pme, release := instances.GetPackageManagerExplorer(req.GetInstance())
	if pme == nil {
		return nil, nil, &arduino.InvalidInstanceError{}
	}
	defer release()

	var fqbnFilter *cores.FQBN
	if f := req.GetFqbn(); f != "" {
		var err error
		fqbnFilter, err = cores.ParseFQBN(f)
		if err != nil {
			return nil, nil, &arduino.InvalidFQBNError{Cause: err}
		}
	}

	dm := pme.DiscoveryManager()
	discoveryStartErrors = dm.Start()
	time.Sleep(time.Duration(req.GetTimeout()) * time.Millisecond)

	retVal := []*rpc.DetectedPort{}
	for _, port := range dm.List() {
		boards, err := identify(pme, port)
		if err != nil {
			return nil, discoveryStartErrors, err
		}

		// boards slice can be empty at this point if neither the cores nor the
		// API managed to recognize the connected board
		b := &rpc.DetectedPort{
			Port:           port.ToRPC(),
			MatchingBoards: boards,
		}

		if fqbnFilter == nil || hasMatchingBoard(b, fqbnFilter) {
			retVal = append(retVal, b)
		}
	}
	return retVal, discoveryStartErrors, nil
}

func hasMatchingBoard(b *rpc.DetectedPort, fqbnFilter *cores.FQBN) bool {
	for _, detectedBoard := range b.GetMatchingBoards() {
		detectedFqbn, err := cores.ParseFQBN(detectedBoard.GetFqbn())
		if err != nil {
			continue
		}
		if detectedFqbn.Match(fqbnFilter) {
			return true
		}
	}
	return false
}

// Watch returns a channel that receives boards connection and disconnection events.
func Watch(ctx context.Context, req *rpc.BoardListWatchRequest) (<-chan *rpc.BoardListWatchResponse, error) {
	pme, release := instances.GetPackageManagerExplorer(req.GetInstance())
	if pme == nil {
		return nil, &arduino.InvalidInstanceError{}
	}
	defer release()
	dm := pme.DiscoveryManager()

	watcher, err := dm.Watch()
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		logrus.Trace("closed watch")
		watcher.Close()
	}()

	outChan := make(chan *rpc.BoardListWatchResponse)
	go func() {
		defer close(outChan)
		for event := range watcher.Feed() {
			port := &rpc.DetectedPort{
				Port: event.Port.ToRPC(),
			}

			boardsError := ""
			if event.Type == "add" {
				boards, err := identify(pme, event.Port)
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
		}
	}()

	return outChan, nil
}
