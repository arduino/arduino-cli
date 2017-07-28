package formatter_test

import (
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
