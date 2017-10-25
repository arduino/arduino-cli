// Copyright (c) 2012-2014 The Goproperties Authors.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
// of the Software, and to permit persons to whom the Software is furnished to do
// so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

/*
Package goproperties implements read operations on .properties file.

.properties is a file extension for files mainly used in Java related
technologies to store the configurable parameters of an application.
They can also be used for storing strings for Internationalization and
localization; these are known as Property Resource Bundles.

Each parameter is stored as a pair of strings, one storing the name of
the parameter (called the key), and the other storing the value.

Each line in a .properties file normally stores a single property.
Several formats are possible for each line, including key=value,
key = value, key:value, and key value.

.properties files can use the number sign (#) or the exclamation mark (!)
as the first non blank character in a line to denote that all text following
it is a comment. The backwards slash is used to escape a character.
An example of a properties file is provided below.

    # You are reading the ".properties" entry.
    ! The exclamation mark can also mark text as comments.
    website = http://en.wikipedia.org/
    language = English
    # The backslash below tells the application to continue reading
    # the value onto the next line.
    message = Welcome to \
              Wikipedia!
    # Add spaces to the key
    key\ with\ spaces = This is the value that could be looked up with the \
    key "key with spaces".
    # Empty lines are skipped

    # Unicode
    unicode=\u041f\u0440\u0438\u0432\u0435\u0442, \u0421\u043e\u0432\u0430!
    # Comment

In the example above, website would be a key, and its corresponding
value would be http://en.wikipedia.org/. While the number sign and the
exclamation mark marks text as comments, it has no effect when it is part
of a property. Thus, the key message has the value Welcome to Wikipedia!
and not Welcome to Wikipedia. Note also that all of the whitespace in
front of Wikipedia! is excluded completely.

The encoding of a .properties file is ISO-8859-1, also known as Latin-1.
All non-Latin-1 characters must be entered by using Unicode escape characters,
e. g. \uHHHH where HHHH is a hexadecimal index of the character in the Unicode
character set. This allows for using .properties files as resource bundles for
localization. A non-Latin-1 text file can be converted to a correct .properties
file by using the native2ascii tool that is shipped with the JDK or by using
a tool, such as po2prop, that manages the transformation from a bilingual
localization format into .properties escaping.

From Wikipedia, the free encyclopedia
http://en.wikipedia.org/wiki/.properties
*/
package properties

import (
	"errors"
	"io"
	"os"
	"strconv"
	"unicode/utf8"
)

type Properties map[string]string

// Creates an instance of Properties and try to fill it with data from file.
// It's safe to ignore error as method always return pointer to the created
// instance and close any opened resources.
func Load(file string) (Properties, error) {
	p := make(Properties)
	f, err := os.Open(file)
	if err != nil {
		return p, err
	}
	defer f.Close()
	if err := p.Load(f); err != nil {
		return p, err
	}
	return p, nil
}

// Uses strconv to convert key's value to bool. Returns def if
// conversion failed or key does not exist.
func (p Properties) Bool(key string, def bool) bool {
	if v, found := p[key]; found {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

// Uses strconv to convert key's value to float64. Returns def if
// conversion failed or key does not exist.
func (p Properties) Float(key string, def float64) float64 {
	if v, found := p[key]; found {
		if b, err := strconv.ParseFloat(v, 64); err == nil {
			return b
		}
	}
	return def
}

// Uses strconv to convert key's value to int64 (base is 0). Returns def if
// conversion failed or key does not exist.
func (p Properties) Int(key string, def int64) int64 {
	if v, found := p[key]; found {
		if b, err := strconv.ParseInt(v, 0, 64); err == nil {
			return b
		}
	}
	return def
}

// Uses strconv to convert key's value to uint64 (base is 0). Returns def if
// conversion failed or key does not exist.
func (p Properties) Uint(key string, def uint64) uint64 {
	if v, found := p[key]; found {
		if b, err := strconv.ParseUint(v, 0, 64); err == nil {
			return b
		}
	}
	return def
}

// Returns def if key does not exist.
func (p Properties) String(key string, def string) string {
	if v, found := p[key]; found {
		return v
	}
	return def
}

// ErrMalformedUtf8Encoding means that it was not possible to convert \uXXXX
// string to utf8 rune.
var ErrMalformedUtf8Encoding error = errors.New("malformed \\uxxxx encoding")

// Reads key value pairs from reader and returns map[string]string
// If source has key already defined then existed value replaced with new one
func (p Properties) Load(src io.Reader) error {
	lr := newLineReader(src)
	for {
		s, e := lr.readLine()
		if e == io.EOF {
			break
		}
		if e != nil {
			return e
		}

		keyLen := 0
		precedingBackslash := false
		hasSep := false
		valueStart := len(s)

		for keyLen < len(s) {
			c := s[keyLen]

			if (c == '=' || c == ':') && !precedingBackslash {
				valueStart = keyLen + 1
				hasSep = true
				break
			}
			if (c == ' ' || c == '\t' || c == '\f') && !precedingBackslash {
				valueStart = keyLen + 1
				break
			}
			if c == '\\' {
				precedingBackslash = !precedingBackslash
			} else {
				precedingBackslash = false
			}

			keyLen++
		}

		for valueStart < len(s) {
			c := s[valueStart]
			if c != ' ' && c != '\t' && c != '\f' {
				if !hasSep && (c == '=' || c == ':') {
					hasSep = true
				} else {
					break
				}
			}
			valueStart++
		}
		key, err := decodeString(s[0:keyLen])
		if err != nil {
			return err
		}
		value, err := decodeString(s[valueStart:len(s)])
		if err != nil {
			return err
		}
		p[key] = value
	}
	return nil
}

// Decodes \t,\n,\r,\f and \uXXXX characters in string
func decodeString(in string) (string, error) {
	out := make([]byte, len(in))
	o := 0
	for i := 0; i < len(in); {
		if in[i] == '\\' {
			i++
			switch in[i] {
			case 'u':
				i++
				utf8rune := 0
				for j := 0; j < 4; j++ {
					switch {
					case in[i] >= '0' && in[i] <= '9':
						utf8rune = (utf8rune << 4) + int(in[i]) - '0'
						break
					case in[i] >= 'a' && in[i] <= 'f':
						utf8rune = (utf8rune << 4) + 10 + int(in[i]) - 'a'
						break
					case in[i] >= 'A' && in[i] <= 'F':
						utf8rune = (utf8rune << 4) + 10 + int(in[i]) - 'A'
						break
					default:
						return "", ErrMalformedUtf8Encoding
					}
					i++
				}
				bytes := make([]byte, utf8.RuneLen(rune(utf8rune)))
				bytesWritten := utf8.EncodeRune(bytes, rune(utf8rune))
				for j := 0; j < bytesWritten; j++ {
					out[o] = bytes[j]
					o++
				}
				continue
			case 't':
				out[o] = '\t'
				o++
				i++
				continue
			case 'r':
				out[o] = '\r'
				o++
				i++
				continue
			case 'n':
				out[o] = '\n'
				o++
				i++
				continue
			case 'f':
				out[o] = '\f'
				o++
				i++
				continue
			}
			out[o] = in[i]
			o++
			i++
			continue
		}
		out[o] = in[i]
		o++
		i++
	}

	return string(out[0:o]), nil
}

// Read in a "logical line" from an InputStream/Reader, skip all comment
// and blank lines and filter out those leading whitespace characters
// (\u0020, \u0009 and \u000c) from the beginning of a "natural line".
type lineReader struct {
	reader     io.Reader
	buffer     []byte
	lineBuffer []byte
	limit      int
	offset     int
	exhausted  bool
}

func newLineReader(r io.Reader) *lineReader {
	n := new(lineReader)
	n.reader = r
	n.buffer = make([]byte, 1024)
	n.lineBuffer = make([]byte, 1024)
	n.limit = 0
	n.offset = 0
	n.exhausted = false
	return n
}

// Returns the "logical line" from given reader
func (lr *lineReader) readLine() (line string, e error) {
	if lr.exhausted {
		return "", io.EOF
	}
	nextCharIndex := 0
	char := byte(0)

	skipLF := false
	skipWhiteSpace := true
	appendedLineBegin := false
	isNewLine := true
	isCommentLine := false
	precedingBackslash := false

	for {
		if lr.offset >= lr.limit {
			lr.limit, e = io.ReadFull(lr.reader, lr.buffer)
			lr.offset = 0
			if e == io.EOF {
				lr.exhausted = true
				if isCommentLine {
					return "", io.EOF
				}
				return string(lr.lineBuffer[0:nextCharIndex]), nil
			}
			if e == io.ErrUnexpectedEOF {
				continue
			}
			if e != nil {
				lr.exhausted = true
				return "", e
			}
		}

		char = lr.buffer[lr.offset]
		lr.offset++

		if skipLF {
			skipLF = false
			if char == '\n' {
				continue
			}
		}

		if skipWhiteSpace {
			if char == ' ' || char == '\t' || char == '\f' {
				continue
			}
			if !appendedLineBegin && (char == '\r' || char == '\n') {
				continue
			}
			skipWhiteSpace = false
			appendedLineBegin = false
		}

		if isNewLine {
			isNewLine = false
			if char == '#' || char == '!' {
				isCommentLine = true
				continue
			}
		}

		if char != '\n' && char != '\r' {
			lr.lineBuffer[nextCharIndex] = char
			nextCharIndex++
			if nextCharIndex == len(lr.lineBuffer) {
				newBuffer := make([]byte, len(lr.lineBuffer)*2)
				copy(lr.lineBuffer, newBuffer)
				lr.lineBuffer = newBuffer
			}
			//flip the preceding backslash flag
			precedingBackslash = char == '\\' && !precedingBackslash
		} else {
			// reached EOL
			if isCommentLine || nextCharIndex == 0 {
				isCommentLine = false
				isNewLine = true
				skipWhiteSpace = true
				nextCharIndex = 0
				continue
			}
			if lr.offset >= lr.limit {
				lr.limit, e = io.ReadFull(lr.reader, lr.buffer)
				lr.offset = 0
				if e != nil {
					lr.exhausted = true
					return string(lr.lineBuffer[0:nextCharIndex]), nil
				}
			}
			if precedingBackslash {
				nextCharIndex--
				//skip the leading whitespace characters in following line
				skipWhiteSpace = true
				appendedLineBegin = true
				precedingBackslash = false
				if char == '\r' {
					skipLF = true
				}
			} else {
				return string(lr.lineBuffer[0:nextCharIndex]), nil
			}
		}
	}
}
