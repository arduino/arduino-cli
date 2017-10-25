package iohelpers

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestCopyNRunes(t *testing.T) {
	// testCase describes one test case
	type testCase struct {
		// s is the input string to be copied
		s string
		// expected describes expected outputs for various
		// values of the n argument to CopyNRunes.
		expected map[int64]string
	}
	for i, tc := range []testCase{
		{"", map[int64]string{
			1:    "",
			1000: "",
		}},
		{"null", map[int64]string{
			1:    "n",
			1000: "null",
		}},
		{"\x00\x00\x00\x00\x00", map[int64]string{
			1:    "\x00",
			1000: "\x00\x00\x00\x00\x00",
		}},
		{"串 串 串 kebab", map[int64]string{
			1:    "串",
			1000: "串 串 串 kebab",
		}},
		{strings.Repeat("X", 1000), map[int64]string{
			1:    "X",
			1000: strings.Repeat("X", 1000),
		}},
		{strings.Repeat("Y", 10000), map[int64]string{
			1:    "Y",
			1000: strings.Repeat("Y", 1000),
		}},
	} {
		sRuneLen := utf8.RuneCountInString(tc.s)

		// Try copying 0, or a negative rune count.
		for _, n := range []int64{
			0,
			-123,
		} {
			var buf bytes.Buffer
			bytes, runes, err := CopyNRunes(&buf, strings.NewReader(tc.s), n)
			// We should never error (even EOF) when copying 0 bytes.
			if err != nil {
				t.Fatalf("case %d: error during copy: %s", i, err)
			}
			if bytes != 0 {
				t.Errorf("case %d: Should have copied 0 bytes of %q; copied %d",
					i, tc.s, bytes)
			}
			if runes != 0 {
				t.Errorf("case %d: Should have copied 0 runes of %q; copied %d",
					i, tc.s, runes)
			}
		}

		// Try copying various hunk sizes of runes.
		for _, n := range []int64{
			1,
			1000,
		} {
			var buf bytes.Buffer
			bytes, runes, err := CopyNRunes(&buf, strings.NewReader(tc.s), n)
			errPrefix := fmt.Sprintf("case %d (n=%d)", i, n)
			if int64(sRuneLen) < n {
				if err != io.EOF {
					t.Errorf("%s: expected EOF reading %d bytes, got %s",
						errPrefix, n, err)
				}
			} else {
				if err != nil {
					t.Fatalf("%s: error during %d-rune copy: %s",
						errPrefix, n, err)
				}
			}
			expectedBytes := len(tc.expected[n])
			if bytes != int64(expectedBytes) {
				t.Errorf("%s: Should copy %d bytes of %q, but copied %d",
					errPrefix, expectedBytes, tc.s, bytes)
			}
			expectedRunes := utf8.RuneCountInString(tc.expected[n])
			if runes != int64(expectedRunes) {
				t.Errorf("%s: Should copy %d runes of %q, but copied %d",
					errPrefix, expectedRunes, tc.s, runes)
			}
			bufStr := buf.String()
			if bufStr != tc.expected[n] {
				t.Errorf("%s: Copy doesn't match original; expected %q, got %q",
					errPrefix, tc.expected[n], bufStr)
			}
		}
	}
}
