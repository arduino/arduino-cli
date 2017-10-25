# Go Serial

[![GoDoc](https://godoc.org/github.com/mikepb/go-serial?status.svg)](https://godoc.org/github.com/mikepb/go-serial)

Package serial provides a binding to libserialport for serial port
functionality. Serial ports are commonly used with embedded systems,
such as the Arduino platform.

## Usage

```go
package main

import (
  "github.com/mikepb/go-serial"
  "log"
)

func main() {
  options := serial.RawOptions
  options.BitRate = 115200
  p, err := options.Open("/dev/tty")
  if err != nil {
    log.Panic(err)
  }

  defer p.Close()

  buf := make([]byte, 1)
  if c, err := p.Read(buf); err != nil {
    log.Panic(err)
  } else {
    log.Println(buf)
  }
}
```

## Documentation

https://godoc.org/github.com/mikepb/go-serial


## License

    Copyright 2014 Michael Phan-Ba

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.

Files from the libserialport library are licensed under the GNU Lesser General
Public License. These files include in the header the notice that follows.

    This file is part of the libserialport project.

    Copyright (C) 2010-2012 Bert Vermeulen <bert@biot.com>
    Copyright (C) 2010-2012 Uwe Hermann <uwe@hermann-uwe.de>
    Copyright (C) 2013-2014 Martin Ling <martin-libserialport@earth.li>
    Copyright (C) 2013 Matthias Heidbrink <m-sigrok@heidbrink.biz>
    Copyright (C) 2014 Aurelien Jacobs <aurel@gnuage.org>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Lesser General Public License as
    published by the Free Software Foundation, either version 3 of the
    License, or (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU Lesser General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.
