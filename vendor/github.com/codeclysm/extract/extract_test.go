package extract_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/codeclysm/extract"
)

type Files map[string]string

var shift = func(path string) string {
	parts := strings.Split(path, string(filepath.Separator))
	parts = parts[1:]
	return strings.Join(parts, string(filepath.Separator))
}

var subfolder = func(path string) string {
	if strings.Contains(path, "archive/folder") {
		return path
	}
	return ""
}

var ExtractCases = []struct {
	Name    string
	Archive string
	Renamer extract.Renamer
	Files   Files
}{
	{"standard bz2", "testdata/archive.tar.bz2", nil, Files{
		"":                          "dir",
		"/archive":                  "dir",
		"/archive/folder":           "dir",
		"/archive/folderlink":       "link",
		"/archive/folder/file1.txt": "folder/File1",
		"/archive/file1.txt":        "File1",
		"/archive/file2.txt":        "File2",
		"/archive/link.txt":         "File1",
	}},
	{"shift bz2", "testdata/archive.tar.bz2", shift, Files{
		"":                  "dir",
		"/folder":           "dir",
		"/folderlink":       "link",
		"/folder/file1.txt": "folder/File1",
		"/file1.txt":        "File1",
		"/file2.txt":        "File2",
		"/link.txt":         "File1",
	}},
	{"subfolder bz2", "testdata/archive.tar.bz2", subfolder, Files{
		"":                          "dir",
		"/archive":                  "dir",
		"/archive/folder":           "dir",
		"/archive/folder/file1.txt": "folder/File1",
		"/archive/folderlink":       "link",
	}},
	{"not tarred bz2", "testdata/singlefile.bz2", nil, Files{
		"": "singlefile",
	}},

	{"standard gz", "testdata/archive.tar.gz", nil, Files{
		"":                          "dir",
		"/archive":                  "dir",
		"/archive/folder":           "dir",
		"/archive/folderlink":       "link",
		"/archive/folder/file1.txt": "folder/File1",
		"/archive/file1.txt":        "File1",
		"/archive/file2.txt":        "File2",
		"/archive/link.txt":         "File1",
	}},
	{"shift gz", "testdata/archive.tar.gz", shift, Files{
		"":                  "dir",
		"/folder":           "dir",
		"/folderlink":       "link",
		"/folder/file1.txt": "folder/File1",
		"/file1.txt":        "File1",
		"/file2.txt":        "File2",
		"/link.txt":         "File1",
	}},
	{"subfolder gz", "testdata/archive.tar.gz", subfolder, Files{
		"":                          "dir",
		"/archive":                  "dir",
		"/archive/folder":           "dir",
		"/archive/folder/file1.txt": "folder/File1",
		"/archive/folderlink":       "link",
	}},
	{"not tarred gz", "testdata/singlefile.gz", nil, Files{
		"": "singlefile",
	}},
	// Note that the zip format doesn't support hard links
	{"standard zip", "testdata/archive.zip", nil, Files{
		"":                          "dir",
		"/archive":                  "dir",
		"/archive/folder":           "dir",
		"/archive/folderlink":       "link",
		"/archive/folder/file1.txt": "folder/File1",
		"/archive/file1.txt":        "File1",
		"/archive/file2.txt":        "File2",
		"/archive/link.txt":         "File1",
	}},
	{"shift zip", "testdata/archive.zip", shift, Files{
		"":                  "dir",
		"/folder":           "dir",
		"/folderlink":       "link",
		"/folder/file1.txt": "folder/File1",
		"/file1.txt":        "File1",
		"/file2.txt":        "File2",
		"/link.txt":         "File1",
	}},
	{"subfolder zip", "testdata/archive.zip", subfolder, Files{
		"":                          "dir",
		"/archive":                  "dir",
		"/archive/folder":           "dir",
		"/archive/folder/file1.txt": "folder/File1",
		"/archive/folderlink":       "link",
	}},

	{"standard inferred", "testdata/archive.mistery", nil, Files{
		"":                          "dir",
		"/archive":                  "dir",
		"/archive/folder":           "dir",
		"/archive/folderlink":       "link",
		"/archive/folder/file1.txt": "folder/File1",
		"/archive/file1.txt":        "File1",
		"/archive/file2.txt":        "File2",
		"/archive/link.txt":         "File1",
	}},
	{"shift inferred", "testdata/archive.mistery", shift, Files{
		"":                  "dir",
		"/folder":           "dir",
		"/folderlink":       "link",
		"/folder/file1.txt": "folder/File1",
		"/file1.txt":        "File1",
		"/file2.txt":        "File2",
		"/link.txt":         "File1",
	}},
	{"subfolder inferred", "testdata/archive.mistery", subfolder, Files{
		"":                          "dir",
		"/archive":                  "dir",
		"/archive/folder":           "dir",
		"/archive/folder/file1.txt": "folder/File1",
		"/archive/folderlink":       "link",
	}},
}

func TestArchiveFailure(t *testing.T) {
	err := extract.Archive(strings.NewReader("not an archive"), "", nil)
	if err == nil || err.Error() != "Not a supported archive" {
		t.Error("Expected error 'Not a supported archive', got", err)
	}
}

func TestExtract(t *testing.T) {
	for _, test := range ExtractCases {
		dir, _ := ioutil.TempDir("", "")
		dir = filepath.Join(dir, "test")
		data, err := ioutil.ReadFile(test.Archive)
		if err != nil {
			t.Fatal(err)
		}
		buffer := bytes.NewBuffer(data)

		switch filepath.Ext(test.Archive) {
		case ".bz2":
			err = extract.Bz2(buffer, dir, test.Renamer)
		case ".gz":
			err = extract.Gz(buffer, dir, test.Renamer)
		case ".zip":
			err = extract.Zip(buffer, dir, test.Renamer)
		case ".mistery":
			err = extract.Archive(buffer, dir, test.Renamer)
		default:
			t.Fatal("unknown error")
		}

		if err != nil {
			t.Fatal(test.Name, ": Should not fail: "+err.Error())
		}

		files := Files{}

		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			path = strings.Replace(path, dir, "", 1)

			if info.IsDir() {
				files[path] = "dir"
			} else if info.Mode()&os.ModeSymlink != 0 {
				files[path] = "link"
			} else {
				data, err := ioutil.ReadFile(filepath.Join(dir, path))
				if err != nil {

				}
				files[path] = strings.TrimSpace(string(data))
			}

			return nil
		})

		for file, kind := range files {
			k, ok := test.Files[file]
			if !ok {
				t.Error(test.Name, ": "+file+" should not exist")
				continue
			}

			if kind != k {
				t.Error(test.Name, ": "+file+" should be "+k+", not "+kind)
				continue
			}
		}

		for file, kind := range test.Files {
			k, ok := files[file]
			if !ok {
				t.Error(test.Name, ": "+file+" should exist")
				continue
			}

			if kind != k {
				t.Error(test.Name, ": "+file+" should be "+kind+", not "+k)
				continue
			}
		}

		err = os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func BenchmarkArchive(b *testing.B) {
	dir, _ := ioutil.TempDir("", "")
	data, _ := ioutil.ReadFile("testdata/archive.tar.bz2")

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer(data)
		err := extract.Archive(buffer, filepath.Join(dir, strconv.Itoa(i)), nil)
		if err != nil {
			b.Error(err)
		}
	}

	b.StopTimer()

	err := os.RemoveAll(dir)
	if err != nil {
		b.Error(err)
	}
}

func BenchmarkTarBz2(b *testing.B) {
	dir, _ := ioutil.TempDir("", "")
	data, _ := ioutil.ReadFile("testdata/archive.tar.bz2")

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer(data)
		err := extract.Bz2(buffer, filepath.Join(dir, strconv.Itoa(i)), nil)
		if err != nil {
			b.Error(err)
		}
	}

	b.StopTimer()

	err := os.RemoveAll(dir)
	if err != nil {
		b.Error(err)
	}
}

func BenchmarkTarGz(b *testing.B) {
	dir, _ := ioutil.TempDir("", "")
	data, _ := ioutil.ReadFile("testdata/archive.tar.gz")

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer(data)
		err := extract.Gz(buffer, filepath.Join(dir, strconv.Itoa(i)), nil)
		if err != nil {
			b.Error(err)
		}
	}

	b.StopTimer()

	err := os.RemoveAll(dir)
	if err != nil {
		b.Error(err)
	}
}

func BenchmarkZip(b *testing.B) {
	dir, _ := ioutil.TempDir("", "")
	data, _ := ioutil.ReadFile("testdata/archive.zip")

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		buffer := bytes.NewBuffer(data)
		err := extract.Zip(buffer, filepath.Join(dir, strconv.Itoa(i)), nil)
		if err != nil {
			b.Error(err)
		}
	}

	b.StopTimer()

	err := os.RemoveAll(dir)
	if err != nil {
		b.Error(err)
	}
}
