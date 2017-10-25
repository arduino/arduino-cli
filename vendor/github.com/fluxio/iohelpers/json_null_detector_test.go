package iohelpers

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// shouldDetectJsonNull is a goconvey assertion that checks that
// JsonNullDetector's Detected() returns the expected value when the full
// contents of actual (which must be a string) are read as input.
var shouldDetectJsonNull = func(actual interface{}, expected ...interface{}) string {
	success := ""

	s, ok := actual.(string)
	if !ok {
		return fmt.Sprintf("shouldDetectJsonNull accepts a string; got: %T: %v",
			actual, actual)
	}

	if len(expected) != 1 {
		return fmt.Sprintf("shouldDetectJsonNull takes 1 expected value, got %d: %v",
			len(expected), expected)
	}
	shouldDetect, ok := expected[0].(bool)
	if !ok {
		return fmt.Sprintf("shouldDetectJsonNull takes a bool expected value, got %T: %v",
			expected, expected)
	}

	d := &JsonNullDetector{Reader: strings.NewReader(s)}
	readBytes, err := ioutil.ReadAll(d)
	if err != nil {
		return fmt.Sprintf("Error reading actual value: %v", err)
	}
	if string(readBytes) != s {
		return fmt.Sprintf("Expected to pass through %q in read, but got %q", s, readBytes)
	}

	detected := d.Detected()
	if detected != shouldDetect {
		return fmt.Sprintf("With %q, JsonNullDetector.Detected() should be %v, but got %v",
			s, shouldDetect, detected)
	}

	return success
}

func TestJsonNullDetector(t *testing.T) {
	Convey("Detected() should incrementally parse a bare null correctly", t, func() {
		d := &JsonNullDetector{Reader: strings.NewReader("null")}

		// Initially, nothing has been read.
		So(d.Detected(), ShouldBeFalse)

		// Read the first 2 bytes; null not detected yet.
		buf := make([]byte, 2)
		n, err := d.Read(buf)
		So(err, ShouldBeNil)
		So(n, ShouldEqual, 2)
		So(d.Detected(), ShouldBeFalse)

		// Read the rest; null should be detected.
		_, err = ioutil.ReadAll(d)
		So(err, ShouldBeNil)
		So(d.Detected(), ShouldBeTrue)
	})

	Convey("Detected() should not detect null on various not-'null' strings", t, func() {
		for _, s := range []string{
			"",
			"    ",
			"    \t\n   ",
			"n",
			"nu",
			"nul",
			"nulll",
			"Null",
			"NULL",
			"n u l l",
			"{null}",
			"{}",
			"[null]",
			"[]   null",
			"abcdef",
			`"null"`,
			`{"a":"null"}`,

			// Other high-byte characters.
			"ñull",
			"   null      \u08aa", // ARABIC LETTER REH WITH LOOP
			"\U00010391   null  ", // UGARITIC LETTER ZU

			// Exotic whitespace
			// (does not count as JSON syntactic whitespace!)
			"\v\v\v\vnull",
			"\u0085null",      // NEXT LINE
			"null  \u00a0",    // NO-BREAK SPACE
			"  null \u1680",   // OGHAM SPACE MARK
			"   \u205F  null", // MEDIUM MATHEMATICAL SPACE
			"  null\u3000",    // IDEOGRAPHIC SPACE
		} {
			So(s, shouldDetectJsonNull, false)
		}
	})

	Convey("Detected() should permit JSON whitespace around null", t, func() {
		for _, s := range []string{
			"   null   ",
			"\tnull",
			"null\n",
			"\rnull\r\n\t",
		} {
			So(s, shouldDetectJsonNull, true)
		}
	})
}
