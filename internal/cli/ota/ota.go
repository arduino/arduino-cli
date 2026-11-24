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

package ota

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
)

// Build artifacts of the Arduino Zephyr core, served under fixed names so the
// OTA sketch on the board does not depend on the sketch name.
var otaFiles = []struct{ name, suffix string }{
	{"SKETCH.bin", ".elf-zsk.bin"},            // sketch only
	{"UPDATE.bin", ".elf-zsk.bin-bundle.bin"}, // loader + sketch
}

// NewCommand creates a new `ota` command
func NewCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var inputDir string
	otaCommand := &cobra.Command{
		Use:   "ota",
		Short: i18n.Tr("Make the sketch binaries available for OTA updates."),
		Long:  i18n.Tr("Make the sketch binaries of the last compilation available as SKETCH.bin (sketch only) and UPDATE.bin (loader + sketch), for OTA updates."),
		Example: "  " + os.Args[0] + " ota serve MySketch\n" +
			"  " + os.Args[0] + " ota export MySketch --output-dir /var/www/ota",
	}
	otaCommand.PersistentFlags().StringVarP(&inputDir, "input-dir", "", "", i18n.Tr("Directory containing binaries to use (default: the sketch build directory)."))

	var port int
	serveCommand := &cobra.Command{
		Use:   "serve [sketchPath]",
		Short: i18n.Tr("Serve SKETCH.bin and UPDATE.bin over HTTP."),
		Long:  i18n.Tr("Serve SKETCH.bin and UPDATE.bin over HTTP. The files are read on every request, so a new compilation is served without restarting."),
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			files := otaFilePaths(cmd.Context(), srv, args, inputDir)
			mux := http.NewServeMux()
			for _, f := range files {
				mux.HandleFunc("/"+f.name, func(w http.ResponseWriter, r *http.Request) {
					http.ServeFile(w, r, f.path.String())
				})
				feedback.Printf("/%s -> %s", f.name, f.path)
			}
			feedback.Print(i18n.Tr("Serving on port %d, press Ctrl+C to stop.", port))
			if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
				feedback.Fatal(err.Error(), feedback.ErrNetwork)
			}
		},
	}
	serveCommand.Flags().IntVarP(&port, "port", "p", 8000, i18n.Tr("HTTP server port."))

	var outputDir string
	exportCommand := &cobra.Command{
		Use:   "export [sketchPath]",
		Short: i18n.Tr("Copy SKETCH.bin and UPDATE.bin to a directory."),
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			out := paths.New(outputDir)
			if err := out.MkdirAll(); err != nil {
				feedback.Fatal(err.Error(), feedback.ErrGeneric)
			}
			for _, f := range otaFilePaths(cmd.Context(), srv, args, inputDir) {
				if err := f.path.CopyTo(out.Join(f.name)); err != nil {
					feedback.Fatal(err.Error(), feedback.ErrGeneric)
				}
				feedback.Printf("%s -> %s", f.path, out.Join(f.name))
			}
		},
	}
	exportCommand.Flags().StringVarP(&outputDir, "output-dir", "o", "", i18n.Tr("Destination directory."))
	_ = exportCommand.MarkFlagRequired("output-dir")

	otaCommand.AddCommand(serveCommand, exportCommand)
	return otaCommand
}

type otaFile struct {
	name string      // published name
	path *paths.Path // build artifact
}

// otaFilePaths returns the existing OTA artifacts of the sketch.
func otaFilePaths(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string, inputDir string) []otaFile {
	path := ""
	if len(args) > 0 {
		path = args[0]
	}
	sk, err := sketch.New(arguments.InitSketchPath(path))
	if err != nil {
		feedback.Fatal(err.Error(), feedback.ErrBadArgument)
	}

	buildDir := paths.New(inputDir)
	if buildDir == nil {
		// Same default build directory used by `compile` and `upload`.
		resp, err := srv.SettingsGetValue(ctx, &rpc.SettingsGetValueRequest{Key: "build_cache.path"})
		if err != nil {
			feedback.Fatal(err.Error(), feedback.ErrGeneric)
		}
		var cachePath string
		if err := json.Unmarshal([]byte(resp.GetEncodedValue()), &cachePath); err != nil {
			feedback.Fatal(err.Error(), feedback.ErrGeneric)
		}
		buildDir = paths.New(cachePath, "sketches", sk.Hash())
	}

	var files []otaFile
	for _, f := range otaFiles {
		if p := buildDir.Join(sk.MainFile.Base() + f.suffix); p.Exist() {
			files = append(files, otaFile{f.name, p})
		}
	}
	if len(files) == 0 {
		feedback.Fatal(i18n.Tr("No OTA binaries found in %s: compile the sketch first.", buildDir), feedback.ErrGeneric)
	}
	return files
}
