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
	"github.com/arduino/arduino-builder/constants"
	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
	"fmt"
	"strconv"
	"strings"
)

type PrototypesAdder struct{}

func (s *PrototypesAdder) Run(ctx *types.Context) error {
	debugOutput := ctx.DebugPreprocessor
	source := ctx.Source

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
	ctx.Source = source

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
