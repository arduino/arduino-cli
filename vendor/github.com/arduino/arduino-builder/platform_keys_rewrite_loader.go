/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package builder

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
	"github.com/arduino/go-properties-map"
)

type PlatformKeysRewriteLoader struct{}

func (s *PlatformKeysRewriteLoader) Run(ctx *types.Context) error {
	folders := ctx.HardwareFolders

	platformKeysRewriteTxtPath, err := findPlatformKeysRewriteTxt(folders)
	if err != nil {
		return i18n.WrapError(err)
	}
	if platformKeysRewriteTxtPath == constants.EMPTY_STRING {
		return nil
	}

	platformKeysRewrite := types.PlatforKeysRewrite{}
	platformKeysRewrite.Rewrites = []types.PlatforKeyRewrite{}

	txt, err := properties.Load(platformKeysRewriteTxtPath)
	if err != nil {
		return i18n.WrapError(err)
	}
	keys := utils.KeysOfMapOfString(txt)
	sort.Strings(keys)

	for _, key := range keys {
		keyParts := strings.Split(key, ".")
		if keyParts[0] == constants.PLATFORM_REWRITE_OLD {
			index, err := strconv.Atoi(keyParts[1])
			if err != nil {
				return i18n.WrapError(err)
			}
			rewriteKey := strings.Join(keyParts[2:], ".")
			oldValue := txt[key]
			newValue := txt[constants.PLATFORM_REWRITE_NEW+"."+strings.Join(keyParts[1:], ".")]
			platformKeyRewrite := types.PlatforKeyRewrite{Key: rewriteKey, OldValue: oldValue, NewValue: newValue}
			platformKeysRewrite.Rewrites = growSliceOfRewrites(platformKeysRewrite.Rewrites, index)
			platformKeysRewrite.Rewrites[index] = platformKeyRewrite
		}
	}

	ctx.PlatformKeyRewrites = platformKeysRewrite

	return nil
}

func findPlatformKeysRewriteTxt(folders []string) (string, error) {
	for _, folder := range folders {
		txtPath := filepath.Join(folder, constants.FILE_PLATFORM_KEYS_REWRITE_TXT)
		_, err := os.Stat(txtPath)
		if err == nil {
			return txtPath, nil
		}
		if !os.IsNotExist(err) {
			return constants.EMPTY_STRING, i18n.WrapError(err)
		}
	}

	return constants.EMPTY_STRING, nil
}

func growSliceOfRewrites(originalSlice []types.PlatforKeyRewrite, maxIndex int) []types.PlatforKeyRewrite {
	if cap(originalSlice) > maxIndex {
		return originalSlice
	}
	newSlice := make([]types.PlatforKeyRewrite, maxIndex+1)
	copy(newSlice, originalSlice)
	return newSlice
}
