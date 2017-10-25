package line

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestPrefixWriter(t *testing.T) {
	var dest bytes.Buffer
	prefixer := &PrefixWriter{&dest, []byte("Hi! My name is "), false}
	fmt.Fprintln(prefixer, "Bob")
	fmt.Fprintln(prefixer, "Al")
	fmt.Fprintln(prefixer, "Frank")

	lines := strings.Split(dest.String(), "\n")
	if len(lines) != 4 {
		t.Fatal("Expected 3 output lines, got: ", dest.String())
	}
	if lines[0] != "Hi! My name is Bob" ||
		lines[1] != "Hi! My name is Al" ||
		lines[2] != "Hi! My name is Frank" ||
		lines[3] != "" {
		t.Fatalf("Output mismatch: %q", lines)
	}
}

func TestPrefixWriterWithPartialLines(t *testing.T) {
	var dest bytes.Buffer
	prefixer := &PrefixWriter{&dest, []byte("ZZ:"), false}
	fmt.Fprint(prefixer, "abc")
	fmt.Fprint(prefixer, "def")
	fmt.Fprint(prefixer, "ghi\n123")
	fmt.Fprint(prefixer, "\n456\n789\n")

	expected := "ZZ:abcdefghi\n" +
		"ZZ:123\n" +
		"ZZ:456\n" +
		"ZZ:789\n"

	if dest.String() != expected {
		t.Fatalf("Output mismatch: %q", dest.String())
	}
}

func TestPrefixWriterSkipsFirstLine(t *testing.T) {
	var dest bytes.Buffer
	prefixer := PrefixWriter{&dest, []byte(" > "), true}
	fmt.Fprintln(&prefixer, "abc\ndef\nghi")

	expected := "abc\n > def\n > ghi\n"
	if dest.String() != expected {
		t.Fatalf("Output mismatch: %q", dest.String())
	}
}
