/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package formatter_test

import (
	"errors"
	"fmt"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
)

type ExType struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
	Field3 struct {
		Inner1 string  `json:"inner1"`
		Inner2 float32 `json:"inner2"`
	} `json:"field3"`
}

func (et ExType) String() string {
	return fmt.Sprintln("Field1:", et.Field1) +
		fmt.Sprintln("Field2:", et.Field2) +
		fmt.Sprintln("Field3.Inner1:", et.Field3.Inner1) +
		fmt.Sprintln("Field3.Inner2:", et.Field3.Inner2)
}

func ExamplePrint() {
	var example ExType

	example.Field1 = "test"
	example.Field2 = 10
	example.Field3.Inner1 = "inner test"
	example.Field3.Inner2 = 10.432412

	formatter.SetFormatter("json")
	formatter.Print(example)
	fmt.Println() //to separe outputs
	formatter.SetFormatter("text")
	formatter.Print(example)
	// Output:
	// {"field1":"test","field2":10,"field3":{"inner1":"inner test","inner2":10.432412}}
	//
	// Field1: test
	// Field2: 10
	// Field3.Inner1: inner test
	// Field3.Inner2: 10.432412
}

func ExamplePrint_alternative() {
	formatter.SetFormatter("text")
	formatter.Print(TestStruct{5})

	formatter.SetFormatter("json")
	formatter.Print(TestStruct{10})

	// Output:
	// VALUE = 5
	// {"value":10}
}

func ExamplePrintError() {
	formatter.SetFormatter("text")
	formatter.PrintError(errors.New("text error"))
	formatter.SetFormatter("json")
	formatter.PrintError(errors.New("json error"))

	// Output:
	// text error
	// {"error":"json error"}
}
