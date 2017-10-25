// Copyright 2014 zieckey. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goini

import (
	"bytes"
	"path/filepath"
	"runtime"
    "os"
	"testing"

	"github.com/bmizerany/assert"
)

func Test1(t *testing.T) {

	filename := filepath.Join(getTestDataDir(t), "ini_parser_testfile.ini")
	ini := New()
	err := ini.ParseFile(filename)
	assert.Equal(t, nil, err)

	v, ok := ini.Get("mid")
	assert.Equal(t, v, "ac9219aa5232c4e519ae5fcb4d77ae5b")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("product")
	assert.Equal(t, v, "ppp")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("combo")
	assert.Equal(t, v, "ccc")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("aa")
	assert.Equal(t, v, "bb")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("axxxa")
	assert.Equal(t, v, "")
	assert.Equal(t, ok, false)

	m, ok := ini.GetKvmap("")
	assert.Equal(t, len(m), 7)
	assert.Equal(t, ok, true)

	n, ok := ini.GetKvmap("n")
	assert.Equal(t, len(n), 0)
	assert.Equal(t, ok, false)

	sss, ok := ini.GetKvmap("sss")
	assert.Equal(t, len(sss), 2)
	assert.Equal(t, ok, true)
	v, ok = ini.SectionGet("sss", "aa")
	assert.Equal(t, v, "bb")
	assert.Equal(t, ok, true)
	v, ok = ini.SectionGet("sss", "appext")
	assert.Equal(t, v, "ab=cd")
	assert.Equal(t, ok, true)

	i, ok := ini.SectionGetInt("ddd", "age")
	assert.Equal(t, i, 30)
	assert.Equal(t, ok, true)
	ini.Delete("ddd", "age")
	i, ok = ini.SectionGetInt("ddd", "age")
	assert.Equal(t, i, 0)
	assert.Equal(t, ok, false)
	ini.Delete("ddd", "age") // delete again
	i, ok = ini.SectionGetInt("ddd", "age")
	assert.Equal(t, i, 0)
	assert.Equal(t, ok, false)

	i, ok = ini.SectionGetInt("ddd", "agexxx")
	assert.Equal(t, ok, false)

	f, ok := ini.GetFloat("version")
	assert.Equal(t, f, 4.4)
	assert.Equal(t, ok, true)
	ini.Delete("", "version") // delete again
	f, ok = ini.GetFloat("version")
	assert.Equal(t, f, 0.0)
	assert.Equal(t, ok, false)

	f, ok = ini.SectionGetFloat("ddd", "height")
	assert.Equal(t, f, 175.6)
	assert.Equal(t, ok, true)

	f, ok = ini.SectionGetFloat("ddd", "heightxxx")
	assert.Equal(t, ok, false)

	b, ok := ini.SectionGetBool("ddd", "debug")
	assert.Equal(t, b, true)
	assert.Equal(t, ok, true)

	b, ok = ini.GetBool("debug")
	assert.Equal(t, b, false)
	assert.Equal(t, ok, true)

}

func Test2(t *testing.T) {

	filename := filepath.Join(getTestDataDir(t), "ini_parser_testfile.ini")
	ini := New()
	err := ini.ParseFile(filename)
	assert.Equal(t, nil, err)

	v, ok := ini.Get("mid")
	assert.Equal(t, v, "ac9219aa5232c4e519ae5fcb4d77ae5b")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("product")
	assert.Equal(t, v, "ppp")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("combo")
	assert.Equal(t, v, "ccc")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("aa")
	assert.Equal(t, v, "bb")
	assert.Equal(t, ok, true)

	m, ok := ini.GetKvmap("")
	assert.Equal(t, len(m), 7)
	assert.Equal(t, ok, true)

	v, ok = ini.Get("axxxa")
	assert.Equal(t, v, "")
	assert.Equal(t, ok, false)
	ini.Set("axxxa", "aval")
	v, ok = ini.Get("axxxa")
	assert.Equal(t, v, "aval")
	assert.Equal(t, ok, true)

	n, ok := ini.GetKvmap("n")
	assert.Equal(t, len(n), 0)
	assert.Equal(t, ok, false)

	sss, ok := ini.GetKvmap("sss")
	assert.Equal(t, len(sss), 2)
	assert.Equal(t, ok, true)
	v, ok = ini.SectionGet("sss", "aa")
	assert.Equal(t, v, "bb")
	assert.Equal(t, ok, true)
	v, ok = ini.SectionGet("sss", "appext")
	assert.Equal(t, v, "ab=cd")
	assert.Equal(t, ok, true)

	i, ok := ini.SectionGetInt("ddd", "age")
	assert.Equal(t, i, 30)
	assert.Equal(t, ok, true)
	ini.SectionSetInt("ddd", "age", 40)
	i, ok = ini.SectionGetInt("ddd", "age")
	assert.Equal(t, i, 40)
	assert.Equal(t, ok, true)

	i, ok = ini.SectionGetInt("ddd", "agexxx")
	assert.Equal(t, ok, false)
	ini.SectionSetInt("ddd", "agexxx", 1)
	i, ok = ini.SectionGetInt("ddd", "agexxx")
	assert.Equal(t, i, 1)
	assert.Equal(t, ok, true)

	f, ok := ini.GetFloat("version")
	assert.Equal(t, f, 4.4)
	assert.Equal(t, ok, true)
	ini.SetFloat("version", 5.5)
	f, ok = ini.GetFloat("version")
	assert.Equal(t, f, 5.5)
	assert.Equal(t, ok, true)

	f, ok = ini.SectionGetFloat("ddd", "height")
	assert.Equal(t, f, 175.6)
	assert.Equal(t, ok, true)
	ini.SectionSetFloat("ddd", "height", 160.1)
	f, ok = ini.SectionGetFloat("ddd", "height")
	assert.Equal(t, f, 160.1)
	assert.Equal(t, ok, true)

	f, ok = ini.SectionGetFloat("ddd", "heightxxx")
	assert.Equal(t, ok, false)

	b, ok := ini.SectionGetBool("ddd", "debug")
	assert.Equal(t, b, true)
	assert.Equal(t, ok, true)
	ini.SectionSetBool("ddd", "debug", false)
	b, ok = ini.SectionGetBool("ddd", "debug")
	assert.Equal(t, b, false)
	assert.Equal(t, ok, true)

	b, ok = ini.GetBool("debug")
	assert.Equal(t, b, false)
	assert.Equal(t, ok, true)
	ini.SetBool("debug", true)
	b, ok = ini.GetBool("debug")
	assert.Equal(t, b, true)
	assert.Equal(t, ok, true)

	ini.SectionSet("asec", "a", "aval")
	v, ok = ini.SectionGet("asec", "a")
	assert.Equal(t, v, "aval")
	assert.Equal(t, ok, true)

	ini.SectionSet("asec", "a", "bval")
	v, ok = ini.SectionGet("asec", "a")
	assert.Equal(t, v, "bval")
	assert.Equal(t, ok, true)
}

func Test3(t *testing.T) {
	filename := filepath.Join(getTestDataDir(t), "ini_parser_testfile.ini")
	ini := New()
	err := ini.ParseFile(filename)
	assert.Equal(t, nil, err)

	v, ok := ini.Get("mid")
	assert.Equal(t, v, "ac9219aa5232c4e519ae5fcb4d77ae5b")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("product")
	assert.Equal(t, v, "ppp")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("combo")
	assert.Equal(t, v, "ccc")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("aa")
	assert.Equal(t, v, "bb")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("axxxa")
	assert.Equal(t, v, "")
	assert.Equal(t, ok, false)

	m, ok := ini.GetKvmap("")
	assert.Equal(t, len(m), 7)
	assert.Equal(t, ok, true)

	n, ok := ini.GetKvmap("n")
	assert.Equal(t, len(n), 0)
	assert.Equal(t, ok, false)

	sss, ok := ini.GetKvmap("sss")
	assert.Equal(t, len(sss), 2)
	assert.Equal(t, ok, true)
	v, ok = ini.SectionGet("sss", "aa")
	assert.Equal(t, v, "bb")
	assert.Equal(t, ok, true)
	v, ok = ini.SectionGet("sss", "appext")
	assert.Equal(t, v, "ab=cd")
	assert.Equal(t, ok, true)

	i, ok := ini.SectionGetInt("ddd", "age")
	assert.Equal(t, i, 30)
	assert.Equal(t, ok, true)

	i, ok = ini.SectionGetInt("ddd", "agexxx")
	assert.Equal(t, ok, false)

	f, ok := ini.GetFloat("version")
	assert.Equal(t, f, 4.4)
	assert.Equal(t, ok, true)

	f, ok = ini.SectionGetFloat("ddd", "height")
	assert.Equal(t, f, 175.6)
	assert.Equal(t, ok, true)

	f, ok = ini.SectionGetFloat("ddd", "heightxxx")
	assert.Equal(t, ok, false)

	b, ok := ini.SectionGetBool("ddd", "debug")
	assert.Equal(t, b, true)
	assert.Equal(t, ok, true)

	b, ok = ini.GetBool("debug")
	assert.Equal(t, b, false)
	assert.Equal(t, ok, true)

	//Wirte
	var buf bytes.Buffer
	err = ini.Write(&buf)
	assert.Equal(t, err, nil)

	ini.Reset()
	err = ini.Parse(buf.Bytes(), "\n", "=")
	assert.Equal(t, nil, err)

	v, ok = ini.Get("mid")
	assert.Equal(t, v, "ac9219aa5232c4e519ae5fcb4d77ae5b")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("product")
	assert.Equal(t, v, "ppp")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("combo")
	assert.Equal(t, v, "ccc")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("aa")
	assert.Equal(t, v, "bb")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("axxxa")
	assert.Equal(t, v, "")
	assert.Equal(t, ok, false)

	m, ok = ini.GetKvmap("")
	assert.Equal(t, len(m), 7)
	assert.Equal(t, ok, true)

	n, ok = ini.GetKvmap("n")
	assert.Equal(t, len(n), 0)
	assert.Equal(t, ok, false)

	sss, ok = ini.GetKvmap("sss")
	assert.Equal(t, len(sss), 2)
	assert.Equal(t, ok, true)
	v, ok = ini.SectionGet("sss", "aa")
	assert.Equal(t, v, "bb")
	assert.Equal(t, ok, true)
	v, ok = ini.SectionGet("sss", "appext")
	assert.Equal(t, v, "ab=cd")
	assert.Equal(t, ok, true)

	i, ok = ini.SectionGetInt("ddd", "age")
	assert.Equal(t, i, 30)
	assert.Equal(t, ok, true)

	i, ok = ini.SectionGetInt("ddd", "agexxx")
	assert.Equal(t, ok, false)

	f, ok = ini.GetFloat("version")
	assert.Equal(t, f, 4.4)
	assert.Equal(t, ok, true)

	f, ok = ini.SectionGetFloat("ddd", "height")
	assert.Equal(t, f, 175.6)
	assert.Equal(t, ok, true)

	f, ok = ini.SectionGetFloat("ddd", "heightxxx")
	assert.Equal(t, ok, false)

	b, ok = ini.SectionGetBool("ddd", "debug")
	assert.Equal(t, b, true)
	assert.Equal(t, ok, true)

	b, ok = ini.GetBool("debug")
	assert.Equal(t, b, false)
	assert.Equal(t, ok, true)
}

func Test4(t *testing.T) {
	filename := filepath.Join(getTestDataDir(t), "ini_parser_testfile.ini")
	ini := New()
	err := ini.ParseFile(filename)
	assert.Equal(t, nil, err)

	v, ok := ini.Get("mid")
	assert.Equal(t, v, "ac9219aa5232c4e519ae5fcb4d77ae5b")
	assert.Equal(t, ok, true)

	s := ini.GetAll()
	assert.Equal(t, len(s), 3)
}

func Test5(t *testing.T) {
	filename := filepath.Join(getTestDataDir(t), "ini_parser_testfile.ini")
	f, err := os.Open(filename)
	defer f.Close()
	assert.NotEqual(t, f, nil)
	assert.Equal(t, err, nil)
	ini := New()
	ini.SetParseSection(true)
	err = ini.ParseFrom(f, "\n", "=")
	assert.Equal(t, nil, err)

	v, ok := ini.Get("mid")
	assert.Equal(t, v, "ac9219aa5232c4e519ae5fcb4d77ae5b")
	assert.Equal(t, ok, true)

	s := ini.GetAll()
	assert.Equal(t, len(s), 3)
}

func TestFileNotExist1(t *testing.T) {
	filename := "the/path/to/a/nonexist/ini/file"
	ini := New()
	err := ini.ParseFile(filename)
	assert.NotEqual(t, nil, err)
}

func TestFileNotExist2(t *testing.T) {
	filename := "the/path/to/a/nonexist/ini/file"
	f, _ := os.Open(filename)
	ini := New()
	err := ini.ParseFrom(f, "\n", "=")
	assert.NotEqual(t, nil, err)
	if f != nil {
		f.Close()
	}
}

func TestUft8(t *testing.T) {
	/*
		title=百度搜索_ipad2
		url=http://www.baidu.com/s?bs=ipad&f=8&rsv_bp=1&wd=ipad2&inputT=397
		url_md5=5844a75423cd3372e1997360bd110a25
		refer=http://www.google.com
		anchor_text= google
		ret_form = json
		 ret_start = 0
		 ret_limit =    50
		page_info   =  0,0,50,1,0,20
		local=0
		mid=c4ca4238a0b923820dcc509a6f75849b
		product=test
		combo=test
		version=1.0.0.1
		debug=1
		encoding=1

	*/

	filename := filepath.Join(getTestDataDir(t), "utf8.ini")
	ini := New()
	err := ini.ParseFile(filename)
	assert.Equal(t, nil, err)

	v, ok := ini.Get("title")
	assert.Equal(t, v, "百度搜索_ipad2")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("url_md5")
	assert.Equal(t, v, "5844a75423cd3372e1997360bd110a25")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("ret_form")
	assert.Equal(t, v, "json")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("ret_start")
	assert.Equal(t, v, "0")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("ret_limit")
	assert.Equal(t, v, "50")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("axxxa")
	assert.Equal(t, v, "")
	assert.Equal(t, ok, false)

	m, ok := ini.GetKvmap("")
	assert.Equal(t, len(m), 16)
	assert.Equal(t, ok, true)

	n, ok := ini.GetKvmap("n")
	assert.Equal(t, len(n), 0)
	assert.Equal(t, ok, false)

	i, ok := ini.GetInt("ret_limit")
	assert.Equal(t, i, 50)
	assert.Equal(t, ok, true)
	ini.SetInt("ret_limit", 40)
	i, ok = ini.GetInt("ret_limit")
	assert.Equal(t, i, 40)
	assert.Equal(t, ok, true)

	i, ok = ini.GetInt("debug")
	assert.Equal(t, i, 1)
	assert.Equal(t, ok, true)
}

func TestErrorFormat(t *testing.T) {
	filename := filepath.Join(getTestDataDir(t), "error.ini")
	ini := New()
	err := ini.ParseFile(filename)
	assert.NotEqual(t, nil, err)
}

func TestMemoryData1(t *testing.T) {
	raw := []byte("a:av||b:bv||c:cv||||d:dv||||||")
	ini := New()
	err := ini.Parse(raw, "|", ":")
	assert.Equal(t, err, nil)

	v, ok := ini.Get("a")
	assert.Equal(t, v, "av")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("b")
	assert.Equal(t, v, "bv")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("c")
	assert.Equal(t, v, "cv")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("d")
	assert.Equal(t, v, "dv")
	assert.Equal(t, ok, true)

	m, ok := ini.GetKvmap("")
	assert.Equal(t, len(m), 4)
	assert.Equal(t, ok, true)

	n, ok := ini.GetKvmap("n")
	assert.Equal(t, len(n), 0)
	assert.Equal(t, ok, false)

	var buf bytes.Buffer
	err = ini.Write(&buf)
	assert.Equal(t, err, nil)

	ini.Reset()
	err = ini.Parse(buf.Bytes(), "|", ":")
	assert.Equal(t, err, nil)

	v, ok = ini.Get("a")
	assert.Equal(t, v, "av")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("b")
	assert.Equal(t, v, "bv")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("c")
	assert.Equal(t, v, "cv")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("d")
	assert.Equal(t, v, "dv")
	assert.Equal(t, ok, true)

	m, ok = ini.GetKvmap("")
	assert.Equal(t, len(m), 4)
	assert.Equal(t, ok, true)

	n, ok = ini.GetKvmap("n")
	assert.Equal(t, len(n), 0)
	assert.Equal(t, ok, false)
}

func TestMemoryData2(t *testing.T) {
	raw := []byte("a:av||b:bv||c:cv||||d:dv||||||")
	ini := New()
	err := ini.Parse(raw, "||", ":") // DIFFERENT with TestMemoryData1. use "||" instead of "|"
	assert.Equal(t, nil, err)

	v, ok := ini.Get("a")
	assert.Equal(t, v, "av")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("b")
	assert.Equal(t, v, "bv")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("c")
	assert.Equal(t, v, "cv")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("d")
	assert.Equal(t, v, "dv")
	assert.Equal(t, ok, true)

	m, ok := ini.GetKvmap("")
	assert.Equal(t, len(m), 4)
	assert.Equal(t, ok, true)

	n, ok := ini.GetKvmap("n")
	assert.Equal(t, len(n), 0)
	assert.Equal(t, ok, false)
}

func TestMemoryData3(t *testing.T) {
	raw := []byte("@|@|@|@|@|@|  a:av  @| b : bv @| c:cv  @|@|d:  dv@|@|@|@|@|@|@|")
	ini := New()
	err := ini.Parse(raw, "@|", ":")
	assert.Equal(t, nil, err)

	v, ok := ini.Get("a")
	assert.Equal(t, v, "av")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("b")
	assert.Equal(t, v, "bv")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("c")
	assert.Equal(t, v, "cv")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("d")
	assert.Equal(t, v, "dv")
	assert.Equal(t, ok, true)

	m, ok := ini.GetKvmap("")
	assert.Equal(t, len(m), 4)
	assert.Equal(t, ok, true)

	n, ok := ini.GetKvmap("n")
	assert.Equal(t, len(n), 0)
	assert.Equal(t, ok, false)
}

func TestMemoryData4(t *testing.T) {
	raw := []byte("@|@|@|@|@|@|  a:av  @| b : bv @| c:cv  @|@|d:  dv@|@|@|@|@|@|@|")
	ini := New()
	err := ini.Parse(raw, "@", ":")
	assert.NotEqual(t, nil, err)

	err = ini.Parse(raw, "@|", ":")
	assert.Equal(t, nil, err)
}

func TestGetBool(t *testing.T) {
	raw := []byte("a:1||b:True||c:true||||d:off||e:on||f:false||g:0||||||")
	ini := New()
	err := ini.Parse(raw, "||", ":") // DIFFERENT with TestMemoryData1. use "||" instead of "|"
	assert.Equal(t, nil, err)

	v, ok := ini.GetBool("a")
	assert.Equal(t, v, true)
	assert.Equal(t, ok, true)

	v, ok = ini.GetBool("c")
	assert.Equal(t, v, true)
	assert.Equal(t, ok, true)

	v, ok = ini.GetBool("d")
	assert.Equal(t, v, false)
	assert.Equal(t, ok, true)

	v, ok = ini.GetBool("e")
	assert.Equal(t, v, true)
	assert.Equal(t, ok, true)

	v, ok = ini.GetBool("f")
	assert.Equal(t, v, false)
	assert.Equal(t, ok, true)

	v, ok = ini.GetBool("g")
	assert.Equal(t, v, false)
	assert.Equal(t, ok, true)
}

func TestSetParseSection(t *testing.T) {
	raw := []byte("[xxx]@|@|@|@|@|@|  a:av  @| b : bv @| c:cv  @|@|d:  dv@|@|@|@|@|@|@|")
	ini := New()
	err := ini.Parse(raw, "@", ":")
	assert.NotEqual(t, nil, err)

	err = ini.Parse(raw, "@|", ":")
	assert.NotEqual(t, nil, err)

	ini.SetParseSection(true)
	err = ini.Parse(raw, "@|", ":")
	assert.Equal(t, nil, err)
}

func TestSetSkipCommits(t *testing.T) {
	raw := []byte(";;;@|@|@|@|@|@|  a:av  @| b : bv @| c:cv  @|@|d:  dv@|@|@|@|@|@|@|")
	ini := New()
	err := ini.Parse(raw, "@", ":")
	assert.NotEqual(t, nil, err)

	err = ini.Parse(raw, "@|", ":")
	assert.NotEqual(t, nil, err)

	ini.SetSkipCommits(true)
	err = ini.Parse(raw, "@|", ":")
	assert.Equal(t, nil, err)
}

func TestReset(t *testing.T) {
	raw := []byte("a:1||b:True||c:true||||d:off||e:on||f:false||g:0||||||")
	ini := New()
	err := ini.Parse(raw, "||", ":") // DIFFERENT with TestMemoryData1. use "||" instead of "|"
	assert.Equal(t, nil, err)

	v, ok := ini.GetBool("a")
	assert.Equal(t, v, true)
	assert.Equal(t, ok, true)

	ini.Reset()
	v, ok = ini.GetBool("a")
	assert.Equal(t, ok, false)

	_, ok = ini.GetKvmap("")
	assert.Equal(t, ok, false)
}

// run this by command : go test -test.bench="Benchmark1"
func Benchmark1(b *testing.B) {
	raw := []byte("a:1||b:True||c:true||||d:off||e:on||f:false||g:0||||||")
	ini := New()

	for i := 0; i < b.N; i++ {
		ini.Parse(raw, "||", ":")
	}
}

// run this by command : go test -test.bench="Benchmark2"
func Benchmark2(b *testing.B) {
	raw := []byte("a:1||b:True||c:true||||d:off||e:on||f:false||g:0||||||")
	ini := New()
	ini.Parse(raw, "||", ":")

	for i := 0; i < b.N; i++ {
		v, _ := ini.GetBool("f")
		if v != false {

		}
	}
}

func getTestDataDir(t *testing.T) string {
	var file string
	var ok bool
	_, file, _, ok = runtime.Caller(0)
	assert.Equal(t, ok, true)

	curdir := filepath.Dir(file)
	return filepath.Join(curdir, "examples/data")
}
