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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/internal/cli/configuration"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/arduino-cli/internal/inventory"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-properties-orderedmap"
	discovery "github.com/arduino/pluggable-discovery-protocol-handler/v2"
	"github.com/sirupsen/logrus"
)

var (
	vidPidURL   = "https://builder.arduino.cc/v3/boards/byVidPid"
	validVidPid = regexp.MustCompile(`0[xX][a-fA-F\d]{4}`)
)

func cachedAPIByVidPid(vid, pid string, settings *configuration.Settings) ([]*rpc.BoardListItem, error) {
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

	resp, err := apiByVidPid(vid, pid, settings) // Perform API requrest

	if err == nil {
		if cachedResp, err := json.Marshal(resp); err == nil {
			inventory.Store.Set(cacheKey+".data", string(cachedResp))
			inventory.Store.Set(cacheKey+".ts", time.Now())
			inventory.WriteStore()
		}
	}
	return resp, err
}

func apiByVidPid(vid, pid string, settings *configuration.Settings) ([]*rpc.BoardListItem, error) {
	// ensure vid and pid are valid before hitting the API
	if !validVidPid.MatchString(vid) {
		return nil, errors.New(i18n.Tr("Invalid vid value: '%s'", vid))
	}
	if !validVidPid.MatchString(pid) {
		return nil, errors.New(i18n.Tr("Invalid pid value: '%s'", pid))
	}

	url := fmt.Sprintf("%s/%s/%s", vidPidURL, vid, pid)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")

	httpClient, err := settings.NewHttpClient()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.Tr("failed to initialize http client"), err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.Tr("error querying Arduino Cloud Api"), err)
	}
	if res.StatusCode == 404 {
		// This is not an error, it just means that the board is not recognized
		return nil, nil
	}
	if res.StatusCode >= 400 {
		return nil, errors.New(i18n.Tr("the server responded with status %s", res.Status))
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
		return nil, fmt.Errorf("%s: %w", i18n.Tr("error processing response from server"), err)
	}
	name, nameFound := dat["name"].(string)
	fqbn, fbqnFound := dat["fqbn"].(string)
	if !nameFound || !fbqnFound {
		return nil, errors.New(i18n.Tr("wrong format in server response"))
	}

	return []*rpc.BoardListItem{
		{
			Name: name,
			Fqbn: fqbn,
		},
	}, nil
}

func identifyViaCloudAPI(props *properties.Map, settings *configuration.Settings) ([]*rpc.BoardListItem, error) {
	// If the port is not USB do not try identification via cloud
	if !props.ContainsKey("vid") || !props.ContainsKey("pid") {
		return nil, nil
	}

	logrus.Debug("Querying builder API for board identification...")
	return cachedAPIByVidPid(props.Get("vid"), props.Get("pid"), settings)
}

// identify returns a list of boards checking first the installed platforms or the Cloud API
func identify(pme *packagemanager.Explorer, port *discovery.Port, settings *configuration.Settings) ([]*rpc.BoardListItem, error) {
	boards := []*rpc.BoardListItem{}
	if port.Properties == nil {
		return boards, nil
	}

	// first query installed cores through the Package Manager
	logrus.Debug("Querying installed cores for board identification...")
	for _, board := range pme.IdentifyBoard(port.Properties) {
		fqbn, err := cores.ParseFQBN(board.FQBN())
		if err != nil {
			return nil, &cmderrors.InvalidFQBNError{Cause: err}
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
		items, err := identifyViaCloudAPI(port.Properties, settings)
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

// BoardList returns a list of boards found by the loaded discoveries.
// In case of errors partial results from discoveries that didn't fail
// are returned.
func (s *arduinoCoreServerImpl) BoardList(ctx context.Context, req *rpc.BoardListRequest) (*rpc.BoardListResponse, error) {
	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return nil, err
	}
	defer release()

	var fqbnFilter *cores.FQBN
	if f := req.GetFqbn(); f != "" {
		var err error
		fqbnFilter, err = cores.ParseFQBN(f)
		if err != nil {
			return nil, &cmderrors.InvalidFQBNError{Cause: err}
		}
	}

	dm := pme.DiscoveryManager()
	warnings := f.Map(dm.Start(), (error).Error)
	time.Sleep(time.Duration(req.GetTimeout()) * time.Millisecond)

	ports := []*rpc.DetectedPort{}
	for _, port := range dm.List() {
		boards, err := identify(pme, port, s.settings)
		if err != nil {
			warnings = append(warnings, err.Error())
		}

		// boards slice can be empty at this point if neither the cores nor the
		// API managed to recognize the connected board
		b := &rpc.DetectedPort{
			Port:           rpc.DiscoveryPortToRPC(port),
			MatchingBoards: boards,
		}

		if fqbnFilter == nil || hasMatchingBoard(b, fqbnFilter) {
			ports = append(ports, b)
		}
	}
	return &rpc.BoardListResponse{
		Ports:    ports,
		Warnings: warnings,
	}, nil
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

// BoardListWatchProxyToChan return a stream, to be used in BoardListWatch method,
// that proxies all the responses to a channel.
func BoardListWatchProxyToChan(ctx context.Context) (rpc.ArduinoCoreService_BoardListWatchServer, <-chan *rpc.BoardListWatchResponse) {
	return streamResponseToChan[rpc.BoardListWatchResponse](ctx)
}

// BoardListWatch FIXMEDOC
func (s *arduinoCoreServerImpl) BoardListWatch(req *rpc.BoardListWatchRequest, stream rpc.ArduinoCoreService_BoardListWatchServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	if req.GetInstance() == nil {
		err := errors.New(i18n.Tr("no instance specified"))
		syncSend.Send(&rpc.BoardListWatchResponse{
			EventType: "error",
			Error:     err.Error(),
		})
		return err
	}

	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return err
	}
	defer release()
	dm := pme.DiscoveryManager()

	watcher, err := dm.Watch()
	if err != nil {
		return err
	}

	go func() {
		<-stream.Context().Done()
		logrus.Trace("closed watch")
		watcher.Close()
	}()

	go func() {
		for event := range watcher.Feed() {
			port := &rpc.DetectedPort{
				Port: rpc.DiscoveryPortToRPC(event.Port),
			}

			boardsError := ""
			if event.Type == "add" {
				boards, err := identify(pme, event.Port, s.settings)
				if err != nil {
					boardsError = err.Error()
				}
				port.MatchingBoards = boards
			}
			stream.Send(&rpc.BoardListWatchResponse{
				EventType: event.Type,
				Port:      port,
				Error:     boardsError,
			})
		}
	}()

	return nil
}
