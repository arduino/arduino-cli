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

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
)

type PrototypesAdder struct{}

func (s *PrototypesAdder) Run(ctx *types.Context) error {
	debugOutput := ctx.DebugPreprocessor

	source := ctx.SketchSourceMerged
	source = strings.Replace(source, "\r\n", "\n", -1)
	source = strings.Replace(source, "\r", "\n", -1)

	sourceRows := strings.Split(source, "\n")

	firstFunctionLine := ctx.PrototypesLineWhereToInsert
	if isFirstFunctionOutsideOfSource(firstFunctionLine, sourceRows) {
		return nil
	}

	insertionLine := firstFunctionLine + ctx.LineOffset - 1
	firstFunctionChar := len(strings.Join(sourceRows[:insertionLine], "\n")) + 1
	prototypeSection := composePrototypeSection(firstFunctionLine, ctx.Prototypes)
	ctx.PrototypesSection = prototypeSection
	source = source[:firstFunctionChar] + prototypeSection + source[firstFunctionChar:]

	if debugOutput {
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
	ctx.SketchSourceAfterArduinoPreprocessing = source

	return nil
}

func composePrototypeSection(line int, prototypes []*types.Prototype) string {
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

func joinPrototypes(prototypes []*types.Prototype) string {
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

func signatureContainsaDefaultArg(proto *types.Prototype) bool {
	return strings.Contains(proto.Prototype, "=")
}

func isFirstFunctionOutsideOfSource(firstFunctionLine int, sourceRows []string) bool {
	return firstFunctionLine > len(sourceRows)-1
}
