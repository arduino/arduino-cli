Extract
=====================
[![Build Status](https://travis-ci.org/codeclysm/extract.svg?branch=master)](https://travis-ci.org/codeclysm/extract)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/codeclysm/extract/master/LICENSE)
[![Godoc Reference](https://img.shields.io/badge/Godoc-Reference-blue.svg)](https://godoc.org/github.com/codeclysm/extract)


    import "github.com/codeclysm/extract"

Package extract allows to extract archives in zip, tar.gz or tar.bz2 formats
easily.

Most of the time you'll just need to call the proper function with a buffer and
a destination:

```go
data, _ := ioutil.ReadFile("path/to/file.tar.bz2")
buffer := bytes.NewBuffer(data)
extract.TarBz2(data, "/path/where/to/extract", nil)
```

Sometimes you'll want a bit more control over the files, such as extracting a
subfolder of the archive. In this cases you can specify a renamer func that will
change the path for every file:

```go
var shift = func(path string) string {
    parts := strings.Split(path, string(filepath.Separator))
    parts = parts[1:]
    return strings.Join(parts, string(filepath.Separator))
}
extract.TarBz2(data, "/path/where/to/extract", shift)
```

If you don't know which archive you're dealing with (life really is always a surprise) you can use Archive, which will infer the type of archive from the first bytes

```go
extract.Archive(data, "/path/where/to/extract", nil)
```