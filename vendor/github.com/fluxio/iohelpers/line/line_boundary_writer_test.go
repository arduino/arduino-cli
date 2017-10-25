package line

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type capturedWriters []string

func (c *capturedWriters) Write(p []byte) (int, error) {
	*c = append(*c, string(p))
	return len(p), nil
}

func TestBoundaryWriter(t *testing.T) {
	Convey("BoundaryWriter", t, func() {
		Convey("should pass completely written lines through", func() {
			var lines capturedWriters
			w := BoundaryWriter{Target: &lines}

			n, _ := fmt.Fprintf(&w, "abc\ndef\nxyz")
			So(n, ShouldEqual, 11)
			So(lines, ShouldResemble, capturedWriters{"abc\ndef\n"})

			// finish off that last line
			n, _ = fmt.Fprintf(&w, "\n---")
			So(n, ShouldEqual, 4)
			So(lines, ShouldResemble, capturedWriters{"abc\ndef\n", "xyz\n"})
		})
		Convey("should buffer partial lines until flushed", func() {
			var lines capturedWriters
			w := BoundaryWriter{Target: &lines}
			var n int
			n, _ = fmt.Fprintf(&w, "ab")
			So(n, ShouldEqual, 2)
			So(len(lines), ShouldEqual, 0)
			n, _ = fmt.Fprintf(&w, "c")
			So(n, ShouldEqual, 1)
			n, _ = fmt.Fprintf(&w, "\n")
			So(n, ShouldEqual, 1)
			So(lines, ShouldResemble, capturedWriters{"abc\n"})
			n, _ = fmt.Fprintf(&w, "de")
			So(n, ShouldEqual, 2)
			n, _ = fmt.Fprintf(&w, "f\nxy")
			So(n, ShouldEqual, 4)
			n, _ = fmt.Fprintf(&w, "z")
			So(n, ShouldEqual, 1)
			So(lines, ShouldResemble, capturedWriters{"abc\n", "def\n"})
			w.Flush()
			So(lines, ShouldResemble, capturedWriters{"abc\n", "def\n", "xyz"})
		})
	})
}
