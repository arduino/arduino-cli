## goini 

[![Build Status](https://secure.travis-ci.org/zieckey/goini.png)](http://travis-ci.org/zieckey/goini) [![Coverage Status](https://img.shields.io/coveralls/zieckey/goini.svg)](https://coveralls.io/r/zieckey/goini?branch=master)
         
This is a Go package to interact with arbitrary INI.

Our applications have a lot of memory data which are key/value pairs in the form of various separators NOT only `\n`. 
I have not found any Golang open source code suitable for it. So I write it myself.

`goini` is designed to be simple, flexible, efficient. It

1. Supports the standard INI format
1. Supports section
1. Supports parsing INI file from local disk
1. Supports parsing INI configuration data from memory
1. Supports parsing data which are key/value pairs in the form of various separators NOT only `\n`
1. Supports UTF8 encoding
1. Supports comments which has a leading character `;` or `#`
1. Supports cascading inheritance
1. Only depends standard Golang libraries
1. Has 100% test coverage

## Importing

    import github.com/zieckey/goini

## Usage

### Example 1 : Parses an INI file

The simplest example code is :
```go
import github.com/zieckey/goini

ini := goini.New()
err := ini.ParseFile(filename)
if err != nil {
	fmt.Printf("parse INI file %v failed : %v\n", filename, err.Error())
	return
}

v, ok := ini.Get("the-key")
//...
```

### Example 2 : Parses the memory data with similar format of INI

```go
raw := []byte("a:av||b:bv||c:cv||||d:dv||||||")
ini := goini.New()
err := ini.Parse(raw, "||", ":")
if err != nil {
    fmt.Printf("parse INI memory data failed : %v\n", err.Error())
    return
}

key := "a"
v, ok := ini.Get(key)
if ok {
    fmt.Printf("The value of %v is [%v]\n", key, v) // Output : The value of a is [av]
}

key = "c"
v, ok = ini.Get(key)
if ok {
    fmt.Printf("The value of %v is [%v]\n", key, v) // Output : The value of c is [cv]
}
```

### Example 3 : Parses an inherited INI file

Assume we have a large project which has several production environments.
Each production environment has its own configuration. 
But there is tiny difference between these configurations of the standalone production environments.
So we use a common INI configuration to store the common configurations.
And each production environment inherits from this common INI configuration.

The `common.ini` is bellow:
 
```ini
product=common
combo=common
debug=0

version=0.0.0.0
encoding=0

[sss]
a = aval
b = bval
```

The `project1.ini` is the configuration of project #1 as below:

```ini
inherited_from=common.ini

;the following config will override the values inherited from common.ini
product=project1
combo=test
debug=1

local=0
mid=c4ca4238a0b923820dcc509a6f75849b

[sss]
a = project1-aval
c = project1-cval
```

If we use `goini.LoadInheritedINI("project1.ini")` is the same as we have the following INI configuration:

```ini
product=project1
combo=test
debug=1

local=0
mid=c4ca4238a0b923820dcc509a6f75849b

version=0.0.0.0
encoding=0

[sss]
a = project1-aval
c = project1-cval
```

The value of the key `product` has been overwritten by value `project1`. 
The value of the key `a` in section `sss` has been overwritten by value `project1-aval`.
