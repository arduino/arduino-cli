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

package lib

import (
	"context"
	"fmt"

	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
)

// LibraryInstall FIXMEDOC
func LibraryInstall(ctx context.Context, req *rpc.LibraryInstallReq,
	downloadCB commands.DownloadProgressCB, taskCB commands.TaskProgressCB) error {

	lm := commands.GetLibraryManager(req.GetInstance().GetId())

	libRelease, err := findLibraryIndexRelease(lm, req)
	if err != nil {
		return fmt.Errorf("looking for library: %s", err)
	}

	if err := downloadLibrary(lm, libRelease, downloadCB, taskCB); err != nil {
		return fmt.Errorf("downloading library: %s", err)
	}

	if err := installLibrary(lm, libRelease, taskCB); err != nil {
		return err
	}

	if _, err := commands.Rescan(req.GetInstance().GetId()); err != nil {
		return fmt.Errorf("rescanning libraries: %s", err)
	}
	return nil
}

func installLibrary(lm *librariesmanager.LibrariesManager, libRelease *librariesindex.Release, taskCB commands.TaskProgressCB) error {
	taskCB(&rpc.TaskProgress{Name: "Installing " + libRelease.String()})
	logrus.WithField("library", libRelease).Info("Installing library")
	libPath, libReplaced, err := lm.InstallPrerequisiteCheck(libRelease)
	if err == librariesmanager.ErrAlreadyInstalled {
		taskCB(&rpc.TaskProgress{Message: "Already installed " + libRelease.String(), Completed: true})
		return nil
	}

	if err != nil {
		return fmt.Errorf("checking lib install prerequisites: %s", err)
	}

	if libReplaced != nil {
		taskCB(&rpc.TaskProgress{Message: fmt.Sprintf("Replacing %s with %s", libReplaced, libRelease)})
	}

	if err := lm.Install(libRelease, libPath); err != nil {
		return err
	}

	taskCB(&rpc.TaskProgress{Message: "Installed " + libRelease.String(), Completed: true})
	return nil
}

//ZipLibraryInstall FIXMEDOC
func ZipLibraryInstall(ctx context.Context, req *rpc.ZipLibraryInstallReq) (*rpc.ZipLibraryInstallResp, error) {
	res := &rpc.ZipLibraryInstallResp{}
	lm := commands.GetLibraryManager(req.GetInstance().GetId())
	Path := req.GetPath()
	if err := installZipLibrary(lm, Path); err != nil {
		res.Status = "Error installing Zip Library"
		return res, err
	}
	res.Status = "Success! Installed Zip Library"
	return res, nil
}

func installZipLibrary(lm *librariesmanager.LibrariesManager, libPath string) error {
	if err := lm.InstallZipLib(libPath); err != nil {
		return err
	}
	return nil
}

//GitLibraryInstall FIXMEDOC
func GitLibraryInstall(ctx context.Context, req *rpc.GitLibraryInstallReq) (*rpc.GitLibraryInstallResp, error) {
	res := &rpc.GitLibraryInstallResp{}
	lm := commands.GetLibraryManager(req.GetInstance().GetId())
	URL := req.GetUrl()
	Name := req.GetName()
	if err := installGitLibrary(lm, Name, URL); err != nil {
		res.Status = "Error installing Git Library"
		return res, err
	}
	res.Status = "Success! Installed Git Library"
	return res, nil
}

func installGitLibrary(lm *librariesmanager.LibrariesManager, Name string, gitURL string) error {
	if err := lm.InstallGitLib(Name, gitURL); err != nil {
		return err
	}
	return nil
}
