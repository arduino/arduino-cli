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

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

var tr = i18n.Tr

// PlatformDownload FIXMEDOC
func PlatformDownload(ctx context.Context, req *rpc.PlatformDownloadRequest, downloadCB rpc.DownloadProgressCB) (*rpc.PlatformDownloadResponse, error) {
	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return nil, err
	}
	defer release()

	version, err := ParseVersion(req.GetVersion())
	if err != nil {
		return nil, &cmderrors.InvalidVersionError{Cause: err}
	}

	ref := &packagemanager.PlatformReference{
		Package:              req.GetPlatformPackage(),
		PlatformArchitecture: req.GetArchitecture(),
		PlatformVersion:      version,
	}
	platform, tools, err := pme.FindPlatformReleaseDependencies(ref)
	if err != nil {
		return nil, &cmderrors.PlatformNotFoundError{Platform: ref.String(), Cause: err}
	}

	if err := pme.DownloadPlatformRelease(platform, nil, downloadCB); err != nil {
		return nil, err
	}

	for _, tool := range tools {
		if err := pme.DownloadToolRelease(tool, nil, downloadCB); err != nil {
			return nil, err
		}
	}

	return &rpc.PlatformDownloadResponse{}, nil
}
