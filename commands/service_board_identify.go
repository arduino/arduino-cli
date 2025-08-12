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
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/internal/cli/configuration"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/arduino-cli/internal/inventory"
	"github.com/arduino/arduino-cli/pkg/fqbn"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
)

// BoardIdentify identifies the board based on the provided properties
func (s *arduinoCoreServerImpl) BoardIdentify(ctx context.Context, req *rpc.BoardIdentifyRequest) (*rpc.BoardIdentifyResponse, error) {
	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return nil, err
	}
	defer release()

	props := properties.NewFromHashmap(req.GetProperties())
	res, err := identify(ctx, pme, props, s.settings, !req.GetUseCloudApiForUnknownBoardDetection())
	if err != nil {
		return nil, err
	}
	return &rpc.BoardIdentifyResponse{
		Boards: res,
	}, nil
}

// identify returns a list of boards checking first the installed platforms or the Cloud API
func identify(ctx context.Context, pme *packagemanager.Explorer, properties *properties.Map, settings *configuration.Settings, skipCloudAPI bool) ([]*rpc.BoardListItem, error) {
	if properties == nil {
		return nil, nil
	}

	// first query installed cores through the Package Manager
	boards := []*rpc.BoardListItem{}
	logrus.Debug("Querying installed cores for board identification...")
	for _, board := range pme.IdentifyBoard(properties) {
		fqbn, err := fqbn.Parse(board.FQBN())
		if err != nil {
			return nil, &cmderrors.InvalidFQBNError{Cause: err}
		}
		fqbn.Configs = board.IdentifyBoardConfiguration(properties)

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
	if len(boards) == 0 && !skipCloudAPI && !settings.SkipCloudApiForBoardDetection() {
		items, err := identifyViaCloudAPI(ctx, properties, settings)
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

func identifyViaCloudAPI(ctx context.Context, props *properties.Map, settings *configuration.Settings) ([]*rpc.BoardListItem, error) {
	// If the port is not USB do not try identification via cloud
	if !props.ContainsKey("vid") || !props.ContainsKey("pid") {
		return nil, nil
	}

	logrus.Debug("Querying builder API for board identification...")
	return cachedAPIByVidPid(ctx, props.Get("vid"), props.Get("pid"), settings)
}

var (
	vidPidURL   = "https://api2.arduino.cc/boards/v1/boards"
	validVidPid = regexp.MustCompile(`0[xX][a-fA-F\d]{4}`)
)

func cachedAPIByVidPid(ctx context.Context, vid, pid string, settings *configuration.Settings) ([]*rpc.BoardListItem, error) {
	var resp []*rpc.BoardListItem

	cacheKey := fmt.Sprintf("cache.api2.arduino.cc/boards/v1/boards?vid-pid=%s-%s", vid, pid)
	if cachedResp := inventory.Store.GetString(cacheKey + ".data"); cachedResp != "" {
		ts := inventory.Store.GetTime(cacheKey + ".ts")
		if time.Since(ts) < time.Hour*24 {
			// Use cached response
			if err := json.Unmarshal([]byte(cachedResp), &resp); err == nil {
				return resp, nil
			}
		}
	}

	resp, err := apiByVidPid(ctx, vid, pid, settings) // Perform API requrest

	if err == nil {
		if cachedResp, err := json.Marshal(resp); err == nil {
			inventory.Store.Set(cacheKey+".data", string(cachedResp))
			inventory.Store.Set(cacheKey+".ts", time.Now())
			inventory.WriteStore()
		}
	}
	return resp, err
}

func apiByVidPid(ctx context.Context, vid, pid string, settings *configuration.Settings) ([]*rpc.BoardListItem, error) {
	// ensure vid and pid are valid before hitting the API
	if !validVidPid.MatchString(vid) {
		return nil, errors.New(i18n.Tr("Invalid vid value: '%s'", vid))
	}
	if !validVidPid.MatchString(pid) {
		return nil, errors.New(i18n.Tr("Invalid pid value: '%s'", pid))
	}

	url := fmt.Sprintf("%s?vid-pid=%s-%s", vidPidURL, vid, pid)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")

	httpClient, err := settings.NewHttpClient(ctx)
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

	type boardsResponse struct {
		Items []struct {
			Name string `json:"name"`
			FQBN string `json:"fqbn"`
		} `json:"items"`
	}
	var dat boardsResponse
	if err := json.Unmarshal(resp, &dat); err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.Tr("error processing response from server"), err)
	}

	response := make([]*rpc.BoardListItem, len(dat.Items))
	for i, v := range dat.Items {
		if v.Name == "" || v.FQBN == "" {
			return nil, errors.New(i18n.Tr("wrong format in server response"))
		}
		response[i] = &rpc.BoardListItem{Name: v.Name, Fqbn: v.FQBN}
	}
	return response, nil
}
