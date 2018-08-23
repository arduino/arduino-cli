//
// Copyright 2014-2017 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

/*
Package enumerator is a golang cross-platform library for USB serial port discovery.

WARNING: this library is still beta-testing code! please consider the library
and the API as *unstable*. Beware that, even if at this point it's unlike to
happen, the API may be subject to change until this notice is removed from
the documentation.

This library has been tested on Linux, Windows and Mac and uses specific OS
services to enumerate USB PID/VID, in particular on MacOSX the use of cgo is
required in order to access the IOKit Framework. This means that the library
cannot be easily cross compiled for GOOS=darwing targets.

*/
package enumerator // import "go.bug.st/serial.v1/enumerator"
