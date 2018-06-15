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
	"strconv"
	"strings"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/arduino-builder/types"
)

const KIND_PROTOTYPE = "prototype"
const KIND_FUNCTION = "function"

//const KIND_PROTOTYPE_MODIFIERS = "prototype_modifiers"

const TEMPLATE = "template"
const STATIC = "static"
const EXTERN = "extern \"C\""

var KNOWN_TAG_KINDS = map[string]bool{
	"prototype": true,
	"function":  true,
}

type CTagsParser struct {
	tags     []*types.CTag
	mainFile *paths.Path
}

func (p *CTagsParser) Parse(ctagsOutput string, mainFile *paths.Path) []*types.CTag {
	rows := strings.Split(ctagsOutput, "\n")
	rows = removeEmpty(rows)

	p.mainFile = mainFile

	for _, row := range rows {
		p.tags = append(p.tags, parseTag(row))
	}

	p.skipTagsWhere(tagIsUnknown)
	p.skipTagsWhere(tagIsUnhandled)
	p.addPrototypes()
	p.removeDefinedProtypes()
	p.skipDuplicates()
	p.skipTagsWhere(p.prototypeAndCodeDontMatch)

	return p.tags
}

func (p *CTagsParser) addPrototypes() {
	for _, tag := range p.tags {
		if !tag.SkipMe {
			addPrototype(tag)
		}
	}
}

func addPrototype(tag *types.CTag) {
	if strings.Index(tag.Prototype, TEMPLATE) == 0 {
		if strings.Index(tag.Code, TEMPLATE) == 0 {
			code := tag.Code
			if strings.Contains(code, "{") {
				code = code[:strings.Index(code, "{")]
			} else {
				code = code[:strings.LastIndex(code, ")")+1]
			}
			tag.Prototype = code + ";"
		} else {
			//tag.Code is 99% multiline, recreate it
			code := findTemplateMultiline(tag)
			tag.Prototype = code + ";"
		}
		return
	}

	tag.PrototypeModifiers = ""
	if strings.Index(tag.Code, STATIC+" ") != -1 {
		tag.PrototypeModifiers = tag.PrototypeModifiers + " " + STATIC
	}

	// Extern "C" modifier is now added in FixCLinkageTagsDeclarations

	tag.PrototypeModifiers = strings.TrimSpace(tag.PrototypeModifiers)
}

func (p *CTagsParser) removeDefinedProtypes() {
	definedPrototypes := make(map[string]bool)
	for _, tag := range p.tags {
		if tag.Kind == KIND_PROTOTYPE {
			definedPrototypes[tag.Prototype] = true
		}
	}

	for _, tag := range p.tags {
		if definedPrototypes[tag.Prototype] {
			//if ctx.DebugLevel >= 10 {
			//	ctx.GetLogger().Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, constants.MSG_SKIPPING_TAG_ALREADY_DEFINED, tag.FunctionName)
			//}
			tag.SkipMe = true
		}
	}
}

func (p *CTagsParser) skipDuplicates() {
	definedPrototypes := make(map[string]bool)

	for _, tag := range p.tags {
		if !definedPrototypes[tag.Prototype] && tag.SkipMe == false {
			definedPrototypes[tag.Prototype] = true
		} else {
			tag.SkipMe = true
		}
	}
}

type skipFuncType func(tag *types.CTag) bool

func (p *CTagsParser) skipTagsWhere(skipFunc skipFuncType) {
	for _, tag := range p.tags {
		if !tag.SkipMe {
			skip := skipFunc(tag)
			//if skip && p.debugLevel >= 10 {
			//	ctx.GetLogger().Fprintln(os.Stdout, constants.LOG_LEVEL_DEBUG, constants.MSG_SKIPPING_TAG_WITH_REASON, tag.FunctionName, runtime.FuncForPC(reflect.ValueOf(skipFunc).Pointer()).Name())
			//}
			tag.SkipMe = skip
		}
	}
}

func removeTralingSemicolon(s string) string {
	return s[0 : len(s)-1]
}

func removeSpacesAndTabs(s string) string {
	s = strings.Replace(s, " ", "", -1)
	s = strings.Replace(s, "\t", "", -1)
	return s
}

func tagIsUnhandled(tag *types.CTag) bool {
	return !isHandled(tag)
}

func isHandled(tag *types.CTag) bool {
	if tag.Class != "" {
		return false
	}
	if tag.Struct != "" {
		return false
	}
	if tag.Namespace != "" {
		return false
	}
	return true
}

func tagIsUnknown(tag *types.CTag) bool {
	return !KNOWN_TAG_KINDS[tag.Kind]
}

func parseTag(row string) *types.CTag {
	tag := &types.CTag{}
	parts := strings.Split(row, "\t")

	tag.FunctionName = parts[0]
	// This unescapes any backslashes in the filename. These
	// filenames that ctags outputs originate from the line markers
	// in the source, as generated by gcc. gcc escapes both
	// backslashes and double quotes, but ctags ignores any escaping
	// and just cuts off the filename at the first double quote it
	// sees. This means any backslashes are still escaped, and need
	// to be unescape, and any quotes will just break the build.
	tag.Filename = strings.Replace(parts[1], "\\\\", "\\", -1)

	parts = parts[2:]

	returntype := ""
	for _, part := range parts {
		if strings.Contains(part, ":") {
			colon := strings.Index(part, ":")
			field := part[:colon]
			value := strings.TrimSpace(part[colon+1:])
			switch field {
			case "kind":
				tag.Kind = value
			case "line":
				val, _ := strconv.Atoi(value)
				// TODO: Check err from strconv.Atoi
				tag.Line = val
			case "typeref":
				tag.Typeref = value
			case "signature":
				tag.Signature = value
			case "returntype":
				returntype = value
			case "class":
				tag.Class = value
			case "struct":
				tag.Struct = value
			case "namespace":
				tag.Namespace = value
			}
		}
	}
	tag.Prototype = returntype + " " + tag.FunctionName + tag.Signature + ";"

	if strings.Contains(row, "/^") && strings.Contains(row, "$/;") {
		tag.Code = row[strings.Index(row, "/^")+2 : strings.Index(row, "$/;")]
	}

	return tag
}

func removeEmpty(rows []string) []string {
	var newRows []string
	for _, row := range rows {
		row = strings.TrimSpace(row)
		if len(row) > 0 {
			newRows = append(newRows, row)
		}
	}

	return newRows
}
