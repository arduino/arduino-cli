package goini

import (
	"log"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bmizerany/assert"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
}

func TestInheritedINI(t *testing.T) {
	filename := filepath.Join(getTestDataDir(t), "project.ini")
	ini, err := LoadInheritedINI(filename)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, ini)

	v, ok := ini.Get("product")
	assert.Equal(t, v, "test")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("combo")
	assert.Equal(t, v, "test")
	assert.Equal(t, ok, true)

	//Override the inherited value
	v, ok = ini.Get("debug")
	assert.Equal(t, v, "1")
	assert.Equal(t, ok, true)

	//The inherited config
	v, ok = ini.Get("version")
	assert.Equal(t, v, "0.0.0.0")
	assert.Equal(t, ok, true)
	v, ok = ini.Get("encoding")
	assert.Equal(t, v, "0")
	assert.Equal(t, ok, true)

	v, ok = ini.SectionGet("sss", "a")
	assert.Equal(t, v, "aaval")
	assert.Equal(t, ok, true)

	v, ok = ini.SectionGet("sss", "b")
	assert.Equal(t, v, "bval")
	assert.Equal(t, ok, true)

	v, ok = ini.SectionGet("sss", "c")
	assert.Equal(t, v, "ccval")
	assert.Equal(t, ok, true)

	ini = New()
	err = ini.ParseFile(filename)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, ini)

	v, ok = ini.Get("product")
	assert.Equal(t, v, "test")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("combo")
	assert.Equal(t, v, "test")
	assert.Equal(t, ok, true)

	v, ok = ini.Get("version")
	assert.Equal(t, v, "")
	assert.Equal(t, ok, false)
	v, ok = ini.Get("encoding")
	assert.Equal(t, v, "")
	assert.Equal(t, ok, false)

	v, ok = ini.SectionGet("sss", "a")
	assert.Equal(t, v, "aaval")
	assert.Equal(t, ok, true)

	v, ok = ini.SectionGet("sss", "b")
	assert.Equal(t, v, "")
	assert.Equal(t, ok, false)

	v, ok = ini.SectionGet("sss", "c")
	assert.Equal(t, v, "ccval")
	assert.Equal(t, ok, true)
}

func TestInheritedINI2Error(t *testing.T) {
	filename := filepath.Join(getTestDataDir(t), "project2.ini")
	ini, err := LoadInheritedINI(filename)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, ini, (*INI)(nil))
}

func TestInheritedINI3Error(t *testing.T) {
	var filename = "project3_linux.ini"
	if runtime.GOOS == "windows" {
		filename = "project3_windows.ini"
	}
	filename = filepath.Join(getTestDataDir(t), filename)
	ini, err := LoadInheritedINI(filename)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, ini, (*INI)(nil))

	ini = New()
	err = ini.ParseFile(filename)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, ini)
}
