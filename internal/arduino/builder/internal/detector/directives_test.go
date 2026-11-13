// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package detector

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirectiveLines(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		directives []string
	}{
		{
			name: "DirectivesAmongCode",
			source: "#include <Arduino.h>\r\n" +
				"#line 1 \"/tmp/sketch/sketch.ino\"\n" +
				"  #include \"config.h\"\n" +
				"void setup() {}\n" +
				"void loop() { Serial.println(\"#1\"); }\n" +
				"#endif",
			directives: []string{"#include <Arduino.h>", "#line 1 \"/tmp/sketch/sketch.ino\"", "#include \"config.h\"", "#endif"},
		},
		{
			name:       "PlainCode",
			source:     "void setup() {}\nvoid loop() {}\n",
			directives: []string{},
		},
		{
			name:       "DirectiveContinuedOnTheNextLine",
			source:     "#if defined(A) ||\\\ndefined(B)\n#endif\n",
			directives: []string{"#if defined(A) ||defined(B)", "#endif"},
		},
		{
			name:       "CommentsAroundDirective",
			source:     "/* once */ #define LIMIT 10 // ten\n",
			directives: []string{"#define LIMIT 10"},
		},
		{
			name:       "DirectiveInBlockComment",
			source:     "/*\n#include <SPI.h>\n*/\n#include <Wire.h>\n",
			directives: []string{"#include <Wire.h>"},
		},
		{
			name:       "DirectiveInLineComment",
			source:     "// #include <SPI.h>\n#include <Wire.h>\n",
			directives: []string{"#include <Wire.h>"},
		},
		{
			name:       "DirectiveInContinuedLineComment",
			source:     "// note \\\n#include <SPI.h>\n#include <Wire.h>\n",
			directives: []string{"#include <Wire.h>"},
		},
		{
			name:       "DirectiveOnContinuedCodeLine",
			source:     "int a = 1 + \\\n#include <SPI.h>\n;\n",
			directives: []string{},
		},
		{
			name:       "CommentMarksInString",
			source:     "const char *open = \"/*\";\n#include <SPI.h>\nconst char *close = \"*/\";\n",
			directives: []string{"#include <SPI.h>"},
		},
		{
			name:       "QuoteInCharacter",
			source:     "char quote = '\"'; /*\n#include <SPI.h>\n*/\n",
			directives: []string{},
		},
		{
			name:       "QuoteInNumber",
			source:     "long big = 1'000; /*\n#include <SPI.h>\n*/\n",
			directives: []string{},
		},
		{
			name:       "QuotesInHexNumber",
			source:     "long color = 0xFF'A5'00; /*\n#include <SPI.h>\n*/\n",
			directives: []string{},
		},
		{
			name:       "QuoteAfterNumber",
			source:     "#if 0\nthe board is 5' long /*\n#endif\n#include <SPI.h>\n/* */\n",
			directives: []string{"#if 0", "#endif", "#include <SPI.h>"},
		},
		{
			name:       "OldMacLineEnds",
			source:     "int a; // note\r#include <SPI.h>\rint b;\r",
			directives: []string{"#include <SPI.h>"},
		},
		{
			name:       "BytesThatAreNotText",
			source:     "#define NAME \"M\xfc\xfcller\"\n",
			directives: []string{"#define NAME \"M\ufffd\ufffdller\""},
		},
		{
			name:       "RawString",
			source:     "const char page[] = R\"html(\n#menu { color: #fff; } /* \"menu\"\n)html\";\n#include <SPI.h>\n",
			directives: []string{"#include <SPI.h>"},
		},
		{
			name:       "Digraph",
			source:     "%:include <SPI.h>\n",
			directives: []string{"%:include <SPI.h>"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.directives, directiveLines([]byte(test.source)))
		})
	}
}

func TestFollowCode(t *testing.T) {
	require.False(t, followCode([]string{"#include <SPI.h>", "#define LIMIT 10"}))
	require.True(t, followCode([]string{"#include <SPI.h>", "#if __LINE__ > 10"}))
	require.True(t, followCode([]string{"#define ID __COUNTER__"}))
}
