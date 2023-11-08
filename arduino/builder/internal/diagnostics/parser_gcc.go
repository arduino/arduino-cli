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

package diagnostics

import (
	"strconv"
	"strings"
)

// Parse output from gcc compiler and extract diagnostics
func parseGccOutput(output []string) ([]*Diagnostic, error) {
	// Output from gcc is a mix of diagnostics and other information.
	//
	// 1. include trace lines:
	//
	//   In file included from /home/megabug/Arduino/libraries/Audio/src/Audio.h:16:0,
	//   ·················from /home/megabug/Arduino/Blink/Blink.ino:1:
	//
	// 2. in-file context lines:
	//
	//   /home/megabug/Arduino/libraries/Audio/src/DAC.h: In member function 'void DACClass::enableInterrupts()':
	//
	// 3. actual diagnostic lines:
	//
	//   /home/megabug/Arduino/libraries/Audio/src/DAC.h:31:44: fatal error: 'isrId' was not declared in this scope
	//
	//   /home/megabug/Arduino/libraries/Audio/src/DAC.h:31:44: error: 'isrId' was not declared in this scope
	//
	//   /home/megabug/Arduino/libraries/Audio/src/DAC.h:31:44: warning: 'isrId' was not declared in this scope
	//
	// 4. annotations or suggestions:
	//
	//   /home/megabug/Arduino/Blink/Blink.ino:4:1: note: suggested alternative: 'rand'
	//
	// 5. extra context lines with an extract of the code that errors refers to:
	//
	//   ·asd;
	//   ·^~~
	//   ·rand
	//
	//   ·void enableInterrupts()  { NVIC_EnableIRQ(isrId); };
	//   ···········································^~~~~

	var fullContext FullContext
	var fullContextRefersTo string
	var inFileContext *Context
	var currentDiagnostic *Diagnostic
	var currentMessage *string
	var res []*Diagnostic

	for _, in := range output {
		isTrace := false
		if strings.HasPrefix(in, "In file included from ") {
			in = strings.TrimPrefix(in, "In file included from ")
			// 1. include trace
			isTrace = true
			inFileContext = nil
			fullContext = nil
			fullContextRefersTo = ""
		} else if strings.HasPrefix(in, "                 from ") {
			in = strings.TrimPrefix(in, "                 from ")
			// 1. include trace continuation
			isTrace = true
		}
		if isTrace {
			in = strings.TrimSuffix(in, ",")
			file, line, col := extractFileLineAndColumn(in)
			context := &Context{
				File:    file,
				Line:    line,
				Column:  col,
				Message: "included from here",
			}
			currentMessage = &context.Message
			fullContext = append(fullContext, context)
			continue
		}

		if split := strings.SplitN(in, ": ", 2); len(split) == 2 {
			file, line, column := extractFileLineAndColumn(split[0])
			msg := split[1]

			if line == 0 && column == 0 {
				// 2. in-file context
				inFileContext = &Context{
					Message: msg,
					File:    file,
				}
				currentMessage = &inFileContext.Message
				continue
			}

			if strings.HasPrefix(msg, "note: ") {
				msg = strings.TrimPrefix(msg, "note: ")
				// 4. annotations or suggestions
				if currentDiagnostic != nil {
					suggestion := &Note{
						Message: msg,
						File:    file,
						Line:    line,
						Column:  column,
					}
					currentDiagnostic.Suggestions = append(currentDiagnostic.Suggestions, suggestion)
					currentMessage = &suggestion.Message
				}
				continue
			}

			severity := SeverityUnspecified
			if strings.HasPrefix(msg, "error: ") {
				msg = strings.TrimPrefix(msg, "error: ")
				severity = SeverityError
			} else if strings.HasPrefix(msg, "warning: ") {
				msg = strings.TrimPrefix(msg, "warning: ")
				severity = SeverityWarning
			} else if strings.HasPrefix(msg, "fatal error: ") {
				msg = strings.TrimPrefix(msg, "fatal error: ")
				severity = SeverityFatal
			}
			if severity != SeverityUnspecified {
				// 3. actual diagnostic lines
				currentDiagnostic = &Diagnostic{
					Severity: severity,
					Message:  msg,
					File:     file,
					Line:     line,
					Column:   column,
				}
				currentMessage = &currentDiagnostic.Message

				if len(fullContext) > 0 {
					if fullContextRefersTo == "" || fullContextRefersTo == file {
						fullContextRefersTo = file
						currentDiagnostic.Context = append(currentDiagnostic.Context, fullContext...)
					}
				}
				if inFileContext != nil && inFileContext.File == file {
					currentDiagnostic.Context = append(currentDiagnostic.Context, inFileContext)
				}

				res = append(res, currentDiagnostic)
				continue
			}
		}

		// 5. extra context lines
		if strings.HasPrefix(in, " ") {
			if currentMessage != nil {
				*currentMessage += "\n" + in
			}
			continue
		}
	}
	return res, nil
}

func extractFileLineAndColumn(file string) (string, int, int) {
	split := strings.Split(file, ":")
	file = split[0]
	if len(split) == 1 {
		return file, 0, 0
	}

	// Special case: handle Windows drive letter `C:\...`
	if len(split) > 1 && len(file) == 1 && strings.HasPrefix(split[1], `\`) {
		file += ":" + split[1]
		split = split[1:]

		if len(split) == 1 {
			return file, 0, 0
		}
	}

	line, err := strconv.Atoi(split[1])
	if err != nil || len(split) == 2 {
		return file, line, 0
	}
	column, _ := strconv.Atoi(split[2])
	return file, line, column
}
