package boards_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"testing"

	"github.com/bcmi-labs/arduino-modules/boards"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var save bool

func init() {
	flag.BoolVar(&save, "save", false, "Wether to override the golden file")
	flag.Parse()
}

type ParseTC struct {
	Desc   string
	Path   string
	Golden string
}

func TestParseBoardsTXT(t *testing.T) {
	testCases := []ParseTC{
		{"arduino:avr:mega", "_test/arduino/avr", "_test/parseboardstxt.mega.golden"},
		{"littlebits:avr:w6_arduino", "_test/littlebits/avr", "_test/parseboardstxt.w6_arduino.golden"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s", tc.Desc), func(t *testing.T) {
			list := boards.Boards{}
			list.ParseBoardsTXT(filepath.Join(tc.Path, "boards.txt"))
			board := list[tc.Desc]
			sort.Strings(board.Vid)
			sort.Strings(board.Pid)
			equals(t, board, tc.Golden)
		})
	}
}

func TestParsePlatformTXT(t *testing.T) {
	testCases := []ParseTC{
		{"arduino:avr", "_test/arduino/avr", "_test/parseplatformtxt.arduinoavr.golden"},
		{"littlebits:avr", "_test/littlebits/avr", "_test/parseplatformtxt.littlebits.golden"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s", tc.Desc), func(t *testing.T) {
			plat := boards.Platform{}
			plat.ParsePlatformTXT(filepath.Join(tc.Path, "platform.txt"))
			sort.Stable(plat.Tools)
			equals(t, &plat, tc.Golden)
		})
	}
}

func TestCompute(t *testing.T) {
	testCases := []ParseTC{
		{"arduino:avr:mega", "_test/arduino/avr", "_test/compute.mega.golden"},
		{"arduino:avr:yun", "_test/arduino/avr", "_test/compute.yun.golden"},
		{"arduino:avr:nano", "_test/arduino/avr", "_test/compute.nano.golden"},
		{"arduino:samd:mzero_pro_bl_dbg", "_test/arduino/samd", "_test/compute.mzero_pro_bl_dbg.golden"},
		{"arduino:samd:arduino_zero_edbg", "_test/arduino/samd", "_test/compute.arduino_zero_edbg.golden"},
		{"arduino:samd:mzero_bl", "_test/arduino/samd", "_test/compute.mzero_bl.golden"},
		{"arduino:avr:leonardo", "_test/arduino/avr", "_test/compute.leonardo.golden"},
		{"littlebits:avr:w6_arduino", "_test/littlebits/avr", "_test/compute.w6_arduino.golden"},
		{"microsoft:win10:w10iotcore", "_test/microsoft/win10", "_test/compute.w10iotcore.golden"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s", tc.Desc), func(t *testing.T) {
			brds, _ := boards.Find("_test")

			board := brds[tc.Desc]

			sort.Strings(board.Vid)
			sort.Strings(board.Pid)
			equals(t, board, tc.Golden)
		})
	}
}

func equals(t *testing.T, test interface{}, path string) {
	golden, _ := ioutil.ReadFile(path)
	dump, _ := json.MarshalIndent(test, "", "\t")

	if string(dump) != string(golden) {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(golden), string(dump), false)
		t.Skipf(dmp.DiffPrettyText(diffs))
		if save {
			ioutil.WriteFile(path, []byte(dump), 0777)
		}
	}
}
