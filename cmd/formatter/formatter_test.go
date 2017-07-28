package formatter

import (
	"fmt"
	"testing"
)

type TestStruct struct {
	Value int `json:"value"`
}

func (ts TestStruct) String() string {
	return fmt.Sprint("VALUE = ", ts.Value)
}
func TestFormat(test *testing.T) {
	SetFormatter("text")
	result, err := Format(TestStruct{5})
	if err == nil {
		fmt.Println("RESULT: ", result)
	} else {
		fmt.Sprintln("ERROR: ", err)
		test.Fail()
	}

	SetFormatter("json")
	result, err = Format(TestStruct{10})
	if err == nil {
		fmt.Println("RESULT: ", result)
	} else {
		fmt.Sprintln("ERROR: ", err)
		test.Fail()
	}
	result, err = Format(5)
	if err == nil {
		if result == "" {
			result = "<empty string>"
			fmt.Println("RESULT: ", result)
		} else {
			fmt.Println("unexpected RESULT: ", result)
			test.Fail()
		}
	} else {
		fmt.Sprintln("ERROR: ", err)
	}

	// Output:
	// RESULT: VALUE = 5
	// RESULT: {value:10}
}

func TestPrint(test *testing.T) {
	SetFormatter("text")
	Print(TestStruct{5})

	SetFormatter("json")
	Print(TestStruct{10})

	// Output:
	// RESULT: VALUE = 5
	// RESULT: {value:10}
}

func TestJSONFormatterPrintDebug(test *testing.T) {
	valid := TestStruct{20}
	invalid := "invalid"
	jf := JSONFormatter{
		debug: false,
	}
	//using struct
	jf.Print(valid)

	//using string (invalid sine it's not a struct or a map)
	jf.Print(invalid)

	jf.debug = true
	jf.Print(valid)
	jf.Print(invalid)

	//using map
	newValue := make(map[string]int)
	newValue["value2"] = 10

	jf.Print(newValue)
	// Output:
	// {value:20}
	//
	// {value:20}
	// Only structs and maps values are accepted
	// {value2:10}
}
