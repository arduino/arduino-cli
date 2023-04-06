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
	"fmt"
	"strconv"
	"strings"

	"github.com/arduino/arduino-cli/arduino/builder/preprocessor/ctags"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
)

var DebugPreprocessor bool

func PrototypesAdder(sketch *sketch.Sketch, source string, ctagsStdout []byte, lineOffset int) string {
	parser := &ctags.CTagsParser{}
	prototypes, firstFunctionLine := parser.Parse(ctagsStdout, sketch.MainFile)
	if firstFunctionLine == -1 {
		firstFunctionLine = 0
	}

	source = strings.Replace(source, "\r\n", "\n", -1)
	source = strings.Replace(source, "\r", "\n", -1)
	sourceRows := strings.Split(source, "\n")
	if isFirstFunctionOutsideOfSource(firstFunctionLine, sourceRows) {
		return ""
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
	return preprocessedSource
}

func composePrototypeSection(line int, prototypes []*ctags.Prototype) string {
	if len(prototypes) == 0 {
		return constants.EMPTY_STRING
	}

	str := joinPrototypes(prototypes)
	str += "\n#line "
	str += strconv.Itoa(line)
	str += " " + utils.QuoteCppString(prototypes[0].File)
	str += "\n"

	return str
}

func joinPrototypes(prototypes []*ctags.Prototype) string {
	prototypesSlice := []string{}
	for _, proto := range prototypes {
		if signatureContainsaDefaultArg(proto) {
			continue
		}
		prototypesSlice = append(prototypesSlice, "#line "+strconv.Itoa(proto.Line)+" "+utils.QuoteCppString(proto.File))
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
