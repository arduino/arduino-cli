package line

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func Write(dest io.Writer, str ...string) error {
	for _, s := range str {
		_, err := fmt.Fprint(dest, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestSquelchingWriter(t *testing.T) {
	var dest bytes.Buffer
	filter := NewSquelchingWriter(&dest, CompileOrDie("trace", "debug"))

	err := Write(filter,
		"[info] ", "hi!\n",
		"this line ", " gets filtered ", " because of [trace]\n",
		"this de", "bug line also dies\n",
		"t", "r", "a", "c", "e\n",
		"this ", "line", " is OK\n")
	if err != nil {
		t.Fatal(err)
	}
	if dest.String() != "[info] hi!\nthis line is OK\n" {
		t.Fatalf("Expected one line, got: [%s]", dest.String())
	}
}
