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
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
)

type RewriteHardwareKeys struct{}

func (s *RewriteHardwareKeys) Run(ctx *types.Context) error {
	if ctx.PlatformKeyRewrites.Empty() {
		return nil
	}

	packages := ctx.Hardware
	platformKeysRewrite := ctx.PlatformKeyRewrites
	hardwareRewriteResults := ctx.HardwareRewriteResults

	for _, aPackage := range packages {
		for _, platform := range aPackage.Platforms {
			for _, platformRelease := range platform.Releases {
				if platformRelease.Properties.Get(constants.REWRITING) != constants.REWRITING_DISABLED {
					for _, rewrite := range platformKeysRewrite.Rewrites {
						if platformRelease.Properties.Get(rewrite.Key) == rewrite.OldValue {
							platformRelease.Properties.Set(rewrite.Key, rewrite.NewValue)
							appliedRewrites := rewritesAppliedToPlatform(platformRelease, hardwareRewriteResults)
							appliedRewrites = append(appliedRewrites, rewrite)
							hardwareRewriteResults[platformRelease] = appliedRewrites
						}
					}
				}
			}
		}
	}

	return nil
}

func rewritesAppliedToPlatform(platform *cores.PlatformRelease, hardwareRewriteResults map[*cores.PlatformRelease][]types.PlatforKeyRewrite) []types.PlatforKeyRewrite {
	if hardwareRewriteResults[platform] == nil {
		hardwareRewriteResults[platform] = []types.PlatforKeyRewrite{}
	}
	return hardwareRewriteResults[platform]
}
