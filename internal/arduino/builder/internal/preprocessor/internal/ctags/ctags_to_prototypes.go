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

func (p *Parser) findLineWhereToInsertPrototypes() int {
	firstFunctionLine := p.firstFunctionAtLine()
	firstFunctionPointerAsArgument := p.firstFunctionPointerUsedAsArgument()
	if firstFunctionLine != -1 && firstFunctionPointerAsArgument != -1 {
		if firstFunctionLine < firstFunctionPointerAsArgument {
			return firstFunctionLine
		}
		return firstFunctionPointerAsArgument
	} else if firstFunctionLine != -1 {
		return firstFunctionLine
	} else if firstFunctionPointerAsArgument != -1 {
		return firstFunctionPointerAsArgument
	} else {
		return 0
	}
}

func (p *Parser) firstFunctionPointerUsedAsArgument() int {
	functionTags := p.collectFunctions()
	for _, tag := range p.tags {
		if functionNameUsedAsFunctionPointerIn(tag, functionTags) {
			return tag.Line
		}
	}
	return -1
}

func functionNameUsedAsFunctionPointerIn(tag *Tag, functionTags []*Tag) bool {
	// ctags escapes the backslashes and the slashes of the source line
	code := strings.NewReplacer(`\\`, `\`, `\/`, `/`).Replace(tag.Code)
	for _, functionTag := range functionTags {
		if tag.Line != functionTag.Line && codeUsesFunctionPointer(code, functionTag.FunctionName) {
			return true
		}
	}
	return false
}

func (p *Parser) collectFunctions() []*Tag {
	functionTags := []*Tag{}
	for _, tag := range p.tags {
		if tag.Kind == kindFunction && !tag.SkipMe {
			functionTags = append(functionTags, tag)
		}
	}
	return functionTags
}

func (p *Parser) firstFunctionAtLine() int {
	for _, tag := range p.tags {
		if !tagIsUnknown(tag) && isHandled(tag) && tag.Kind == kindFunction && tag.Filename == p.mainFile.String() {
			return tag.Line
		}
	}
	return -1
}

func (p *Parser) toPrototypes() []*Prototype {
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

// codeUsesFunctionPointer returns true if a line of code uses the function as a pointer, &name or (name), outside of its strings and comments.
func codeUsesFunctionPointer(code, name string) bool {
	pointer, argument := "&"+name, "("+name+")"
	for i := 0; i < len(code); i++ {
		switch {
		case strings.HasPrefix(code[i:], "//"):
			// The rest of the line is a comment
			return false
		case strings.HasPrefix(code[i:], "/*"):
			// Skip to the last char of the closing */
			end := strings.Index(code[i+2:], "*/")
			if end == -1 {
				return false
			}
			i += end + 3
		case code[i] == '"':
			// Skip to the closing quote, jumping over the escaped chars
			for i++; i < len(code) && code[i] != '"'; i++ {
				if code[i] == '\\' {
					i++
				}
			}
		case strings.HasPrefix(code[i:], pointer), strings.HasPrefix(code[i:], argument):
			return true
		}
	}
	return false
}
