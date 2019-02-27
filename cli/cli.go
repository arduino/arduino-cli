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

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/arduino/sketches"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/configs"
	"github.com/arduino/arduino-cli/rpc"
	paths "github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// Error codes to be used for os.Exit().
const (
	_          = iota // 0 is not a valid exit error code
	ErrGeneric        // 1 is the reserved "catchall" code in Unix
	_                 // 2 is reserved in Unix
	ErrNoConfigFile
	ErrBadCall
	ErrNetwork
	// ErrCoreConfig represents an error in the cli core config, for example some basic
	// files shipped with the installation are missing, or cannot create or get basic
	// directories vital for the CLI to work.
	ErrCoreConfig
	ErrBadArgument
)

// Version is the current CLI version
var Version = "0.3.4-alpha.preview"

// ErrLogrus represents the logrus instance, which has the role to
// log all non info messages.
var ErrLogrus = logrus.New()

// GlobalFlags represents flags available in all the program.
var GlobalFlags struct {
	Debug      bool // If true, dump debug output to stderr.
	OutputJSON bool // true output in JSON, false output as Text
}

// OutputJSONOrElse outputs the JSON encoding of v if the JSON output format has been
// selected by the user and returns false. Otherwise no output is produced and the
// function returns true.
func OutputJSONOrElse(v interface{}) bool {
	if !GlobalFlags.OutputJSON {
		return true
	}
	d, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		formatter.PrintError(err, "Error during JSON encoding of the output")
		os.Exit(ErrGeneric)
	}
	fmt.Print(string(d))
	return false
}

// AppName is the command line name of the Arduino CLI executable
var AppName = filepath.Base(os.Args[0])

var Config *configs.Configuration

// InitPackageAndLibraryManagerWithoutBundles initializes the PackageManager
// and the LibraryManager but ignores bundles and user installed cores
func InitPackageAndLibraryManagerWithoutBundles() (*packagemanager.PackageManager, *librariesmanager.LibrariesManager) {
	logrus.Info("Package manager will scan only managed hardware folder")

	fakeResult := false
	Config.IDEBundledCheckResult = &fakeResult
	Config.SketchbookDir = nil
	return InitPackageAndLibraryManager()
}

func packageManagerInitReq() *rpc.InitReq {
	urls := []string{}
	for _, URL := range Config.BoardManagerAdditionalUrls {
		urls = append(urls, URL.String())
	}

	conf := &rpc.Configuration{}
	conf.DataDir = Config.DataDir.String()
	conf.DownloadsDir = Config.DownloadsDir().String()
	conf.BoardManagerAdditionalUrls = urls
	if Config.SketchbookDir != nil {
		conf.SketchbookDir = Config.SketchbookDir.String()
	}

	return &rpc.InitReq{Configuration: conf}
}

// CreateInstance creates and return an instance of the Arduino Core engine
func CreateInstance() *rpc.Instance {
	logrus.Info("Initializing package manager")
	resp, err := commands.Init(context.Background(), packageManagerInitReq())
	if err != nil {
		formatter.PrintError(err, "Error initializing package manager")
		os.Exit(rpc.ErrGeneric)
	}
	return resp.GetInstance()
}

// InitPackageAndLibraryManager initializes the PackageManager and the LibaryManager
// TODO: for the daemon mode, this might be called at startup, but for now only commands needing the PM will call it
func InitPackageAndLibraryManager() (*packagemanager.PackageManager, *librariesmanager.LibrariesManager) {
	logrus.Info("Initializing package manager")
	resp, err := commands.Init(context.Background(), packageManagerInitReq())
	if err != nil {
		formatter.PrintError(err, "Error initializing package manager")
		os.Exit(rpc.ErrGeneric)
	}
	return commands.GetPackageManager(resp), commands.GetLibraryManager(resp)
}

// InitLibraryManager initializes the LibraryManager. If pm is nil, the library manager will not handle core-libraries.
// TODO: for the daemon mode, this might be called at startup, but for now only commands needing the PM will call it
func InitLibraryManager(cfg *configs.Configuration) *librariesmanager.LibrariesManager {
	req := packageManagerInitReq()
	req.LibraryManagerOnly = true

	logrus.Info("Initializing library manager")
	resp, err := commands.Init(context.Background(), req)
	if err != nil {
		formatter.PrintError(err, "Error initializing library manager")
		os.Exit(rpc.ErrGeneric)
	}
	return commands.GetLibraryManager(resp)
}

// UpdateLibrariesIndex updates the library_index.json
func UpdateLibrariesIndex(lm *librariesmanager.LibrariesManager) {
	logrus.Info("Updating libraries index")
	d, err := lm.UpdateIndex()
	if err != nil {
		formatter.PrintError(err, "Error downloading librarires index")
		os.Exit(ErrNetwork)
	}
	formatter.DownloadProgressBar(d, "Updating index: library_index.json")
	if d.Error() != nil {
		formatter.PrintError(d.Error(), "Error downloading librarires index")
		os.Exit(ErrNetwork)
	}
}

func InitSketch(sketchPath *paths.Path) (*sketches.Sketch, error) {
	if sketchPath != nil {
		return sketches.NewSketchFromPath(sketchPath)
	}

	wd, err := paths.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting current directory: %s", err)
	}
	logrus.Infof("Reading sketch from dir: %s", wd)
	return sketches.NewSketchFromPath(wd)
}
