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
		err := checkSizeAdvanced(ctx, buildProperties)
		return errors.WithStack(err)
	}

	err := checkSize(ctx, buildProperties)
	return errors.WithStack(err)
}

func checkSizeAdvanced(ctx *types.Context, properties *properties.Map) error {
	command, err := builder_utils.PrepareCommandForRecipe(properties, "recipe.advanced_size.pattern", false)
	if err != nil {
		return errors.New("Error while determining sketch size: " + err.Error())
	}

	out, _, err := utils.ExecCommand(ctx, command, utils.Capture /* stdout */, utils.Show /* stderr */)
	if err != nil {
		return errors.New("Error while determining sketch size: " + err.Error())
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
		return errors.New("Error while determining sketch size: " + err.Error())
	}

	ctx.ExecutableSectionsSize = resp.Sections
	logger := ctx.GetLogger()
	switch resp.Severity {
	case "error":
		logger.Println("error", "{0}", resp.Output)
		return errors.New(resp.ErrorMessage)
	case "warning":
		logger.Println("warn", "{0}", resp.Output)
	default: // "info"
		logger.Println("info", "{0}", resp.Output)
	}
	return nil
}

func checkSize(ctx *types.Context, buildProperties *properties.Map) error {
	logger := ctx.GetLogger()

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
		logger.Println("warn", "Couldn't determine program size")
		return nil
	}

	logger.Println("info",
		"Sketch uses {0} bytes ({2}%%) of program storage space. Maximum is {1} bytes.",
		strconv.Itoa(textSize), strconv.Itoa(maxTextSize), strconv.Itoa(textSize*100/maxTextSize))
	if dataSize >= 0 {
		if maxDataSize > 0 {
			logger.Println("info",
				"Global variables use {0} bytes ({2}%%) of dynamic memory, leaving {3} bytes for local variables. Maximum is {1} bytes.",
				strconv.Itoa(dataSize), strconv.Itoa(maxDataSize), strconv.Itoa(dataSize*100/maxDataSize), strconv.Itoa(maxDataSize-dataSize))
		} else {
			logger.Println("info",
				"Global variables use {0} bytes of dynamic memory.",
				strconv.Itoa(dataSize))
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
		logger.Println("error", "Sketch too big; see https://support.arduino.cc/hc/en-us/articles/360013825179 for tips on reducing it.")
		return errors.New("text section exceeds available space in board")
	}

	if maxDataSize > 0 && dataSize > maxDataSize {
		logger.Println("error", "Not enough memory; see https://support.arduino.cc/hc/en-us/articles/360013825179 for tips on reducing your footprint.")
		return errors.New("data section exceeds available space in board")
	}

	if w := properties.Get("build.warn_data_percentage"); w != "" {
		warnDataPercentage, err := strconv.Atoi(w)
		if err != nil {
			return err
		}
		if maxDataSize > 0 && dataSize > maxDataSize*warnDataPercentage/100 {
			logger.Println("warn", "Low memory available, stability problems may occur.")
		}
	}

	return nil
}

func execSizeRecipe(ctx *types.Context, properties *properties.Map) (textSize int, dataSize int, eepromSize int, resErr error) {
	command, err := builder_utils.PrepareCommandForRecipe(properties, "recipe.size.pattern", false)
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

	textSize, err = computeSize(properties.Get("recipe.size.regex"), out)
	if err != nil {
		resErr = errors.New("Invalid size regexp: " + err.Error())
		return
	}
	if textSize == -1 {
		resErr = errors.New("Missing size regexp")
		return
	}

	dataSize, err = computeSize(properties.Get("recipe.size.regex.data"), out)
	if err != nil {
		resErr = errors.New("Invalid data size regexp: " + err.Error())
		return
	}

	eepromSize, err = computeSize(properties.Get("recipe.size.regex.eeprom"), out)
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
