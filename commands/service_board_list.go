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
	"errors"
	"time"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/arduino-cli/pkg/fqbn"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
)

// BoardList returns a list of boards found by the loaded discoveries.
// In case of errors partial results from discoveries that didn't fail
// are returned.
func (s *arduinoCoreServerImpl) BoardList(ctx context.Context, req *rpc.BoardListRequest) (*rpc.BoardListResponse, error) {
	var fqbnFilter *fqbn.FQBN
	if f := req.GetFqbn(); f != "" {
		var err error
		fqbnFilter, err = fqbn.Parse(f)
		if err != nil {
			return nil, &cmderrors.InvalidFQBNError{Cause: err}
		}
	}

	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return nil, err
	}
	defer release()
	dm := pme.DiscoveryManager()
	warnings := f.Map(dm.Start(), (error).Error)
	time.Sleep(time.Duration(req.GetTimeout()) * time.Millisecond)

	ports := []*rpc.DetectedPort{}
	for _, port := range dm.List() {
		boards, err := identify(pme, port.Properties, s.settings, req.GetSkipCloudApiForBoardDetection())
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

func hasMatchingBoard(b *rpc.DetectedPort, fqbnFilter *fqbn.FQBN) bool {
	for _, detectedBoard := range b.GetMatchingBoards() {
		detectedFqbn, err := fqbn.Parse(detectedBoard.GetFqbn())
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
	dm := pme.DiscoveryManager()

	watcher, err := dm.Watch()
	release()
	if err != nil {
		return err
	}

	go func() {
		for event := range watcher.Feed() {
			port := &rpc.DetectedPort{
				Port: rpc.DiscoveryPortToRPC(event.Port),
			}

			boardsError := ""
			if event.Type == "add" {
				boards, err := identify(pme, event.Port.Properties, s.settings, req.GetSkipCloudApiForBoardDetection())
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

	<-stream.Context().Done()
	logrus.Trace("closed watch")
	watcher.Close()
	return nil
}
