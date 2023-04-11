// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package preprocessor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/arduino/arduino-cli/arduino/builder/cpp"
	"github.com/arduino/arduino-cli/arduino/builder/preprocessor/ctags"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

var tr = i18n.Tr

// DebugPreprocessor when set to true the CTags preprocessor will output debugging info to stdout
// this is useful for unit-testing to provide more infos
var DebugPreprocessor bool

// CTags performs a run of ctags on the given source file. Returns the ctags output and the stderr contents.
func CTags(sourceFile *paths.Path, targetFile *paths.Path, sketch *sketch.Sketch, lineOffset int, buildProperties *properties.Map) ([]byte, error) {
	ctagsOutput, ctagsStdErr, err := RunCTags(sourceFile, buildProperties)
	if err != nil {
		return ctagsStdErr, err
	}

	// func PrototypesAdder(sketch *sketch.Sketch, source string, ctagsStdout []byte, lineOffset int) string {
	parser := &ctags.CTagsParser{}
	prototypes, firstFunctionLine := parser.Parse(ctagsOutput, sketch.MainFile)
	if firstFunctionLine == -1 {
		firstFunctionLine = 0
	}

	var source string
	if sourceData, err := targetFile.ReadFile(); err != nil {
		return nil, err
	} else {
		source = string(sourceData)
	}
	source = strings.Replace(source, "\r\n", "\n", -1)
	source = strings.Replace(source, "\r", "\n", -1)
	sourceRows := strings.Split(source, "\n")
	if isFirstFunctionOutsideOfSource(firstFunctionLine, sourceRows) {
		return nil, nil
	}

	insertionLine := firstFunctionLine + lineOffset - 1
	firstFunctionChar := len(strings.Join(sourceRows[:insertionLine], "\n")) + 1
	prototypeSection := composePrototypeSection(firstFunctionLine, prototypes)
	preprocessedSource := source[:firstFunctionChar] + prototypeSection + source[firstFunctionChar:]

	if DebugPreprocessor {
		fmt.Println("#PREPROCESSED SOURCE")
		prototypesRows := strings.Split(prototypeSection, "\n")
		prototypesRows = prototypesRows[:len(prototypesRows)-1]
		for i := 0; i < len(sourceRows)+len(prototypesRows); i++ {
			if i < insertionLine {
				fmt.Printf("   |%s\n", sourceRows[i])
			} else if i < insertionLine+len(prototypesRows) {
				fmt.Printf("PRO|%s\n", prototypesRows[i-insertionLine])
			} else {
				fmt.Printf("   |%s\n", sourceRows[i-len(prototypesRows)])
			}
		}
		fmt.Println("#END OF PREPROCESSED SOURCE")
	}

	err = targetFile.WriteFile([]byte(preprocessedSource))
	return ctagsStdErr, err
}

func composePrototypeSection(line int, prototypes []*ctags.Prototype) string {
	if len(prototypes) == 0 {
		return ""
	}

	str := joinPrototypes(prototypes)
	str += "\n#line "
	str += strconv.Itoa(line)
	str += " " + cpp.QuoteString(prototypes[0].File)
	str += "\n"

	return str
}

func joinPrototypes(prototypes []*ctags.Prototype) string {
	prototypesSlice := []string{}
	for _, proto := range prototypes {
		if signatureContainsaDefaultArg(proto) {
			continue
		}
		prototypesSlice = append(prototypesSlice, "#line "+strconv.Itoa(proto.Line)+" "+cpp.QuoteString(proto.File))
		prototypeParts := []string{}
		if proto.Modifiers != "" {
			prototypeParts = append(prototypeParts, proto.Modifiers)
		}
		prototypeParts = append(prototypeParts, proto.Prototype)
		prototypesSlice = append(prototypesSlice, strings.Join(prototypeParts, " "))
	}
	return strings.Join(prototypesSlice, "\n")
}

func signatureContainsaDefaultArg(proto *ctags.Prototype) bool {
	return strings.Contains(proto.Prototype, "=")
}

func isFirstFunctionOutsideOfSource(firstFunctionLine int, sourceRows []string) bool {
	return firstFunctionLine > len(sourceRows)-1
}

// RunCTags performs a run of ctags on the given source file. Returns the ctags output and the stderr contents.
func RunCTags(sourceFile *paths.Path, buildProperties *properties.Map) ([]byte, []byte, error) {
	ctagsBuildProperties := properties.NewMap()
	ctagsBuildProperties.Set("tools.ctags.path", "{runtime.tools.ctags.path}")
	ctagsBuildProperties.Set("tools.ctags.cmd.path", "{path}/ctags")
	ctagsBuildProperties.Set("tools.ctags.pattern", `"{cmd.path}" -u --language-force=c++ -f - --c++-kinds=svpf --fields=KSTtzns --line-directives "{source_file}"`)
	ctagsBuildProperties.Merge(buildProperties)
	ctagsBuildProperties.Merge(ctagsBuildProperties.SubTree("tools").SubTree("ctags"))
	ctagsBuildProperties.SetPath("source_file", sourceFile)

	pattern := ctagsBuildProperties.Get("pattern")
	if pattern == "" {
		return nil, nil, errors.Errorf(tr("%s pattern is missing"), "ctags")
	}

	commandLine := ctagsBuildProperties.ExpandPropsInString(pattern)
	parts, err := properties.SplitQuotedString(commandLine, `"'`, false)
	if err != nil {
		return nil, nil, err
	}
	proc, err := executils.NewProcess(nil, parts...)
	if err != nil {
		return nil, nil, err
	}
	stdout, stderr, err := proc.RunAndCaptureOutput(context.Background())

	// Append ctags arguments to stderr
	args := fmt.Sprintln(strings.Join(parts, " "))
	stderr = append([]byte(args), stderr...)
	return stdout, stderr, err
}
