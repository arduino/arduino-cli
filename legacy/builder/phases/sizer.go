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

package phases

import (
	"regexp"
	"strconv"

	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

type Sizer struct {
	SketchError bool
}

func (s *Sizer) Run(ctx *types.Context) error {
	if ctx.OnlyUpdateCompilationDatabase {
		return nil
	}
	if s.SketchError {
		return nil
	}

	buildProperties := ctx.BuildProperties

	err := checkSize(ctx, buildProperties)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func checkSize(ctx *types.Context, buildProperties *properties.Map) error {
	logger := ctx.GetLogger()

	properties := buildProperties.Clone()
	properties.Set(constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS, properties.Get(constants.BUILD_PROPERTIES_COMPILER_WARNING_FLAGS+"."+ctx.WarningsLevel))

	maxTextSizeString := properties.Get(constants.PROPERTY_UPLOAD_MAX_SIZE)
	maxDataSizeString := properties.Get(constants.PROPERTY_UPLOAD_MAX_DATA_SIZE)

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

	textSize, dataSize, _, err := execSizeRecipe(ctx, properties)
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

	ctx.ExecutableSectionsSize = []types.ExecutableSectionSize{
		{
			Name:    "text",
			Size:    textSize,
			MaxSize: maxTextSize,
		},
	}
	if maxDataSize > 0 {
		ctx.ExecutableSectionsSize = append(ctx.ExecutableSectionsSize, types.ExecutableSectionSize{
			Name:    "data",
			Size:    dataSize,
			MaxSize: maxDataSize,
		})
	}

	if textSize > maxTextSize {
		logger.Println(constants.LOG_LEVEL_ERROR, constants.MSG_SIZER_TEXT_TOO_BIG)
		return errors.New("text section exceeds available space in board")
	}

	if maxDataSize > 0 && dataSize > maxDataSize {
		logger.Println(constants.LOG_LEVEL_ERROR, constants.MSG_SIZER_DATA_TOO_BIG)
		return errors.New("data section exceeds available space in board")
	}

	if properties.Get(constants.PROPERTY_WARN_DATA_PERCENT) != "" {
		warnDataPercentage, err := strconv.Atoi(properties.Get(constants.PROPERTY_WARN_DATA_PERCENT))
		if err != nil {
			return err
		}
		if maxDataSize > 0 && dataSize > maxDataSize*warnDataPercentage/100 {
			logger.Println(constants.LOG_LEVEL_WARN, constants.MSG_SIZER_LOW_MEMORY)
		}
	}

	return nil
}

func execSizeRecipe(ctx *types.Context, properties *properties.Map) (textSize int, dataSize int, eepromSize int, resErr error) {
	command, err := builder_utils.PrepareCommandForRecipe(properties, constants.RECIPE_SIZE_PATTERN, false)
	if err != nil {
		resErr = errors.New("Error while determining sketch size: " + err.Error())
		return
	}

	out, _, err := utils.ExecCommand(ctx, command, utils.Capture /* stdout */, utils.Show /* stderr */)
	if err != nil {
		resErr = errors.New("Error while determining sketch size: " + err.Error())
		return
	}

	// force multiline match prepending "(?m)" to the actual regexp
	// return an error if RECIPE_SIZE_REGEXP doesn't exist

	textSize, err = computeSize(properties.Get(constants.RECIPE_SIZE_REGEXP), out)
	if err != nil {
		resErr = errors.New("Invalid size regexp: " + err.Error())
		return
	}
	if textSize == -1 {
		resErr = errors.New("Missing size regexp")
		return
	}

	dataSize, err = computeSize(properties.Get(constants.RECIPE_SIZE_REGEXP_DATA), out)
	if err != nil {
		resErr = errors.New("Invalid data size regexp: " + err.Error())
		return
	}

	eepromSize, err = computeSize(properties.Get(constants.RECIPE_SIZE_REGEXP_EEPROM), out)
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
