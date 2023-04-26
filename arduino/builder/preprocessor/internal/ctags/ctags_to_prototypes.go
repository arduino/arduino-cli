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

package ctags

import (
	"strconv"
	"strings"
)

// Prototype is a C++ prototype generated from ctags output
type Prototype struct {
	FunctionName string
	File         string
	Prototype    string
	Modifiers    string
	Line         int
}

func (proto *Prototype) String() string {
	return proto.Modifiers + " " + proto.Prototype + " @ " + strconv.Itoa(proto.Line)
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

func functionNameUsedAsFunctionPointerIn(tag *CTag, functionTags []*CTag) bool {
	for _, functionTag := range functionTags {
		if tag.Line != functionTag.Line && strings.Contains(tag.Code, "&"+functionTag.FunctionName) {
			return true
		}
		if tag.Line != functionTag.Line && strings.Contains(tag.Code, "("+functionTag.FunctionName+")") {
			return true
		}
	}
	return false
}

func (p *CTagsParser) collectFunctions() []*CTag {
	functionTags := []*CTag{}
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

func (p *CTagsParser) toPrototypes() []*Prototype {
	prototypes := []*Prototype{}
	for _, tag := range p.tags {
		if strings.TrimSpace(tag.Prototype) == "" {
			continue
		}
		if !tag.SkipMe {
			prototype := &Prototype{
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
