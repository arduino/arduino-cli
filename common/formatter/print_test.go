/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package formatter_test

import (
	"fmt"

	"github.com/bcmi-labs/arduino-cli/common/formatter"
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
	formatter.SetFormatter("text")
	formatter.Print(example)
	// Output:
	// {"field1":"test","field2":10,"field3":{"inner1":"inner test","inner2":10.432412}}
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
