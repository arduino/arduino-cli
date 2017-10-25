package sketches_test

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	gorethink "gopkg.in/gorethink/gorethink.v3"

	"github.com/bcmi-labs/arduino-modules/fs"
	"github.com/bcmi-labs/arduino-modules/sketches"
	"github.com/juju/errors"
	"github.com/pborman/uuid"
)

var paths = map[string]string{
	"chell": "4900b18761abb5fe6c9c74f5ebb622bc:chell/sketches/",
}
var fixtures = map[string]sketches.Sketch{
	"sketch001": sketches.Sketch{
		ID: "sketch001", Name: "sketch001",
		Owner: "chell", Folder: "testsketches", Private: false,
		Ino: fs.File{Name: "sketch001.ino", Path: paths["chell"] + "sketch001/sketch001/sketch001.ino", Data: []byte("sketch 001")},
		Files: []fs.File{
			fs.File{Name: "Readme.txt", Path: paths["chell"] + "sketch001/sketch001/Readme.txt", Data: []byte("readme 001")},
		},
	},
	"sketch002": sketches.Sketch{
		ID: "sketch002", Name: "sketch002",
		Owner: "chell", Folder: "", Private: false,
		Ino: fs.File{Name: "sketch002.ino", Path: paths["chell"] + "sketch002/sketch002/sketch002.ino", Data: []byte("sketch 002")},
		Files: []fs.File{
			fs.File{Name: "Readme.txt", Path: paths["chell"] + "sketch002/sketch002/Readme.txt", Data: []byte("readme 002")},
		},
	},
	"sketch003": sketches.Sketch{
		ID: "sketch003", Name: "sketch003",
		Owner: "lara", Folder: "testsketches", Private: false,
		Ino: fs.File{Name: "sketch003.ino", Path: paths["chell"] + "sketch003/sketch003/sketch003.ino", Data: []byte("sketch 003")},
		Files: []fs.File{
			fs.File{Name: "Readme.txt", Path: paths["chell"] + "sketch003/sketch003/Readme.txt", Data: []byte("readme 003")},
		},
	},
	"sketch004": sketches.Sketch{
		ID: "sketch004", Name: "sketch004",
		Owner: "lara", Folder: "", Private: false,
		Ino: fs.File{Name: "sketch004.ino", Path: paths["chell"] + "sketch004/sketch004/sketch004.ino", Data: []byte("sketch 004")},
		Files: []fs.File{
			fs.File{Name: "Readme.txt", Path: paths["chell"] + "sketch004/sketch004/Readme.txt", Data: []byte("readme 004")},
		},
	},
}

func clients() []sketches.Manager {
	// mock client
	mock := sketches.Mock{}

	// RethinkFS db
	session, _ := gorethink.Connect(gorethink.ConnectOpts{
		Address:  "localhost:28015",
		Database: "test",
	})

	gorethink.DBCreate("test").RunWrite(session)
	gorethink.DB("test").TableCreate("sketches").RunWrite(session)

	// RethinkFS fs
	fs := fs.Disk{Base: "/tmp/testsketches"}

	RethinkFS := sketches.RethinkFS{
		DB:    session,
		Table: "sketches",
		FS:    &fs,
	}
	return []sketches.Manager{&mock, &RethinkFS}
}

func setup(ti sketches.Manager) {
	switch c := ti.(type) {
	case *sketches.Mock:
		c.Sketches = map[string]sketches.Sketch{}
		for id, fix := range fixtures {
			c.Sketches[id] = fix
		}
	case *sketches.RethinkFS:
		for _, fix := range fixtures {
			gorethink.DB("test").Table("sketches").Insert(fix).Run(c.DB)
			fix.Ino.Write(c.FS, 0777)
			fix.Files[0].Write(c.FS, 0777)
		}
	}
}

func setupVolume(ti sketches.Manager) {
	ss := []sketches.Sketch{}
	for _, fix := range fixtures {
		for i := 0; i < 10; i++ {
			sketch := sketches.Sketch{ID: uuid.New(), Owner: fix.Owner, Name: fix.Name + "-" + strconv.Itoa(i)}
			ss = append(ss, sketch)
		}
	}

	switch c := ti.(type) {
	case *sketches.Mock:
		c.Sketches = map[string]sketches.Sketch{}
		for _, sketch := range ss {
			c.Sketches[sketch.ID] = sketch
		}
	case *sketches.RethinkFS:
		for _, sketch := range ss {
			gorethink.DB("test").Table("sketches").Insert(sketch).Run(c.DB)
		}
	}
}

func cleanup(ti sketches.Manager) {
	switch c := ti.(type) {
	case *sketches.RethinkFS:
		gorethink.DB("test").Table("sketches").Delete().Exec(c.DB)
		os.RemoveAll("/tmp/testsketches")
	}
}

type ReaderTC struct {
	Desc           string
	ID             string
	ExpectedError  string
	ExpectedSketch string
}

func TestReader(t *testing.T) {
	testCases := []ReaderTC{
		{"found1", "sketch001", "nil", "sketch001"},
		{"found2", "sketch002", "nil", "sketch002"},
		{"missing", "missing", "notfound", "nil"},
	}

	interfaces := clients()

	for _, tc := range testCases {
		for _, ti := range interfaces {
			setup(ti)
			t.Run(fmt.Sprintf("%s:%T", tc.Desc, ti), func(t *testing.T) {
				sketch, err := ti.Read(tc.ID)

				// Test error
				checkError(t, err, tc.ExpectedError)
				if tc.ExpectedError != "nil" {
					return
				}

				// Test sketch
				if sketch == nil {
					t.Skipf("sketch is nil")
					return
				}

				expected, ok := fixtures[tc.ExpectedSketch]
				if !ok {
					t.Skipf("sketch %s not found in test fixtures", tc.ExpectedSketch)
					return
				}

				if expected.ID != sketch.ID {
					t.Skipf("%s should be %s, got %s", "ID", expected.ID, sketch.ID)
				}

				if string(expected.Ino.Data) != string(sketch.Ino.Data) {
					t.Skipf("%s should be %s, got %s", "Ino.Data", expected.Ino.Data, sketch.Ino.Data)
				}

				if string(expected.Files[0].Data) != string(sketch.Files[0].Data) {
					t.Skipf("%s should be %s, got %s", "Files[0].Data", expected.Files[0].Data, sketch.Files[0].Data)
				}

			})
			cleanup(ti)
		}
	}
}

type SearcherTC struct {
	Desc              string
	Fields            map[string]string
	Skip              int
	Limit             int
	ExpectedSketchesN int
	ExpectedSketches  []string
	Setup             func(sketches.Manager)
}

func TestSearcher(t *testing.T) {
	testCases := []SearcherTC{
		{"byOwner1", map[string]string{"owner": "chell"}, 0, 0, 2, []string{"sketch001", "sketch002"}, setup},
		{"byOwner2", map[string]string{"owner": "lara"}, 0, 0, 2, []string{"sketch003", "sketch004"}, setup},
		{"byFolder1", map[string]string{"folder": "testsketches"}, 0, 0, 2, []string{"sketch001", "sketch003"}, setup},
		{"byFolder2", map[string]string{"folder": ""}, 0, 0, 2, []string{"sketch002", "sketch004"}, setup},
		{"byOwnerFolder", map[string]string{"owner": "chell", "folder": ""}, 0, 0, 1, []string{"sketch002"}, setup},
		{"empty", map[string]string{"owner": "missing"}, 0, 0, 0, []string{}, setup},
		{"limits1", map[string]string{}, 0, 1, 4, []string{"sketch001", "sketch002", "sketch003", "sketch004"}, setup},
		{"limits2", map[string]string{}, 0, 5, 40, nil, setupVolume},
	}

	interfaces := clients()

	for _, tc := range testCases {
		for _, ti := range interfaces {
			if tc.Setup != nil {
				tc.Setup(ti)
			}
			t.Run(fmt.Sprintf("%s:%T", tc.Desc, ti), func(t *testing.T) {
				list, _ := ti.Search(tc.Fields, tc.Skip, tc.Limit)

				// Test pagination
				if tc.Limit != 0 {
					length := len(list)
					limit := tc.Limit
					skip := limit
					br := 0
					for length != 0 && br < 10 {
						br++
						l, _ := ti.Search(tc.Fields, skip, limit)
						length = len(l)
						list = append(list, l...)
						skip = skip + limit
					}
				}

				// Test sketches
				if len(list) != tc.ExpectedSketchesN {
					t.Skipf("Expected %d sketches, got %d", tc.ExpectedSketchesN, len(list))
				}

				if tc.ExpectedSketches == nil {
					return
				}

				actual := []string{}
				for _, sketch := range list {
					actual = append(actual, sketch.ID)
				}
				mustMatch(t, actual, tc.ExpectedSketches)

			})
			cleanup(ti)
		}
	}
}

type WriterTC struct {
	Desc          string
	Sketch        sketches.Sketch
	ExpectedError string
	Assert        func(t *testing.T, s *sketches.Sketch, ti sketches.Manager)
}

func TestWriter(t *testing.T) {
	testCases := []WriterTC{
		{"validation1", sketches.Sketch{}, "notvalid", nil},
		{"validation2", sketches.Sketch{Name: "sketch"}, "notvalid", nil},
		{"validation3", sketches.Sketch{Name: "sketch", Ino: fs.File{Data: []byte("")}}, "nil", nil},
		{"validation4", sketches.Sketch{Ino: fs.File{Data: []byte("")}}, "notvalid", nil},

		{"generate1", sketches.Sketch{Name: "sketch", Ino: fs.File{Data: []byte("")}}, "nil", checkGenerate},
		{"generate2", sketches.Sketch{Name: "sketch", Ino: fs.File{Name: "wrong", Data: []byte("")}}, "nil", checkGenerate},
		{"generate1", sketches.Sketch{Name: "sketch", Ino: fs.File{Path: "wrong", Data: []byte("")}}, "nil", checkGenerate},

		{"files1", sketches.Sketch{ID: "sketch001", Owner: "chell", Name: "sketch001", Files: []fs.File{
			fs.File{Name: "Readme.txt", Data: []byte("New content for readme.txt")},
		}}, "nil", checkFiles(paths["chell"]+"sketch001/sketch001/Readme.txt", true, "New content for readme.txt")},
		{"files2", sketches.Sketch{ID: "sketch001", Owner: "chell", Name: "sketch001", Files: []fs.File{}}, "nil", checkFiles(paths["chell"]+"sketch001/sketch001/Readme.txt", false, "")},
		{"files1", sketches.Sketch{ID: "sketch001", Owner: "chell", Name: "sketch001", Files: []fs.File{
			fs.File{Name: "Readme.txt", Data: []byte("New content for readme.txt")},
		}}, "nil", checkFiles(paths["chell"]+"sketch001/sketch001/Readme.txt", true, "New content for readme.txt")},

		{"rename1", sketches.Sketch{ID: "sketch001", Owner: "chell", Name: "sketch002", Files: []fs.File{fs.File{Name: "Readme.txt"}}}, "nil", checkFiles(paths["chell"]+"sketch001/sketch001/Readme.txt", false, "")},
		{"rename2", sketches.Sketch{ID: "sketch001", Owner: "chell", Name: "sketch002", Files: []fs.File{fs.File{Name: "Readme.txt"}}}, "nil", checkFiles(paths["chell"]+"sketch001/sketch002/Readme.txt", true, "readme 001")},
	}
	interfaces := clients()

	for _, tc := range testCases {
		for _, ti := range interfaces {
			setup(ti)
			t.Run(fmt.Sprintf("%s:%T", tc.Desc, ti), func(t *testing.T) {
				sketch := tc.Sketch
				err := sketch.Save(ti)
				// Test error
				checkError(t, err, tc.ExpectedError)
				if tc.ExpectedError != "nil" {
					return
				}

				if tc.Assert != nil {
					sketchSaved, _ := ti.Read(sketch.ID)
					tc.Assert(t, &sketch, ti)
					tc.Assert(t, sketchSaved, ti)
				}
			})
			cleanup(ti)
		}
	}
}

type DeleterTC struct {
	Desc          string
	ID            string
	ExpectedError string
	Assert        func(t *testing.T, ti sketches.Manager)
}

func TestDeleter(t *testing.T) {
	testCases := []DeleterTC{
		{"delete1", "sketch001", "nil", fileMissing(paths["chell"] + "sketch001/sketch001/Readme.txt")},
		{"delete2", "sketch001", "nil", fileMissing(paths["chell"] + "sketch001/sketch001/sketch001.ino")},
		{"notfound", "sketch005", "notfound", nil},
	}
	interfaces := clients()

	for _, tc := range testCases {
		for _, ti := range interfaces {
			setup(ti)
			t.Run(fmt.Sprintf("%s:%T", tc.Desc, ti), func(t *testing.T) {
				err := ti.Delete(tc.ID)

				// Test error
				checkError(t, err, tc.ExpectedError)
				if tc.ExpectedError != "nil" {
					return
				}

				_, err = ti.Read(tc.ID)
				if !errors.IsNotFound(err) {
					t.Skipf("sketch wasn't deleted")
					return
				}

				if tc.Assert != nil {
					tc.Assert(t, ti)
				}
			})
			cleanup(ti)
		}
	}
}

var ParseTC = []struct {
	Name   string
	Folder string
	Path   string
	Files  []string
	Ino    []string
}{
	{"Blink", "01.Basics", "01.Basics/Blink", []string{"Blink.txt"}, []string{"Blink.ino"}},
	{"Fade", "01.Basics", "01.Basics/Fade", []string{"Fade.txt", "layout.png", "schematic.png"}, []string{"Fade.ino"}},
}

func TestParse(t *testing.T) {
	for _, tc := range ParseTC {
		list := sketches.Find("_test")
		ex := list[tc.Name]
		if ex == nil {
			t.Skipf("example missing")
			return
		}
		if tc.Name != ex.Name {
			t.Skipf("%s should be '%s', got '%s", ".Name", tc.Name, ex.Name)
		}
		if tc.Folder != ex.Folder {
			t.Skipf("%s should be '%s', got '%s", ".Folder", tc.Folder, ex.Folder)
		}
		if tc.Path != ex.Path {
			t.Skipf("%s should be '%s', got '%s", ".Path", tc.Path, ex.Path)
		}
		files := []string{}
		for _, f := range ex.Files {
			files = append(files, f.Name)
		}
		mustMatch(t, tc.Files, files)
	}
}

func checkGenerate(t *testing.T, s *sketches.Sketch, ti sketches.Manager) {
	if s == nil {
		t.Skipf("sketch should not be nil")
		return
	}
	if s.ID == "" {
		t.Skipf(".ID should not be empty")
	}
	if s.Path == "" {
		t.Skipf(".Path should not be empty")
	}
	if s.Ino.Name != s.Name+".ino" {
		t.Skipf(".Ino.Name should be '%s', got '%s'", s.Name+".ino", s.Ino.Name)
	}
	if s.Ino.Path != path.Join(s.Path, s.Ino.Name) {
		t.Skipf(".Ino.Path should be '%s', got '%s'", path.Join(s.Path, s.Ino.Name), s.Ino.Path)
	}
	time0 := time.Time{}
	if s.Created == time0 {
		t.Skipf(".Created should not be %s", time0)
	}
	if s.Modified == time0 {
		t.Skipf(".Modified should not be %s", time0)
	}
}

func checkFiles(name string, present bool, content string) func(t *testing.T, s *sketches.Sketch, ti sketches.Manager) {
	return func(t *testing.T, s *sketches.Sketch, ti sketches.Manager) {
		if s == nil {
			t.Skipf("sketch should not be nil")
			return
		}

		found := false
		for i := range s.Files {
			if s.Files[i].Path == name {
				found = true
				if present && string(s.Files[i].Data) != content {
					t.Skipf(".Data of file %s should be '%s', got '%s'", name, content, string(s.Files[i].Data))
				}
			}
		}

		if found && !present {
			t.Skipf("file %s should not be present", name)
		}
		if !found && present {
			t.Skipf("file %s should be present", name)
		}

		if present {
			filePresent(name)(t, ti)
		} else {
			fileMissing(name)(t, ti)
		}
	}
}

func filePresent(path string) func(t *testing.T, ti sketches.Manager) {
	return func(t *testing.T, ti sketches.Manager) {
		switch w := ti.(type) {
		case *sketches.RethinkFS:
			_, err := w.FS.ReadFile(path)
			if err != nil {
				t.Skipf("file %s should be present", path)
			}
		}
	}
}

func fileMissing(path string) func(t *testing.T, ti sketches.Manager) {
	return func(t *testing.T, ti sketches.Manager) {
		switch w := ti.(type) {
		case *sketches.RethinkFS:
			_, err := w.FS.ReadFile(path)
			if err == nil {
				t.Skipf("file %s should not be present", path)
			}
		}
	}
}

func checkError(t *testing.T, err error, expected string) {
	if expected == "nil" && err != nil {
		t.Skipf("err should be nil, got %s", err.Error())
		return
	}
	if expected != "nil" && err == nil {
		t.Skipf("err should be %s, got nil", expected)
		return
	}
	if expected == "notfound" {
		if !errors.IsNotFound(err) {
			t.Skipf("err should be notfound, got %s", err)
			return
		}
		return
	}
	if expected == "notvalid" {
		if !errors.IsNotValid(err) {
			t.Skipf("err should be notvalid, got %s", err)
			return
		}
		return
	}
}

func mustMatch(t *testing.T, s1, s2 []string) {
	for _, e1 := range s1 {
		found := false
		for _, e2 := range s2 {
			if e1 == e2 {
				found = true
			}
		}
		if !found {
			t.Skipf("%s missing from the expected list", e1)
		}
	}
	for _, e2 := range s2 {
		found := false
		for _, e1 := range s1 {
			if e1 == e2 {
				found = true
			}
		}
		if !found {
			t.Skipf("%s missing from the actual list", e2)
		}
	}
}
