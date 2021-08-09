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

package upload

import (
	"context"
	"io"

	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
)

// BurnBootloader FIXMEDOC
func BurnBootloader(ctx context.Context, req *rpc.BurnBootloaderRequest, outStream io.Writer, errStream io.Writer) (*rpc.BurnBootloaderResponse, error) {
	logrus.
		WithField("fqbn", req.GetFqbn()).
		WithField("port", req.GetPort()).
		WithField("programmer", req.GetProgrammer()).
		Trace("BurnBootloader started", req.GetFqbn())

	pm := commands.GetPackageManager(req.GetInstance().GetId())

	err := runProgramAction(
		pm,
		nil, // sketch
		"",  // importFile
		"",  // importDir
		req.GetFqbn(),
		req.GetPort(),
		req.GetProgrammer(),
		req.GetVerbose(),
		req.GetVerify(),
		true, // burnBootloader
		outStream,
		errStream,
		req.GetDryRun(),
		map[string]string{}, // User fields
	)
	if err != nil {
		return nil, err
	}
	return &rpc.BurnBootloaderResponse{}, nil
}
