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

	"github.com/arduino/arduino-cli/common/formatter"
)

type TestStruct struct {
	Value int `json:"value"`
}

func (ts TestStruct) String() string {
	return fmt.Sprint("VALUE = ", ts.Value)
}

func ExampleJSONFormatter_Format() {
	var example struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
		Field3 struct {
			Inner1 string  `json:"inner1"`
			Inner2 float32 `json:"inner2"`
		} `json:"field3"`
	}

	example.Field1 = "test"
	example.Field2 = 10
	example.Field3.Inner1 = "inner test"
	example.Field3.Inner2 = 10.432412

	var jf formatter.JSONFormatter
	fmt.Println(jf.Format(example))

	var example2 = 3.14
	fmt.Println(jf.Format(example2))

	var example3 float32 = 3.14
	fmt.Println(jf.Format(example3))

	// Output:
	// {"field1":"test","field2":10,"field3":{"inner1":"inner test","inner2":10.432412}} <nil>
	//  float64 ignored
	//  float32 ignored
}

func ExampleJSONFormatter_Format_debug() {
	valid := TestStruct{20}
	invalid := "invalid"
	jf := formatter.JSONFormatter{
		Debug: false,
	}
	// using struct
	fmt.Println(jf.Format(valid))

	// using string (invalid sine it's not a struct or a map)
	fmt.Println(jf.Format(invalid))

	jf.Debug = true
	fmt.Println(jf.Format(valid))
	fmt.Println(jf.Format(invalid))

	// using map
	newValue := make(map[string]int)
	newValue["value2"] = 10

	fmt.Println(jf.Format(newValue))

	// Output:
	// {"value":20} <nil>
	//  string ignored
	// {"value":20} <nil>
	//  string ignored
	// {"value2":10} <nil>
}

func ExampleSetFormatter() {
	formatter.SetFormatter("text")
	fmt.Println(formatter.Format(TestStruct{5}))
	formatter.SetFormatter("json")
	fmt.Println(formatter.Format(TestStruct{10}))
	fmt.Println(formatter.Format(5))

	// Output:
	// VALUE = 5 <nil>
	// {"value":10} <nil>
	//  int ignored
}
