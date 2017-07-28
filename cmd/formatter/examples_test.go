package formatter_test

import (
	"fmt"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
)

func ExampleJSONFormatter_Print() {
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
	jf.Print(example)

	var example2 float64 = 3.14
	jf.Print(example2)

	// Output:
	// {"field1":"test","field2":10,"field3":{"inner1":"inner test","inner2":10.432412}}
	// "Only struct values and maps are accepted"
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
	var result string
	result, err := jf.Format(example)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(result)

	var example2 float32 = 3.14
	result, err = jf.Format(example2)
	if err != nil {
		fmt.Print("ERROR: ", err)
	} else {
		fmt.Print("RESULT: ", result)
	}

	// Output:
	// RESULT: {"field1":"test","field2":10,"field3":{"inner1":"inner test","inner2":10.432412}}
	// ERROR: "Only struct values and maps are accepted"
}
