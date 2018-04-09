// Package sketches handles the piece of arduino code called sketches, and provides
// a way to represent them in go code and store them on database and filesystem.
package sketches

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/bcmi-labs/arduino-modules/fs"
	"github.com/fluxio/multierror"
	"github.com/juju/errors"
	"github.com/satori/go.uuid"
)

// Reader is used to retrieve sketches from the storage
// A notfound error is returned if the sketch wasn't found
type Reader interface {
	Read(id string) (*Sketch, error)
}

// Searcher is used to retrieve from the storage the sketches that match the fields.
// Skip and limit are used to implement pagination. If limit is 0 all the items are returned
type Searcher interface {
	Search(fields map[string]string, skip, limit int) ([]Sketch, error)
}

// Writer is used to persist sketches on the storage
type Writer interface {
	Write(sketch, old *Sketch) error
}

// Deleter is used to remove sketches from the storage
// A notfound error is returned if the sketch wasn't found
type Deleter interface {
	Delete(id string, fromFSOnly bool) error
}

// ReadWriter combines the reader and Searcher interface
type ReadWriter interface {
	Reader
	Writer
}

// Manager combines the reader and Searcher interface
type Manager interface {
	Reader
	Searcher
	Writer
	Deleter
}

// Sketch is the representation of a piece of arduino code
type Sketch struct {
	Name     string    `json:"name" gorethink:"name"`
	FullPath string    `json:"-" gorethink:"-"`                 //set by Find, not always required
	BasePath string    `json:"base_path" gorethink:"base_path"` //the real base folder (without inner folders), important during deletion
	Path     string    `json:"path" gorethink:"path"`
	Ino      fs.File   `json:"ino" gorethink:"ino"`
	Files    []fs.File `json:"files,omitempty" gorethink:"files"`
	Folder   string    `json:"folder" gorethink:"folder"`

	// Used only by Create examples
	Types []string `json:"types" gorethink:"types"`

	// Used only by Create sketches
	ID        string    `json:"id" gorethink:"id"`
	Owner     string    `json:"owner" gorethink:"owner"`
	Created   time.Time `json:"created" gorethink:"created"`
	Modified  time.Time `json:"timestamp" gorethink:"modified"`
	Private   bool      `json:"private" gorethink:"private"`
	Tutorials []string  `json:"tutorials" gorethink:"tutorials"`
	Metadata  *Metadata `json:"metadata" gorethink:"metadata"`

	// This is not something we need to expose through API
	StorageProvider string `json:"-" gorethink:"storage_provider"`
}

// ImportMetadata imports metadata into the sketch from a sketch.json file in the root
// path of the sketch.
func (s *Sketch) ImportMetadata() error {
	sketchJSON := filepath.Join(s.FullPath, "sketch.json")
	content, err := ioutil.ReadFile(sketchJSON)
	if err != nil {
		if s.Metadata == nil {
			s.Metadata = new(Metadata)
		}
		return errors.Annotate(err, "ImportMetadata")
	}
	var temp Metadata
	err = json.Unmarshal(content, &temp)
	if err != nil {
		if s.Metadata == nil {
			s.Metadata = new(Metadata)
		}
		return errors.Annotate(err, "ImportMetadata")
	}
	s.Metadata = &temp
	return nil
}

// ExportMetadata exports metadata of the sketch into a sketch.json file in the root
// path of the sketch.
func (s Sketch) ExportMetadata() error {
	sketchJSON := filepath.Join(s.FullPath, "sketch.json")
	if s.Metadata == nil {
		return errors.Annotate(errors.New("Cannot export nil metadata"), "ImportMetadata")
	}
	content, err := json.Marshal(s.Metadata)
	if err != nil {
		return errors.Annotate(err, "ImportMetadata")
	}
	err = ioutil.WriteFile(sketchJSON, content, 0666)
	if err != nil {
		return errors.Annotate(err, "ImportMetadata")
	}
	return nil
}

// MetadataCPU contains the info about the board associated to the sketch
type MetadataCPU struct {
	Fqbn    string `json:"fqbn,required" gorethink:"fqbn"`
	Name    string `json:"name,required" gorethink:"name"`
	Port    string `json:"-" gorethink:"com_name"`
	Network bool   `json:"network" gorethink:"network"` // DEPRECATED IN FAVOR OF TYPE
	Type    string `json:"type" gorethink:"type"`
}

// MetadataIncludedLib contains the info of a library used by the sketch
type MetadataIncludedLib struct {
	Name    string `json:"name" gorethink:"name"`
	Version string `json:"version" gorethink:"version"`
}

// Metadata is the kind of data associated to a project such as the board
type Metadata struct {
	CPU          MetadataCPU           `json:"cpu,omitempty" gorethink:"cpu"`
	IncludedLibs []MetadataIncludedLib `json:"included_libs,omitempty" gorethink:"included_libs"`
	Secrets      []MetadataSecret      `json:"secrets,omitempty" gorethink:"secrets"`
}

// MetadataSecret represent a key:value secret value
type MetadataSecret struct {
	Name  string `json:"name" gorethink:"name"`
	Value string `json:"value" gorethink:"value"`
}

// Validate returns nil if the sketch is ready to be saved, or a notvalid error
// containing the missing fields
func (s *Sketch) Validate(old *Sketch) error {
	var errs multierror.Accumulator
	if s.Name == "" {
		errs.Push(errors.New(".Name missing"))
	}
	if s.ID == "" && s.Ino.Data == nil {
		errs.Push(errors.New(".Ino.Data missing"))
	}

	for i := range s.Files {
		if s.Files[i].Data == nil {
			if old == nil {
				errs.Push(errors.New("file" + s.Files[i].Name + " to be created without .Data"))
			}
			oldFile := old.GetFile(s.Files[i].Name)
			if oldFile == nil {
				errs.Push(errors.New("file" + s.Files[i].Name + " to be created without .Data"))
			}
		}
	}

	if errs.Error() == nil {
		return nil
	}
	return errors.NewNotValid(errs.Error(), "sketch is not ready to be saved")
}

// Save is used to persist a sketch on a storage
// A notvalid error is returned if the sketch doesn't contain the required fields
// When you are creating a new sketch you should provide the Data property for the
// ino and the files.
// If you are updating a previous sketch you can omit the file data for the existing files.
// Note that if you omit the file entirely it will be deleted from the filesystem.
//
// Examples:
//     // will write in the filesystem `sketch1.ino` with data
//     new := Sketch{Name: "sketch1", Owner: "user", Ino: fs.File{Data: data}}
// 	   new.Save(w)
//
//     // will write in the filesystem a `file.file` with data
//     files := []fs.File{fs.File{Name: "file.file", Data: data}}
//     addFile: = Sketch{ID: "ABC", Name: "sketch1", Owner: "user", Files: files}
//     addFile.Save(w)
//
//     // will remove the previously added `file.file`
//     removeFile: = Sketch{ID: "ABC", Name: "sketch1", Owner: "user", Files: []fs.File}
//     removeFile.Save(w)
func (s *Sketch) Save(rw ReadWriter) error {
	old, _ := rw.Read(s.ID)

	err := s.Validate(old)
	if err != nil {
		return err
	}

	s.Generate()

	err = rw.Write(s, old)
	if err != nil {
		return err
	}
	return nil
}

// Generate fills some of the calculated fields, such as Path and ID
func (s *Sketch) Generate() {
	if s.ID == "" {
		s.ID = uuid.Must(uuid.NewV4()).String()
	}

	// Ino
	s.Ino.Name = s.Name + ".ino"
	s.Ino.Mime = mime.TypeByExtension(".ino")

	// Path
	hashed := md5.Sum([]byte(s.Owner))
	base := hex.EncodeToString(hashed[:16]) + ":" + s.Owner
	s.BasePath = path.Join(base, "sketches", s.ID)
	s.Path = path.Join(s.BasePath, s.Name)
	s.Ino.Path = path.Join(s.Path, s.Ino.Name)
	for i := range s.Files {
		s.Files[i].Path = path.Join(s.Path, s.Files[i].Name)
	}

	// timestamp
	time0 := time.Time{}
	if s.Created == time0 {
		s.Created = time.Now()
	}
	s.Modified = time.Now()
}

// GetFile returns the address of the file with the given name, if it exist
func (s *Sketch) GetFile(name string) *fs.File {
	if s == nil {
		return nil
	}
	if s.Ino.Name == name {
		return &s.Ino
	}
	for i := range s.Files {
		if s.Files[i].Name == name {
			return &s.Files[i]
		}
	}
	return nil
}

// Merge copies the properties of new into the sketch
func (s *Sketch) Merge(new *Sketch) {
	if new.Name != "" {
		s.Name = new.Name
	}
	if new.Folder != "" {
		s.Folder = new.Folder
	}
	if new.Metadata != nil {
		s.Metadata = new.Metadata
	}
	if new.Tutorials != nil {
		s.Tutorials = new.Tutorials
	}
	if new.Ino.Data != nil {
		s.Ino.Data = new.Ino.Data
	}
	if new.Files != nil {
		s.Files = new.Files
	}
}

// ByID is a sortable collection of sketches
type ByID []Sketch

func (s ByID) Len() int           { return len(s) }
func (s ByID) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByID) Less(i, j int) bool { return s[i].ID < s[j].ID }

// Find traverses the filesystem starting at location searching for arduino
// examples. It returns a map of Sketches.
// EDIT: now can be used to search for any sketch from location, recursively.
func Find(location string, excludeFolders ...string) map[string]*Sketch {
	files := walk(location, excludeFolders...)

	sketches := map[string]*Sketch{}

	for _, path := range files {
		parts := strings.Split(path, string(filepath.Separator))
		// Ignore orphan files
		if len(parts) < 2 {
			continue
		}
		filename := parts[len(parts)-1]
		skname := parts[len(parts)-2]
		folder := filepath.Join(parts[:len(parts)-2]...)
		skpath := filepath.Join(folder, skname)

		sk, ok := sketches[skname]
		if !ok {
			sk = &Sketch{Name: skname, Folder: folder, FullPath: filepath.Join(location, skpath), Path: skpath, Types: []string{"builtin"}}
		}
		if skname+".ino" == filename {
			sk.Ino = fs.File{Name: filename, Path: path}
		} else {
			sk.Files = append(sk.Files, fs.File{Name: filename, Path: path})
		}
		sk.ImportMetadata()
		sketches[skname] = sk

	}

	return sketches
}

// walk returns a list of all the files in location and subfolders
func walk(location string, excludeFolders ...string) []string {
	// Get a list of the files contained in the library
	files := []string{}
	filepath.Walk(location, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			name := filepath.Base(path)
			for _, excludeFolder := range excludeFolders {
				if excludeFolder == name {
					return filepath.SkipDir
				}
			}
			return nil
		}
		path, err = filepath.Rel(location, path)
		if err != nil {
			panic(err)
		}
		if path != "" {
			files = append(files, path)
		}

		return nil
	})
	return files
}

// Implementing common.StoredItem interface
func (s *Sketch) GetID() string              { return s.ID }
func (s *Sketch) GetUser() string            { return s.Owner }
func (s *Sketch) GetStorageProvider() string { return s.StorageProvider }
