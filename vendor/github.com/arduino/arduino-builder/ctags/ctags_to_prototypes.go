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

package ctags

import (
	"strings"

	"github.com/arduino/arduino-builder/types"
)

func (p *CTagsParser) GeneratePrototypes() ([]*types.Prototype, int) {
	return p.toPrototypes(), p.findLineWhereToInsertPrototypes()
}

func (p *CTagsParser) findLineWhereToInsertPrototypes() int {
	firstFunctionLine := p.firstFunctionAtLine()
	firstFunctionPointerAsArgument := p.firstFunctionPointerUsedAsArgument()
	if firstFunctionLine != -1 && firstFunctionPointerAsArgument != -1 {
		if firstFunctionLine < firstFunctionPointerAsArgument {
			return firstFunctionLine
		} else {
			return firstFunctionPointerAsArgument
		}
	} else if firstFunctionLine != -1 {
		return firstFunctionLine
	} else if firstFunctionPointerAsArgument != -1 {
		return firstFunctionPointerAsArgument
	} else {
		return 0
	}
}

func (p *CTagsParser) firstFunctionPointerUsedAsArgument() int {
	functionTags := p.collectFunctions()
	for _, tag := range p.tags {
		if functionNameUsedAsFunctionPointerIn(tag, functionTags) {
			return tag.Line
		}
	}
	return -1
}

func functionNameUsedAsFunctionPointerIn(tag *types.CTag, functionTags []*types.CTag) bool {
	for _, functionTag := range functionTags {
		if tag.Line != functionTag.Line && strings.Index(tag.Code, "&"+functionTag.FunctionName) != -1 {
			return true
		}
		if tag.Line != functionTag.Line && strings.Index(tag.Code, functionTag.FunctionName) != -1 &&
			(functionTag.Signature == "(void)" || functionTag.Signature == "()") {
			return true
		}
	}
	return false
}

func (p *CTagsParser) collectFunctions() []*types.CTag {
	functionTags := []*types.CTag{}
	for _, tag := range p.tags {
		if tag.Kind == KIND_FUNCTION && !tag.SkipMe {
			functionTags = append(functionTags, tag)
		}
	}
	return functionTags
}

func (p *CTagsParser) firstFunctionAtLine() int {
	for _, tag := range p.tags {
		if !tagIsUnknown(tag) && isHandled(tag) && tag.Kind == KIND_FUNCTION && tag.Filename == p.mainFile.String() {
			return tag.Line
		}
	}
	return -1
}

func (p *CTagsParser) toPrototypes() []*types.Prototype {
	prototypes := []*types.Prototype{}
	for _, tag := range p.tags {
		if strings.TrimSpace(tag.Prototype) == "" {
			continue
		}
		if !tag.SkipMe {
			prototype := &types.Prototype{
				FunctionName: tag.FunctionName,
				File:         tag.Filename,
				Prototype:    tag.Prototype,
				Modifiers:    tag.PrototypeModifiers,
				Line:         tag.Line,
				//Fields:       tag,
			}
			prototypes = append(prototypes, prototype)
		}
	}
	return prototypes
}
