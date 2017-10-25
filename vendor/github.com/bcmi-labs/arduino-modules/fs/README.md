# fs
--
    import "github.com/bcmi-labs/arduino-modules/fs"

Package fs is an abstraction of the filesystem. In our apps we often have to
perform operations both on the filesystem and on s3, and we need this operation
to be completely transparent to the library. Also it doesn't hurt to have
something ready to be marshalled into json or rethinkdb.

## Usage

#### type Disk

```go
type Disk struct {
}
```

Disk is a wrapper around the usual os and ioutil golang libraries

#### func (*Disk) ReadFile

```go
func (d *Disk) ReadFile(filename string) ([]byte, error)
```
ReadFile reads the file named by filename and returns the contents. A successful
call returns err == nil, not err == EOF. Because ReadFile reads the whole file,
it does not treat an EOF from Read as an error to be reported.

#### func (*Disk) Remove

```go
func (d *Disk) Remove(name string) error
```
Remove removes the named file or directory. If there is an error, it will be of
type *PathError.

#### func (*Disk) WriteFile

```go
func (d *Disk) WriteFile(filename string, data []byte, perm os.FileMode) error
```
WriteFile writes data to a file named by filename. If the file does not exist,
WriteFile creates it with permissions perm; otherwise WriteFile truncates it
before writing.

#### type File

```go
type File struct {
	Path         string    `json:"-" gorethink:"path"`
	Data         []byte    `json:"data,omitempty" gorethink:"-"`
	Name         string    `json:"name,omitempty" gorethink:"name"`
	Mime         string    `json:"mimetype,omitempty" gorethink:"mime"`
	LastModified time.Time `json:"last_modified" gorethink:"last_modified"`
}
```

File represents a single file in the filesystem. It is ready to be marshalled as
json or as a rethinkdb document. As json the path will be hidden As rethinkdb
document the data will be hidden.

#### func (*File) Read

```go
func (f *File) Read(fs Reader) error
```
Read populates the Data property of the file by using the fs param

#### func (*File) Write

```go
func (f *File) Write(fs Writer, perm os.FileMode) error
```
Write persists the data property on the fs, with the permission given

#### type Reader

```go
type Reader interface {
	ReadFile(filename string) ([]byte, error)
}
```

Reader allows to read the content of a file on a filesystem

#### type S3

```go
type S3 struct {
	Bucket  string
	Service s3iface.S3API
}
```

S3 is a filesystem that uses Amazon's S3 to persist files and directories

#### func (*S3) ReadFile

```go
func (s *S3) ReadFile(filename string) ([]byte, error)
```
ReadFile reads the file named by filename and returns the contents.

#### func (*S3) Remove

```go
func (s *S3) Remove(name string) error
```
Remove removes the named file or directory. If there is an error, it will be of
type *PathError.

#### func (*S3) WriteFile

```go
func (s *S3) WriteFile(filename string, data []byte, perm os.FileMode) error
```
WriteFile writes data to a file named by filename. If the file does not exist,
WriteFile creates it with permissions perm; otherwise WriteFile truncates it
before writing. It ignores the perm field

#### type Writer

```go
type Writer interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
}
```

Writer allows to write the content of a file on a filesystem
