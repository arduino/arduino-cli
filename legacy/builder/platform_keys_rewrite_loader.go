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
	"sort"
	"strconv"
	"strings"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

type PlatformKeysRewriteLoader struct{}

func (s *PlatformKeysRewriteLoader) Run(ctx *types.Context) error {
	folders := ctx.HardwareDirs

	platformKeysRewriteTxtPath, err := findPlatformKeysRewriteTxt(folders)
	if err != nil {
		return errors.WithStack(err)
	}
	if platformKeysRewriteTxtPath == nil {
		return nil
	}

	platformKeysRewrite := types.PlatforKeysRewrite{}
	platformKeysRewrite.Rewrites = []types.PlatforKeyRewrite{}

	txt, err := properties.LoadFromPath(platformKeysRewriteTxtPath)
	if err != nil {
		return errors.WithStack(err)
	}
	keys := txt.Keys()
	sort.Strings(keys)

	for _, key := range keys {
		keyParts := strings.Split(key, ".")
		if keyParts[0] == constants.PLATFORM_REWRITE_OLD {
			index, err := strconv.Atoi(keyParts[1])
			if err != nil {
				return errors.WithStack(err)
			}
			rewriteKey := strings.Join(keyParts[2:], ".")
			oldValue := txt.Get(key)
			newValue := txt.Get(constants.PLATFORM_REWRITE_NEW + "." + strings.Join(keyParts[1:], "."))
			platformKeyRewrite := types.PlatforKeyRewrite{Key: rewriteKey, OldValue: oldValue, NewValue: newValue}
			platformKeysRewrite.Rewrites = growSliceOfRewrites(platformKeysRewrite.Rewrites, index)
			platformKeysRewrite.Rewrites[index] = platformKeyRewrite
		}
	}

	ctx.PlatformKeyRewrites = platformKeysRewrite

	return nil
}

func findPlatformKeysRewriteTxt(folders paths.PathList) (*paths.Path, error) {
	for _, folder := range folders {
		txtPath := folder.Join(constants.FILE_PLATFORM_KEYS_REWRITE_TXT)
		if exist, err := txtPath.ExistCheck(); exist {
			return txtPath, nil
		} else if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return nil, nil
}

func growSliceOfRewrites(originalSlice []types.PlatforKeyRewrite, maxIndex int) []types.PlatforKeyRewrite {
	if cap(originalSlice) > maxIndex {
		return originalSlice
	}
	newSlice := make([]types.PlatforKeyRewrite, maxIndex+1)
	copy(newSlice, originalSlice)
	return newSlice
}
