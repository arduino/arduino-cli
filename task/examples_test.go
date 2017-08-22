package task_test

import (
	"errors"
	"fmt"

	"github.com/bcmi-labs/arduino-cli/task"
)

func ExampleWrapper_Execute() {
	i := 0

	taskToAccomplish := func() task.Result {
		var ret task.Result
		i++
		// if I call this function an odd number of times it triggers an error.
		if i%2 == 0 {
			ret.Error = errors.New("I am triggering an error: I hate odd numbers!!!")
		}
		// if I call this function an even and multiple of 3 number of times it triggers a result.
		if i%3 == 0 && i%2 != 0 {
			ret.Result = "Triggering a result, I love even multiples of 3 and I want to let you know it"
		}
		fmt.Println("I am printing something on the console")
		return ret
	}

	exampleTask := task.Wrapper{
		BeforeMessage: []string{
			"doing something",
			"doing something talking a lot",
			"doing something talking a lot and explaining step by step what I am doing, really",
		},
		AfterMessage: []string{
			"Done",
			"I finished",
			"I accomplished my job in a very accurate way, and without errors",
		},
		ErrorMessage: []string{
			"error happened",
			"I got an error during the execution of the function",
			"I got an error during the execution of the function, which blocked me to continue my job, making me sad",
		},
		Task: taskToAccomplish,
	}

	//crescent level of verbosity.
	for i := 0; i < 5; i++ {
		result := exampleTask.Execute(i)
		if result.Result != nil {
			fmt.Println("RESULT: ", result.Result)
		}
		if result.Error != nil {
			fmt.Println("ERROR: ", result.Error)
		}
		fmt.Println()
	}

	// Output:
	// doing something ...
	// I am printing something on the console
	// Done
	//
	// doing something talking a lot ...
	// I am printing something on the console
	// I got an error during the execution of the function
	// ERROR:  I am triggering an error: I hate odd numbers!!!
	//
	// doing something talking a lot and explaining step by step what I am doing, really ...
	// I am printing something on the console
	// I accomplished my job in a very accurate way, and without errors
	// RESULT:  Triggering a result, I love even multiples of 3 and I want to let you know it
	//
	// doing something talking a lot and explaining step by step what I am doing, really ...
	// I am printing something on the console
	// I got an error during the execution of the function, which blocked me to continue my job, making me sad
	// ERROR:  I am triggering an error: I hate odd numbers!!!
	//
	// doing something talking a lot and explaining step by step what I am doing, really ...
	// I am printing something on the console
	// I accomplished my job in a very accurate way, and without errors
}
