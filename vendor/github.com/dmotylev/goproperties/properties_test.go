// Copyright (c) 2012-2014 The Goproperties Authors.
//
// Permission is hereby granted, free of charge, to any person obtaining nl copy of
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

package properties

import (
	"bytes"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"testing"
	"testing/iotest"
)

func mustLoad(s string) Properties {
	p := make(Properties)
	err := p.Load(bytes.NewReader([]byte(s)))
	if err != nil {
		panic(err)
	}
	return p
}

func eqString(l, r string) bool {
	return l == r
}

func eqBool(l, r bool) bool {
	return l == r
}

func eqUint64(l, r uint64) bool {
	return l == r
}
func eqFloat64(l, r float64) bool {
	return l == r
}
func eqInt64(l, r int64) bool {
	return l == r
}

func TestGeneric(t *testing.T) {
	p := mustLoad(source)
	if !eqString(p["website"], "http://en.wikipedia.org/") {
		t.Fail()
	}
	if !eqString(p["language"], "English") {
		t.Fail()
	}
	if !eqString(p["message"], "Welcome to Wikipedia!") {
		t.Fail()
	}
	if !eqString(p["unicode"], "Привет, Сова!") {
		t.Fail()
	}
	if !eqString(p["key with spaces"], "This is the value that could be looked up with the key \"key with spaces\".") {
		t.Fail()
	}
}

func mkTempFile() *os.File {
	if f, err := ioutil.TempFile(os.TempDir(), "goproperties"); err != nil {
		panic(err)
		return nil
	} else {
		return f
	}
}

func TestLoadReal(t *testing.T) {
	f := mkTempFile()
	defer os.Remove(f.Name())
	_, err := Load(f.Name())
	if err != nil {
		t.FailNow()
	}
}

func TestLoadFiction(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "goproperties")
	if err != nil {
		t.Fatal(err)
	}
	os.Remove(f.Name())
	_, err = Load(f.Name())
	if err == nil {
		t.FailNow()
	}
	if matched, _ := regexp.MatchString(".*no such file.*", err.Error()); !matched {
		t.Errorf("got %s, want 'file not found'", err)
	}
}

func TestLoadTimeoutReader(t *testing.T) {
	p := make(Properties)
	err := p.Load(iotest.TimeoutReader(bytes.NewReader([]byte(source))))
	if err == nil {
		t.Fail()
	}
}

func testLoadMalformed(t *testing.T, s string) {
	p := make(Properties)
	err := p.Load(bytes.NewReader([]byte(s)))
	if err != ErrMalformedUtf8Encoding {
		t.Errorf("got %s, want %s", err, ErrMalformedUtf8Encoding)
	}
}

func TestLoadMalformedKey(t *testing.T) {
	testLoadMalformed(t, malformedKey)
}

func TestLoadMalformedValue(t *testing.T) {
	testLoadMalformed(t, malformedValue)
}

func TestString(t *testing.T) {
	p := mustLoad(source)
	if !eqString(p.String("string", "not found"), "found\t\n\r\f") {
		t.Fail()
	}
	if !eqString(p.String("missed", "not found"), "not found") {
		t.Fail()
	}
}

func TestBool(t *testing.T) {
	p := mustLoad(source)
	if !eqBool(p.Bool("bool", false), true) {
		t.Fail()
	}
	if !eqBool(p.Bool("missed", true), true) {
		t.Fail()
	}
}

func TestFloat(t *testing.T) {
	p := mustLoad(source)
	if !eqFloat64(p.Float("float", math.MaxFloat64), math.SmallestNonzeroFloat64) {
		t.Fail()
	}
	if !eqFloat64(p.Float("missed", math.MaxFloat64), math.MaxFloat64) {
		t.Fail()
	}
}

func TestInt(t *testing.T) {
	p := mustLoad(source)
	if !eqInt64(p.Int("int", math.MaxInt64), int64(math.MinInt64)) {
		t.Fail()
	}
	if !eqInt64(p.Int("missed", math.MaxInt64), int64(math.MaxInt64)) {
		t.Fail()
	}
	if !eqInt64(p.Int("hex", 0xCAFEBABE), int64(0xCAFEBABE)) {
		t.Fail()
	}
}

func TestUint(t *testing.T) {
	p := mustLoad(source)
	if !eqUint64(p.Uint("uint", 42), uint64(math.MaxUint64)) {
		t.Fail()
	}
	if !eqUint64(p.Uint("missed", 42), uint64(42)) {
		t.Fail()
	}
	if !eqUint64(p.Uint("hex", 0xCAFEBABE), uint64(0xCAFEBABE)) {
		t.Fail()
	}
}

// Test1024Comment verifies that if the lineReader is in the middle
// of parsing nl comment when it goes to read the last < 1024 byte block,
// it doesn't get confused and return EOF.
func Test1024Comment(t *testing.T) {
	config := "nl = b\n"
	for i := 0; i < 1024; i++ {
		config += "#"
	}
	config += "\nc = d\n"

	p := make(Properties)
	p.Load(bytes.NewReader([]byte(config)))
	if !eqString(p.String("nl", "not found"), "b") {
		t.Fail()
	}
	if !eqString(p.String("c", "not found"), "d") {
		t.Fail()
	}
}

func TestCornerCase1(t *testing.T) {
	s := "k=1\\\r\n2\nl=1\n"
	for i := 0; i < 1025; i++ {
		s += "#"
	}
	p := make(Properties)
	p.Load(iotest.DataErrReader(bytes.NewReader([]byte(s))))
	if !eqString(p.String("k", "-"), "12") {
		t.Fail()
	}
}

const (
	source = `\
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
unicode=\u041F\u0440\u0438\u0432\u0435\u0442, \u0421\u043e\u0432\u0430!
# Comment
string=found\t\n\r\f
bool=true
float=4.940656458412465441765687928682213723651e-324
int=-9223372036854775808
uint=18446744073709551615
hex=0xCAFEBABE
`
	malformedValue = `unicode=\uOHNO`
	malformedKey   = `\uOHNO=`
)
