## paths: a golang library to simplify handling of paths

This library aims to simplify handling of the most common operations with paths.

For example code that looked like this:

```go
buildPath := getPathFromSomewhere() // returns string
if buildPath != "" {
	cachePath, err := filepath.Abs(filepath.Join(buildPath, "cache"))
	...
}
```

can be transformed to:

```go
buildPath := getPathFromSomewhere() // returns *paths.Path
if buildPath != nil {
	cachePath, err := buildPath.Join("cache").Abs()
	...
}
```

most operations that usually requires a bit of convoluted system calls are now simplified, for example to check if a path is a directory:

```go
buildPath := "/path/to/somewhere"
srcPath := filepath.Join(buildPath, "src")
if info, err := os.Stat(srcPath); err == nil && !info.IsDir() {
    os.MkdirAll(srcPath)
}
```

using this library can be done this way:

```go
buildPath := paths.New("/path/to/somewhere")
srcPath := buildPath.Join("src")
if !srcPath.IsDir() {
    scrPath.MkdirAll()
}
```

