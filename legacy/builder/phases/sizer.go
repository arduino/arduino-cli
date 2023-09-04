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
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
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

	if buildProperties.ContainsKey("recipe.advanced_size.pattern") {
		return checkSizeAdvanced(ctx, buildProperties)
	}

	return checkSize(ctx, buildProperties)
}

func checkSizeAdvanced(ctx *types.Context, properties *properties.Map) error {
	command, err := builder_utils.PrepareCommandForRecipe(properties, "recipe.advanced_size.pattern", false)
	if err != nil {
		return errors.New(tr("Error while determining sketch size: %s", err))
	}

	out, _, err := utils.ExecCommand(ctx, command, utils.Capture /* stdout */, utils.Show /* stderr */)
	if err != nil {
		return errors.New(tr("Error while determining sketch size: %s", err))
	}

	type AdvancedSizerResponse struct {
		// Output are the messages displayed in console to the user
		Output string `json:"output"`
		// Severity may be one of "info", "warning" or "error". Warnings and errors will
		// likely be printed in red. Errors will stop build/upload.
		Severity string `json:"severity"`
		// Sections are the sections sizes for machine readable use
		Sections []types.ExecutableSectionSize `json:"sections"`
		// ErrorMessage is a one line error message like:
		// "text section exceeds available space in board"
		// it must be set when Severity is "error"
		ErrorMessage string `json:"error"`
	}

	var resp AdvancedSizerResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return errors.New(tr("Error while determining sketch size: %s", err))
	}

	ctx.ExecutableSectionsSize = resp.Sections
	switch resp.Severity {
	case "error":
		ctx.Warn(resp.Output)
		return errors.New(resp.ErrorMessage)
	case "warning":
		ctx.Warn(resp.Output)
	case "info":
		ctx.Info(resp.Output)
	default:
		return fmt.Errorf("invalid '%s' severity from sketch sizer: it must be 'error', 'warning' or 'info'", resp.Severity)
	}
	return nil
}

func checkSize(ctx *types.Context, buildProperties *properties.Map) error {
	properties := buildProperties.Clone()
	properties.Set("compiler.warning_flags", properties.Get("compiler.warning_flags."+ctx.WarningsLevel))

	maxTextSizeString := properties.Get("upload.maximum_size")
	maxDataSizeString := properties.Get("upload.maximum_data_size")

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

	if w := properties.Get("build.warn_data_percentage"); w != "" {
		warnDataPercentage, err := strconv.Atoi(w)
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
	command, err := builder_utils.PrepareCommandForRecipe(properties, "recipe.size.pattern", false)
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

	textSize, err = computeSize(properties.Get("recipe.size.regex"), out)
	if err != nil {
		resErr = fmt.Errorf(tr("Invalid size regexp: %s"), err)
		return
	}
	if textSize == -1 {
		resErr = errors.New(tr("Missing size regexp"))
		return
	}

	dataSize, err = computeSize(properties.Get("recipe.size.regex.data"), out)
	if err != nil {
		resErr = fmt.Errorf(tr("Invalid data size regexp: %s"), err)
		return
	}

	eepromSize, err = computeSize(properties.Get("recipe.size.regex.eeprom"), out)
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
