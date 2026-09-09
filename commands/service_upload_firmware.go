// This file is part of arduino-cli.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package commands

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/resources"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/codeclysm/extract/v4"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
)

func (s *arduinoCoreServerImpl) ReadFirmwareFileDetails(ctx context.Context, req *rpc.ReadFirmwareFileDetailsRequest) (*rpc.ReadFirmwareFileDetailsResponse, error) {
	firmwareFile := req.GetFirmwareFile()
	fwDetails, err := readFirmwareFileDetails(ctx, paths.New(firmwareFile))
	if err != nil {
		return nil, err
	}
	return &rpc.ReadFirmwareFileDetailsResponse{FirmwareFileDetails: fwDetails}, nil
}

// UploadFirmwareFileToServerStreams return a server stream that forwards the output and error streams to the provided writers.
// It also returns a function that can be used to retrieve the result of the upload.
func UploadFirmwareFileToServerStreams(ctx context.Context, outStream io.Writer, errStream io.Writer) (rpc.ArduinoCoreService_UploadFirmwareFileServer, func() *rpc.UploadResult) {
	var result *rpc.UploadResult
	stream := streamResponseToCallback(ctx, func(resp *rpc.UploadFirmwareFileResponse) error {
		if errData := resp.GetErrStream(); len(errData) > 0 {
			_, err := errStream.Write(errData)
			return err
		}
		if outData := resp.GetOutStream(); len(outData) > 0 {
			_, err := outStream.Write(outData)
			return err
		}
		if res := resp.GetResult(); res != nil {
			result = res
		}
		return nil
	})
	return stream, func() *rpc.UploadResult {
		return result
	}
}

// UploadFirmwareFile performs the upload of a firmware file to a board.
func (s *arduinoCoreServerImpl) UploadFirmwareFile(req *rpc.UploadFirmwareFileRequest, stream rpc.ArduinoCoreService_UploadFirmwareFileServer) error {
	logrus.Tracef("Upload firmware file %s on %s started", req.GetFirmwareFile(), req.GetPort())
	syncSend := NewSynchronizedSend(stream.Send)

	ctx := stream.Context()
	outStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.UploadFirmwareFileResponse{
			Message: &rpc.UploadFirmwareFileResponse_OutStream{OutStream: data},
		})
	})
	defer outStream.Close()
	errStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.UploadFirmwareFileResponse{
			Message: &rpc.UploadFirmwareFileResponse_ErrStream{ErrStream: data},
		})
	})
	defer errStream.Close()
	taskCB := func(msg *rpc.TaskProgress) {
		syncSend.Send(&rpc.UploadFirmwareFileResponse{
			Message: &rpc.UploadFirmwareFileResponse_TaskProgress{TaskProgress: msg},
		})
	}
	downloadCB := func(progress *rpc.DownloadProgress) {
		syncSend.Send(&rpc.UploadFirmwareFileResponse{
			Message: &rpc.UploadFirmwareFileResponse_DownloadProgress{DownloadProgress: progress},
		})
	}

	// Open the firmware file and get its properties
	fwDetails, fwDir, err := loadFirmwareFile(stream.Context(), paths.New(req.GetFirmwareFile()))
	if err != nil {
		return err
	}
	defer fwDir.RemoveAll()

	pme, pmeRelease, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return err
	}
	defer pmeRelease()

	// Install required tools if needed
	installedTools := []*cores.ToolRelease{}
	for _, tool := range fwDetails.GetRequiredTools() {
		dep := cores.FromRpcToolDependencies(tool)
		installedTool := pme.FindToolDependency(dep)
		if installedTool == nil {
			toolRelease := cores.ToolReleaseFromRpcToolDependencies(tool)
			if err := pme.DownloadToolRelease(ctx, toolRelease, downloadCB); err != nil {
				return err
			}
			if err := pme.InstallTool(toolRelease, taskCB, false /* Skip post-install */, resources.IntegrityCheckFull); err != nil {
				return err
			}
			installedTool = pme.FindToolDependency(dep)
			if installedTool == nil {
				return fmt.Errorf("%s: %w", i18n.Tr("tool not found after installation"), err)
			}
		}
		installedTools = append(installedTools, installedTool)
	}

	// Load the runtime properties for the tools
	uploadProperties := properties.NewFromHashmap(fwDetails.GetUploadProperties())
	for _, installedTool := range installedTools {
		uploadProperties.Merge(installedTool.RuntimeProperties())
	}

	// Perform upload
	updatedPort, fwFileDetails, err := s.runProgramAction(
		stream.Context(),
		pme,
		req.GetPort(),
		req.GetVerbose(),
		req.GetVerify(),
		outStream,
		errStream,
		req.GetDryRun(),
		req.GetUserFields(),
		uploadProperties,
	)
	if err != nil {
		return err
	}
	return syncSend.Send(&rpc.UploadFirmwareFileResponse{
		Message: &rpc.UploadFirmwareFileResponse_Result{
			Result: &rpc.UploadResult{
				UpdatedUploadPort:   updatedPort,
				FirmwareFileDetails: fwFileDetails,
			},
		},
	})
}

// readFirmwareFileDetails only reads the firmware.json file from the firmware file
// and returns its contents as a FirmwareFileDetails object, without extracting the
// entire firmware file.
func readFirmwareFileDetails(ctx context.Context, firmwareFile *paths.Path) (*rpc.FirmwareFileDetails, error) {
	f, err := firmwareFile.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tmpDir, err := paths.MkTempDir("", "")
	if err != nil {
		return nil, err
	}
	defer tmpDir.RemoveAll()

	if err := extract.Archive(ctx, f, tmpDir.String(), func(s string) string {
		if strings.HasSuffix(s, "firmware.json") {
			return "firmware.json"
		}
		return ""
	}); err != nil {
		return nil, err
	}

	fwJson, err := tmpDir.Join("firmware.json").ReadFile()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.Tr("reading firmware details"), err)
	}
	details := &rpc.FirmwareFileDetails{}
	if err := protojson.Unmarshal(fwJson, details); err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.Tr("parsing firmware details"), err)
	}
	return details, nil
}

// loadFirmwareFile extracts the firmware file to a temporary directory and returns the
// FirmwareFileDetails object and the path to the temporary directory.
func loadFirmwareFile(ctx context.Context, firmwareFile *paths.Path) (_ *rpc.FirmwareFileDetails, _ *paths.Path, _err error) {
	// Create a temporary directory to extract the firmware file
	fwDir, err := paths.MkTempDir("", "")
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		// Remove the temporary directory if an error occurred
		if _err != nil {
			fwDir.RemoveAll()
		}
	}()

	// Unpack the firmware file into the temporary directory
	f, err := firmwareFile.Open()
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	if err := extract.Archive(ctx, f, fwDir.String(), func(s string) (r string) {
		// Cut the root folder in the archive, to avoid creating a nested directory structure in the temporary directory.
		_, file, found := strings.Cut(s, "/")
		if !found {
			return ""
		}
		return file
	}); err != nil {
		return nil, nil, err
	}

	// Now load the firmware file details
	fwFile, err := fwDir.Join("firmware.json").ReadFile()
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", i18n.Tr("reading firmware details"), err)
	}
	fwDetails := &rpc.FirmwareFileDetails{}
	if err := protojson.Unmarshal(fwFile, fwDetails); err != nil {
		return nil, nil, fmt.Errorf("%s: %w", i18n.Tr("parsing firmware details"), err)
	}

	return fwDetails, fwDir, nil
}
