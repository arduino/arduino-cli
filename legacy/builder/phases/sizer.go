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
	"fmt"
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
		ctx.Warn(tr("Couldn't determine program size"))
		return nil
	}

	ctx.Info(tr("Sketch uses %[1]s bytes (%[3]s%%) of program storage space. Maximum is %[2]s bytes.",
		strconv.Itoa(textSize),
		strconv.Itoa(maxTextSize),
		strconv.Itoa(textSize*100/maxTextSize)))
	if dataSize >= 0 {
		if maxDataSize > 0 {
			ctx.Info(tr("Global variables use %[1]s bytes (%[3]s%%) of dynamic memory, leaving %[4]s bytes for local variables. Maximum is %[2]s bytes.",
				strconv.Itoa(dataSize),
				strconv.Itoa(maxDataSize),
				strconv.Itoa(dataSize*100/maxDataSize),
				strconv.Itoa(maxDataSize-dataSize)))
		} else {
			ctx.Info(tr("Global variables use %[1]s bytes of dynamic memory.", strconv.Itoa(dataSize)))
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
		ctx.Warn(tr("Sketch too big; see %[1]s for tips on reducing it.", "https://support.arduino.cc/hc/en-us/articles/360013825179"))
		return errors.New(tr("text section exceeds available space in board"))
	}

	if maxDataSize > 0 && dataSize > maxDataSize {
		ctx.Warn(tr("Not enough memory; see %[1]s for tips on reducing your footprint.", "https://support.arduino.cc/hc/en-us/articles/360013825179"))
		return errors.New(tr("data section exceeds available space in board"))
	}

	if properties.Get(constants.PROPERTY_WARN_DATA_PERCENT) != "" {
		warnDataPercentage, err := strconv.Atoi(properties.Get(constants.PROPERTY_WARN_DATA_PERCENT))
		if err != nil {
			return err
		}
		if maxDataSize > 0 && dataSize > maxDataSize*warnDataPercentage/100 {
			ctx.Warn(tr("Low memory available, stability problems may occur."))
		}
	}

	return nil
}

func execSizeRecipe(ctx *types.Context, properties *properties.Map) (textSize int, dataSize int, eepromSize int, resErr error) {
	command, err := builder_utils.PrepareCommandForRecipe(properties, constants.RECIPE_SIZE_PATTERN, false, ctx.PackageManager.GetEnvVarsForSpawnedProcess())
	if err != nil {
		resErr = fmt.Errorf(tr("Error while determining sketch size: %s"), err)
		return
	}

	out, _, err := utils.ExecCommand(ctx, command, utils.Capture /* stdout */, utils.Show /* stderr */)
	if err != nil {
		resErr = fmt.Errorf(tr("Error while determining sketch size: %s"), err)
		return
	}

	// force multiline match prepending "(?m)" to the actual regexp
	// return an error if RECIPE_SIZE_REGEXP doesn't exist

	textSize, err = computeSize(properties.Get(constants.RECIPE_SIZE_REGEXP), out)
	if err != nil {
		resErr = fmt.Errorf(tr("Invalid size regexp: %s"), err)
		return
	}
	if textSize == -1 {
		resErr = errors.New(tr("Missing size regexp"))
		return
	}

	dataSize, err = computeSize(properties.Get(constants.RECIPE_SIZE_REGEXP_DATA), out)
	if err != nil {
		resErr = fmt.Errorf(tr("Invalid data size regexp: %s"), err)
		return
	}

	eepromSize, err = computeSize(properties.Get(constants.RECIPE_SIZE_REGEXP_EEPROM), out)
	if err != nil {
		resErr = fmt.Errorf(tr("Invalid eeprom size regexp: %s"), err)
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
