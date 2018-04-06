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
 * Copyright 2016 Arduino LLC (http://www.arduino.cc/)
 */

package phases

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/arduino/arduino-builder/builder_utils"
	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/i18n"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/go-properties-map"
)

type Sizer struct {
	SketchError bool
}

func (s *Sizer) Run(ctx *types.Context) error {

	if s.SketchError {
		return nil
	}

	buildProperties := ctx.BuildProperties
	verbose := ctx.Verbose
	warningsLevel := ctx.WarningsLevel
	logger := ctx.GetLogger()

	err := checkSize(buildProperties, verbose, warningsLevel, logger)
	if err != nil {
		return i18n.WrapError(err)
	}

	return nil
}

func checkSize(buildProperties properties.Map, verbose bool, warningsLevel string, logger i18n.Logger) error {

	properties := buildProperties.Clone()
	properties[constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS] = properties[constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS+"."+warningsLevel]

	maxTextSizeString := properties[constants.PROPERTY_UPLOAD_MAX_SIZE]
	maxDataSizeString := properties[constants.PROPERTY_UPLOAD_MAX_DATA_SIZE]

	if maxTextSizeString == "" {
		return nil
	}

	maxTextSize, err := strconv.Atoi(maxTextSizeString)
	if err != nil {
		return err
	}

	maxDataSize := -1
	if maxDataSizeString != "" {
		maxDataSize, err = strconv.Atoi(maxDataSizeString)
		if err != nil {
			return err
		}
	}

	textSize, dataSize, _, err := execSizeReceipe(properties, logger)
	if err != nil {
		logger.Println(constants.LOG_LEVEL_WARN, constants.MSG_SIZER_ERROR_NO_RULE)
		return nil
	}

	logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_SIZER_TEXT_FULL, strconv.Itoa(textSize), strconv.Itoa(maxTextSize), strconv.Itoa(textSize*100/maxTextSize))
	if dataSize >= 0 {
		if maxDataSize > 0 {
			logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_SIZER_DATA_FULL, strconv.Itoa(dataSize), strconv.Itoa(maxDataSize), strconv.Itoa(dataSize*100/maxDataSize), strconv.Itoa(maxDataSize-dataSize))
		} else {
			logger.Println(constants.LOG_LEVEL_INFO, constants.MSG_SIZER_DATA, strconv.Itoa(dataSize))
		}
	}

	if textSize > maxTextSize {
		logger.Println(constants.LOG_LEVEL_ERROR, constants.MSG_SIZER_TEXT_TOO_BIG)
		return errors.New("")
	}

	if maxDataSize > 0 && dataSize > maxDataSize {
		logger.Println(constants.LOG_LEVEL_ERROR, constants.MSG_SIZER_DATA_TOO_BIG)
		return errors.New("")
	}

	if properties[constants.PROPERTY_WARN_DATA_PERCENT] != "" {
		warnDataPercentage, err := strconv.Atoi(properties[constants.PROPERTY_WARN_DATA_PERCENT])
		if err != nil {
			return err
		}
		if maxDataSize > 0 && dataSize > maxDataSize*warnDataPercentage/100 {
			logger.Println(constants.LOG_LEVEL_WARN, constants.MSG_SIZER_LOW_MEMORY)
		}
	}

	return nil
}

func execSizeReceipe(properties properties.Map, logger i18n.Logger) (textSize int, dataSize int, eepromSize int, resErr error) {
	out, err := builder_utils.ExecRecipe(properties, constants.RECIPE_SIZE_PATTERN, false, false, false, logger)
	if err != nil {
		resErr = errors.New("Error while determining sketch size: " + err.Error())
		return
	}

	// force multiline match prepending "(?m)" to the actual regexp
	// return an error if RECIPE_SIZE_REGEXP doesn't exist

	textSize, err = computeSize(properties[constants.RECIPE_SIZE_REGEXP], out)
	if err != nil {
		resErr = errors.New("Invalid size regexp: " + err.Error())
		return
	}
	if textSize == -1 {
		resErr = errors.New("Missing size regexp")
		return
	}

	dataSize, err = computeSize(properties[constants.RECIPE_SIZE_REGEXP_DATA], out)
	if err != nil {
		resErr = errors.New("Invalid data size regexp: " + err.Error())
		return
	}

	eepromSize, err = computeSize(properties[constants.RECIPE_SIZE_REGEXP_EEPROM], out)
	if err != nil {
		resErr = errors.New("Invalid eeprom size regexp: " + err.Error())
		return
	}

	return
}

func computeSize(re string, output []byte) (int, error) {
	if re == "" {
		return -1, nil
	}
	r, err := regexp.Compile("(?m)" + re)
	if err != nil {
		return -1, err
	}
	result := r.FindAllSubmatch(output, -1)
	size := 0
	for _, b := range result {
		for _, c := range b {
			if res, err := strconv.Atoi(string(c)); err == nil {
				size += res
			}
		}
	}
	return size, nil
}
