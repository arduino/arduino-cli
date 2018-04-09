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
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
	"bytes"
	"io/ioutil"
	"path/filepath"
)

type AdditionalSketchFilesCopier struct{}

func (s *AdditionalSketchFilesCopier) Run(ctx *types.Context) error {
	sketch := ctx.Sketch
	sketchBuildPath := ctx.SketchBuildPath

	err := utils.EnsureFolderExists(sketchBuildPath)
	if err != nil {
		return i18n.WrapError(err)
	}

	sketchBasePath := filepath.Dir(sketch.MainFile.Name)

	for _, file := range sketch.AdditionalFiles {
		relativePath, err := filepath.Rel(sketchBasePath, file.Name)
		if err != nil {
			return i18n.WrapError(err)
		}

		targetFilePath := filepath.Join(sketchBuildPath, relativePath)
		err = utils.EnsureFolderExists(filepath.Dir(targetFilePath))
		if err != nil {
			return i18n.WrapError(err)
		}

		bytes, err := ioutil.ReadFile(file.Name)
		if err != nil {
			return i18n.WrapError(err)
		}

		if targetFileChanged(bytes, targetFilePath) {
			err := utils.WriteFileBytes(targetFilePath, bytes)
			if err != nil {
				return i18n.WrapError(err)
			}
		}
	}

	return nil
}

func targetFileChanged(currentBytes []byte, targetFilePath string) bool {
	oldBytes, err := ioutil.ReadFile(targetFilePath)
	if err != nil {
		return true
	}

	return bytes.Compare(currentBytes, oldBytes) != 0
}
