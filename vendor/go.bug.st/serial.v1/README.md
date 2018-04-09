[![Build Status](https://travis-ci.org/bugst/go-serial.svg?branch=v1)](https://travis-ci.org/bugst/go-serial)

# go.bug.st/serial.v1

A cross-platform serial library for go-lang.

## Documentation and examples

See the godoc here: https://godoc.org/go.bug.st/serial.v1

## Development

If you want to contribute to the development of this library, you must clone this git repository directly into your `src` folder under `src/go.bug.st/serial.v1` and checkout the branch `v1`.

```
cd $GOPATH
mkdir -p src/go.bug.st/
git clone https://github.com/bugst/go-serial.git -b v1 src/go.bug.st/serial.v1
go test go.bug.st/serial.v1
```

## What's new in v1

There are some API improvements, in particular object naming is now more idiomatic, class names are less redundant (for example `serial.SerialPort` is now called `serial.Port`), some internal class fields, constants or enumerations are now private and some methods have been moved into the proper interface.

If you come from the version v0 and want to see the full list of API changes, please check this pull request:

https://github.com/bugst/go-serial/pull/5/files

## License

The software is release under a BSD 3-clause license

https://github.com/bugst/go-serial/blob/v1/LICENSE

