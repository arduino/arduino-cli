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
	"fmt"

	"github.com/bcmi-labs/arduino-cli/common/formatter"
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

	var example2 float64 = 3.14
	fmt.Println(jf.Format(example2))

	var example3 float32 = 3.14
	fmt.Println(jf.Format(example3))

	// Output:
	// {"field1":"test","field2":10,"field3":{"inner1":"inner test","inner2":10.432412}} <nil>
	//  float64 ignored
	//  float32 ignored
}

func ExampleJSONFormatter_Print_debug() {
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

func ExampleFormat() {
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
