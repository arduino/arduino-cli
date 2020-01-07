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
	"bufio"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/legacy/builder/types"
)

func (p *CTagsParser) FixCLinkageTagsDeclarations(tags []*types.CTag) {

	linesMap := p.FindCLinkageLines(tags)
	for i, _ := range tags {

		if sliceContainsInt(linesMap[tags[i].Filename], tags[i].Line) &&
			!strings.Contains(tags[i].PrototypeModifiers, EXTERN) {
			tags[i].PrototypeModifiers = tags[i].PrototypeModifiers + " " + EXTERN
		}
	}
}

func sliceContainsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (p *CTagsParser) prototypeAndCodeDontMatch(tag *types.CTag) bool {
	if tag.SkipMe {
		return true
	}

	code := removeSpacesAndTabs(tag.Code)

	if strings.Index(code, ")") == -1 {
		// Add to code non-whitespace non-comments tokens until we find a closing round bracket
		file, err := os.Open(tag.Filename)
		if err == nil {
			defer file.Close()

			scanner := bufio.NewScanner(file)
			line := 1

			// skip lines until we get to the start of this tag
			for scanner.Scan() && line < tag.Line {
				line++
			}

			// read up to 10 lines in search of a closing paren
			multilinecomment := false
			temp := ""

			code, multilinecomment = removeComments(scanner.Text(), multilinecomment)
			for scanner.Scan() && line < (tag.Line+10) && strings.Index(temp, ")") == -1 {
				temp, multilinecomment = removeComments(scanner.Text(), multilinecomment)
				code += temp
			}
		}
	}

	code = removeSpacesAndTabs(code)

	prototype := removeSpacesAndTabs(tag.Prototype)
	prototype = removeTralingSemicolon(prototype)

	// Prototype matches exactly with the code?
	ret := strings.Index(code, prototype)

	if ret == -1 {
		// If the definition is multiline ctags uses the function name as line number
		// Try to match functions in the form
		// void
		// foo() {}

		// Add to code n non-whitespace non-comments tokens before the code line

		code = removeEverythingAfterClosingRoundBracket(code)
		// Get how many characters are "missing"
		n := strings.Index(prototype, code)
		line := 0
		// Add these characters to "code" string
		code, line = getFunctionProtoWithNPreviousCharacters(tag, code, n)
		// Check again for perfect matching
		ret = strings.Index(code, prototype)
		if ret != -1 {
			tag.Line = line
		}
	}

	return ret == -1
}

func findTemplateMultiline(tag *types.CTag) string {
	code, _ := getFunctionProtoUntilTemplateToken(tag, tag.Code)
	return removeEverythingAfterClosingRoundBracket(code)
}

func removeEverythingAfterClosingRoundBracket(s string) string {
	n := strings.Index(s, ")")
	return s[0 : n+1]
}

func getFunctionProtoUntilTemplateToken(tag *types.CTag, code string) (string, int) {

	/* FIXME I'm ugly */
	line := 0

	file, err := os.Open(tag.Filename)
	if err == nil {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		multilinecomment := false
		var textBuffer []string

		// buffer lines until we get to the start of this tag
		for scanner.Scan() && line < (tag.Line-1) {
			line++
			text := scanner.Text()
			textBuffer = append(textBuffer, text)
		}

		for line > 0 && !strings.Contains(code, TEMPLATE) {

			line = line - 1
			text := textBuffer[line]

			text, multilinecomment = removeComments(text, multilinecomment)

			code = text + code
		}
	}
	return code, line
}

func getFunctionProtoWithNPreviousCharacters(tag *types.CTag, code string, n int) (string, int) {

	/* FIXME I'm ugly */
	expectedPrototypeLen := len(code) + n
	line := 0

	file, err := os.Open(tag.Filename)
	if err == nil {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		multilinecomment := false
		var textBuffer []string

		// buffer lines until we get to the start of this tag
		for scanner.Scan() && line < (tag.Line-1) {
			line++
			text := scanner.Text()
			textBuffer = append(textBuffer, text)
		}

		for line > 0 && len(code) < expectedPrototypeLen {

			line = line - 1
			text := textBuffer[line]

			text, multilinecomment = removeComments(text, multilinecomment)

			code = text + code
			code = removeSpacesAndTabs(code)
		}
	}
	return code, line
}

func removeComments(text string, multilinecomment bool) (string, bool) {
	// Remove C++ style comments
	if strings.Index(text, "//") != -1 {
		text = text[0:strings.Index(text, "//")]
	}

	// Remove C style comments
	if strings.Index(text, "*/") != -1 {
		if strings.Index(text, "/*") != -1 {
			// C style comments on the same line
			text = text[0:strings.Index(text, "/*")] + text[strings.Index(text, "*/")+1:len(text)-1]
		} else {
			text = text[strings.Index(text, "*/")+1 : len(text)-1]
			multilinecomment = true
		}
	}

	if multilinecomment {
		if strings.Index(text, "/*") != -1 {
			text = text[0:strings.Index(text, "/*")]
			multilinecomment = false
		} else {
			text = ""
		}
	}
	return text, multilinecomment
}

/* This function scans the source files searching for "extern C" context
 * It save the line numbers in a map filename -> {lines...}
 */
func (p *CTagsParser) FindCLinkageLines(tags []*types.CTag) map[string][]int {

	lines := make(map[string][]int)

	for _, tag := range tags {

		if lines[tag.Filename] != nil {
			break
		}

		file, err := os.Open(tag.Filename)
		if err == nil {
			defer file.Close()

			lines[tag.Filename] = append(lines[tag.Filename], -1)

			scanner := bufio.NewScanner(file)

			// we can't remove the comments otherwise the line number will be wrong
			// there are three cases:
			// 1 - extern "C" void foo()
			// 2 - extern "C" {
			//		void foo();
			//		void bar();
			//	}
			// 3 - extern "C"
			//	{
			//		void foo();
			//		void bar();
			//	}
			// case 1 and 2 can be simply recognized with string matching and indent level count
			// case 3 needs specia attention: if the line ONLY contains `extern "C"` string, don't bail out on indent level = 0

			inScope := false
			enteringScope := false
			indentLevels := 0
			line := 0

			externCDecl := removeSpacesAndTabs(EXTERN)

			for scanner.Scan() {
				line++
				str := removeSpacesAndTabs(scanner.Text())

				if len(str) == 0 {
					continue
				}

				// check if we are on the first non empty line after externCDecl in case 3
				if enteringScope == true {
					enteringScope = false
				}

				// check if the line contains externCDecl
				if strings.Contains(str, externCDecl) {
					inScope = true
					if len(str) == len(externCDecl) {
						// case 3
						enteringScope = true
					}
				}
				if inScope == true {
					lines[tag.Filename] = append(lines[tag.Filename], line)
				}
				indentLevels += strings.Count(str, "{") - strings.Count(str, "}")

				// Bail out if indentLevel is zero and we are not in case 3
				if indentLevels == 0 && enteringScope == false {
					inScope = false
				}
			}
		}

	}
	return lines
}
