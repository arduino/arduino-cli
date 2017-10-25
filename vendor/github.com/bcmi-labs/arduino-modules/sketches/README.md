use 'godoc cmd/github.com/bcmi-labs/arduino-modules/sketches' for documentation on the github.com/bcmi-labs/arduino-modules/sketches command 

Package sketches
=====================

    import "github.com/bcmi-labs/arduino-modules/sketches"




Functions
---------


```go
func Find(location string) map[string]*Sketch
```

Find traverses the filesystem starting at location searching for arduino
examples. It returns a map of Sketches.

Types
-----


```go
type Deleter interface {
    Delete(id string) error
}
```
Deleter is used to remove sketches from the storage A notfound error is returned
if the sketch wasn't found


```go
type Manager interface {
    Reader
    Searcher
    Writer
    Deleter
}
```
Manager combines the reader and Searcher interface


```go
type Metadata struct {
    CPU struct {
        Fqbn    string `json:"fqbn" gorethink:"fqbn"`
        Name    string `json:"name" gorethink:"name"`
        Port    string `json:"com_name" gorethink:"com_name"`
        Network bool   `json:"network" gorethink:"network"`
    } `json:"cpu" gorethink:"cpu"`
    IncludedLibs []struct {
        Name    string `json:"name" gorethink:"name"`
        Version string `json:"version" gorethink:"version"`
    } `json:"included_libs" gorethink:"included_libs"`
}
```
Metadata is the kind of data associated to a project such as the board


```go
type Mock struct {
    Sketches map[string]Sketch
}
```
Mock is a test helper that allows to persists sketches in memory while
respecting the contract of the sketches interfaces


```go
func (c *Mock) Delete(id string) error
```
Delete removes the sketch from memory


```go
func (c *Mock) Read(id string) (*Sketch, error)
```
Read retrieves a sketch based on the id A notfound error is returned if the
sketch wasn't found


```go
func (c *Mock) Search(fields map[string]string) ([]Sketch, error)
```
Search retrieves from the storage the sketches that match the fields.


```go
func (c *Mock) Write(sketch *Sketch) error
```
Write persists the sketch in memory


```go
type Reader interface {
    Read(id string) (*Sketch, error)
}
```
Reader is used to retrieve sketches from the storage A notfound error is
returned if the sketch wasn't found


```go
type RethinkFS struct {
    DB    r.QueryExecutor
    Table string
    FS    filesystem
}
```
RethinkFS saves the sketches info on rethinkdb and the file data on s3


```go
func (c *RethinkFS) Delete(id string) error
```
Delete removes the sketch both from rethinkdb and the filesystem


```go
func (c *RethinkFS) Read(id string) (*Sketch, error)
```
Read retrieves the sketch from rethinkdb and the file contents from the
filesystem


```go
func (c *RethinkFS) Search(fields map[string]string) ([]Sketch, error)
```
Search retrieves from the storage the sketches that match the fields. It uses
rethinkdb filtering to achieve it. The files in the sketches are without data.


```go
func (c *RethinkFS) Write(sketch *Sketch) error
```
Write persists the sketch on rethinkdb, saving the file contents on the
filesystem


```go
type Searcher interface {
    Search(fields map[string]string) ([]Sketch, error)
}
```
Searcher is used to retrieve from the storage the sketches that match the
fields.


```go
type Sketch struct {
    Name   string    `json:"name" gorethink:"name"`
    Path   string    `json:"path" gorethink:"path"`
    Ino    fs.File   `json:"ino" gorethink:"ino"`
    Files  []fs.File `json:"files,omitempty" gorethink:"files"`
    Folder string    `json:"folder" gorethink:"folder"`

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
}
```
Sketch is the representation of a piece of arduino code


```go
func (s *Sketch) Generate()
```
Generate fills some of the calculated fields, such as Path and ID


```go
func (s *Sketch) GetFile(name string) *fs.File
```
GetFile returns the address of the file with the given name, if it exist


```go
func (s *Sketch) Save(w Writer) error
```
Save is used to persist a sketch on a storage A notvalid error is returned if
the sketch doesn't contain the required fields When you are creating a new
sketch you should provide the Data property for the ino and the files. If you
are updating a previous sketch you can omit the file data for the existing
files. Note that if you omit the file entirely it will be deleted from the
filesystem.

Examples:

	    // will write in the filesystem `sketch1.ino` with data
	    new := Sketch{Name: "sketch1", Owner: "user", Ino: fs.File{Data: data}}
		   new.Save(w)

	    // will write in the filesystem a `file.file` with data
	    files := []fs.File{fs.File{Name: "file.file", Data: data}}
	    addFile: = Sketch{ID: "ABC", Name: "sketch1", Owner: "user", Files: files}
	    addFile.Save(w)

	    // will remove the previously added `file.file`
	    removeFile: = Sketch{ID: "ABC", Name: "sketch1", Owner: "user", Files: []fs.File}
	    removeFile.Save(w)


```go
func (s *Sketch) Validate() error
```
Validate returns nil if the sketch is ready to be saved, or a notvalid error
containing the missing fields


```go
type Writer interface {
    Write(sketch *Sketch) error
}
```
Writer is used to persist sketches on the storage


