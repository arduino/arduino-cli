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

package builder

import (
	"os"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
)

type WarnAboutPlatformRewrites struct{}

func (s *WarnAboutPlatformRewrites) Run(ctx *types.Context) error {
	if ctx.DebugLevel < 0 {
		return nil
	}

	logger := ctx.GetLogger()
	hardwareRewriteResults := ctx.HardwareRewriteResults
	targetPlatform := ctx.TargetPlatform
	actualPlatform := ctx.ActualPlatform

	platforms := []*cores.PlatformRelease{targetPlatform}
	if actualPlatform != targetPlatform {
		platforms = append(platforms, actualPlatform)
	}

	for _, platform := range platforms {
		if hardwareRewriteResults[platform] != nil {
			for _, rewrite := range hardwareRewriteResults[platform] {
				logger.Fprintln(os.Stdout, constants.LOG_LEVEL_WARN,
					constants.MSG_WARNING_PLATFORM_OLD_VALUES,
					platform.Properties.Get(constants.PLATFORM_NAME),
					rewrite.Key+"="+rewrite.OldValue,
					rewrite.Key+"="+rewrite.NewValue)
			}
		}
	}

	return nil
}
