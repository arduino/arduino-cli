// This file is part of arduino-cli.
//
// Copyright 2026 ARDUINO SA (http://www.arduino.cc/)
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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	properties "github.com/arduino/go-properties-orderedmap"
)

// ReadMemoryToServerStreams forwards read-tool diagnostics and collects its result.
// Memory contents are saved to the requested server-side file, not these streams.
func ReadMemoryToServerStreams(ctx context.Context, stdout, stderr io.Writer) (rpc.ArduinoCoreService_ReadMemoryServer, func() *rpc.ReadMemoryResult) {
	var result *rpc.ReadMemoryResult
	stream := streamResponseToCallback(ctx, func(resp *rpc.ReadMemoryResponse) error {
		if data := resp.GetOutStream(); len(data) != 0 {
			_, err := stdout.Write(data)
			return err
		}
		if data := resp.GetErrStream(); len(data) != 0 {
			_, err := stderr.Write(data)
			return err
		}
		if resp.GetResult() != nil {
			result = resp.GetResult()
		}
		return nil
	})
	return stream, func() *rpc.ReadMemoryResult { return result }
}

// ReadMemory invokes only an explicitly declared read recipe. It never compiles
// a sketch, runs an erase/upload recipe, or attempts to unlock protected memory.
func (s *arduinoCoreServerImpl) ReadMemory(req *rpc.ReadMemoryRequest, stream rpc.ArduinoCoreService_ReadMemoryServer) error {
	options, err := newReadMemoryOptions(req)
	if err != nil {
		return err
	}
	defer options.cleanup()
	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return err
	}
	defer release()

	send := NewSynchronizedSend(stream.Send)
	stdout := feedStreamTo(func(data []byte) {
		send.Send(&rpc.ReadMemoryResponse{Message: &rpc.ReadMemoryResponse_OutStream{OutStream: data}})
	})
	defer stdout.Close()
	stderr := feedStreamTo(func(data []byte) {
		send.Send(&rpc.ReadMemoryResponse{Message: &rpc.ReadMemoryResponse_ErrStream{ErrStream: data}})
	})
	defer stderr.Close()
	port, err := s.runProgramAction(stream.Context(), pme, nil, "", "", req.GetFqbn(), req.GetPort(),
		req.GetProgrammer(), req.GetVerbose(), false, false, stdout, stderr, req.GetDryRun(), nil, nil, options)
	if err != nil {
		return err
	}
	if err := stream.Context().Err(); err != nil {
		return err
	}
	size, err := options.publish()
	if err != nil {
		return err
	}
	stdout.Close()
	stderr.Close()
	return send.Send(&rpc.ReadMemoryResponse{Message: &rpc.ReadMemoryResponse_Result{Result: &rpc.ReadMemoryResult{
		OutputFile: options.outputFile, Size: size, DryRun: options.dryRun, UpdatedPort: port,
	}}})
}

var memoryIdentifier = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type readMemoryOptions struct {
	memory, format, outputFile string
	temporaryDir, toolOutput   string
	dryRun, verbose            bool
}

func newReadMemoryOptions(req *rpc.ReadMemoryRequest) (*readMemoryOptions, error) {
	if !memoryIdentifier.MatchString(req.GetMemory()) || !memoryIdentifier.MatchString(req.GetFormat()) {
		return nil, errors.New(i18n.Tr("Memory and format must be lowercase identifiers containing letters, digits, or underscores"))
	}
	if strings.TrimSpace(req.GetOutputFile()) == "" || req.GetOutputFile() == "-" || strings.ContainsAny(req.GetOutputFile(), "\"\r\n") {
		return nil, errors.New(i18n.Tr("A new output file is required; quotes and line breaks are not supported in its path"))
	}
	outputFile, err := filepath.Abs(req.GetOutputFile())
	if err != nil {
		return nil, err
	}
	if _, err := os.Lstat(outputFile); err == nil {
		return nil, fmt.Errorf("%s: %s", i18n.Tr("Output file already exists"), outputFile)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	parent, err := os.Stat(filepath.Dir(outputFile))
	if err != nil {
		return nil, err
	}
	if !parent.IsDir() {
		return nil, errors.New(i18n.Tr("Output directory does not exist"))
	}
	return &readMemoryOptions{memory: req.GetMemory(), format: req.GetFormat(), outputFile: outputFile,
		dryRun: req.GetDryRun(), verbose: req.GetVerbose()}, nil
}

func (o *readMemoryOptions) prepareRecipe(action string, props *properties.Map) (string, error) {
	recipe := action + "." + o.memory + "." + o.format + ".pattern"
	if strings.TrimSpace(props.Get(recipe)) == "" {
		return "", errors.New(i18n.Tr("Readback is not supported for memory %s in format %s using this board/programmer", o.memory, o.format))
	}
	o.toolOutput = o.outputFile
	if !o.dryRun {
		var err error
		o.temporaryDir, err = os.MkdirTemp(filepath.Dir(o.outputFile), ".arduino-read-")
		if err != nil {
			return "", err
		}
		o.toolOutput = filepath.Join(o.temporaryDir, "memory."+o.format)
	}
	props.Set("read.output_file", filepath.ToSlash(o.toolOutput))
	verbosity := "quiet"
	if o.verbose {
		verbosity = "verbose"
	}
	props.Set("read.verbose", props.Get(action+".params."+verbosity))
	return recipe, nil
}

func (o *readMemoryOptions) cleanup() {
	if o.temporaryDir != "" {
		_ = os.RemoveAll(o.temporaryDir)
	}
}

func (o *readMemoryOptions) publish() (int64, error) {
	if o.dryRun {
		return 0, nil
	}
	info, err := os.Lstat(o.toolOutput)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", i18n.Tr("Read tool did not produce an output file"), err)
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return 0, errors.New(i18n.Tr("Read tool must produce a non-empty regular file"))
	}
	source, err := os.Open(o.toolOutput)
	if err != nil {
		return 0, err
	}
	defer source.Close()
	// Check again at publication time: another process may have created the
	// destination while the tool was running. Never truncate an existing backup.
	destination, err := os.OpenFile(o.outputFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return 0, err
	}
	size, copyErr := io.Copy(destination, source)
	syncErr := destination.Sync()
	closeErr := destination.Close()
	if err := errors.Join(copyErr, syncErr, closeErr); err != nil {
		_ = os.Remove(o.outputFile)
		return 0, err
	}
	return size, nil
}
