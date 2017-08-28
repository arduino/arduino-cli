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

func ExampleCreateSequence() {
	verbosity := 1

	tasks := []task.Wrapper{
		task.Wrapper{
			BeforeMessage: []string{
				"starting first task",
			},
			AfterMessage: []string{
				"first task over",
			},
			Task: func() task.Result {
				fmt.Println("First task")
				return task.Result{}
			},
		}, task.Wrapper{
			BeforeMessage: []string{
				"starting second task",
			},
			AfterMessage: []string{
				"second task over",
			},
			Task: func() task.Result {
				fmt.Println("second task")
				return task.Result{}
			},
		},
		task.Wrapper{
			BeforeMessage: []string{
				"starting third task",
			},
			AfterMessage: []string{
				"third task over",
			},
			Task: func() task.Result {
				fmt.Println("third task")
				return task.Result{}
			},
		},
	}

	sequence := task.Wrapper{
		BeforeMessage: []string{
			"Starting sequence",
		},
		AfterMessage: []string{
			"Sequence over",
		},
		Task: task.CreateSequence(tasks, []bool{true, true, true}, verbosity).Task(),
	}

	sequence.Execute(verbosity)

	// Output:
	// Starting sequence ...
	// starting first task ...
	// First task
	// first task over
	// starting second task ...
	// second task
	// second task over
	// starting third task ...
	// third task
	// third task over
	// Sequence over
}

func ExampleCreateSequence_errors_ignored() {
	verbosity := 1

	tasks := []task.Wrapper{
		task.Wrapper{
			BeforeMessage: []string{
				"starting first task",
			},
			AfterMessage: []string{
				"first task over",
			},
			ErrorMessage: []string{
				"first task with error",
			},
			Task: func() task.Result {
				fmt.Println("First task")
				return task.Result{}
			},
		}, task.Wrapper{
			BeforeMessage: []string{
				"starting second task",
			},
			AfterMessage: []string{
				"second task over",
			},
			ErrorMessage: []string{
				"second task with error",
			},
			Task: func() task.Result {
				fmt.Println("second task (with error)")
				return task.Result{Error: errors.New("Error Triggered")}
			},
		},
		task.Wrapper{
			BeforeMessage: []string{
				"starting third task",
			},
			AfterMessage: []string{
				"third task over",
			},
			ErrorMessage: []string{
				"third task with error",
			},
			Task: func() task.Result {
				fmt.Println("third task (with error)")
				return task.Result{Error: errors.New("Second Error Triggered")}
			},
		},
		task.Wrapper{
			BeforeMessage: []string{
				"starting fourth task",
			},
			AfterMessage: []string{
				"fourth task over",
			},
			ErrorMessage: []string{
				"fourth task with error",
			},
			Task: func() task.Result {
				fmt.Println("fourth task")
				return task.Result{}
			},
		},
	}

	sequence := task.Wrapper{
		BeforeMessage: []string{
			"Starting sequence",
		},
		AfterMessage: []string{
			"Sequence over",
		},
		ErrorMessage: []string{
			"Sequence with errors",
		},
		// Do not ignore errors for every step.
		Task: task.CreateSequence(tasks, []bool{true, true, true, true}, verbosity).Task(),
	}

	sequence.Execute(verbosity)

	// Output:
	// Starting sequence ...
	// starting first task ...
	// First task
	// first task over
	// starting second task ...
	// second task (with error)
	// second task with error
	// starting third task ...
	// third task (with error)
	// third task with error
	// starting fourth task ...
	// fourth task
	// fourth task over
	// Sequence over
}

func ExampleCreateSequence_errors_shown_all() {
	verbosity := 1

	tasks := []task.Wrapper{
		task.Wrapper{
			BeforeMessage: []string{
				"starting first task",
			},
			AfterMessage: []string{
				"first task over",
			},
			ErrorMessage: []string{
				"first task with error",
			},
			Task: func() task.Result {
				fmt.Println("First task")
				return task.Result{}
			},
		}, task.Wrapper{
			BeforeMessage: []string{
				"starting second task",
			},
			AfterMessage: []string{
				"second task over",
			},
			ErrorMessage: []string{
				"second task with error",
			},
			Task: func() task.Result {
				fmt.Println("second task (with error)")
				return task.Result{Error: errors.New("Error Triggered")}
			},
		},
		task.Wrapper{
			BeforeMessage: []string{
				"starting third task",
			},
			AfterMessage: []string{
				"third task over",
			},
			ErrorMessage: []string{
				"third task with error",
			},
			Task: func() task.Result {
				fmt.Println("third task (with error)")
				return task.Result{Error: errors.New("Second Error Triggered")}
			},
		},
		task.Wrapper{
			BeforeMessage: []string{
				"starting fourth task",
			},
			AfterMessage: []string{
				"fourth task over",
			},
			ErrorMessage: []string{
				"fourth task with error",
			},
			Task: func() task.Result {
				fmt.Println("fourth task")
				return task.Result{}
			},
		},
	}

	sequence := task.Wrapper{
		BeforeMessage: []string{
			"Starting sequence",
		},
		AfterMessage: []string{
			"Sequence over",
		},
		ErrorMessage: []string{
			"Sequence with errors",
		},
		// Do not ignore errors for every step.
		Task: task.CreateSequence(tasks, []bool{false, false, false, false}, verbosity).Task(),
	}

	sequence.Execute(verbosity)

	// Output:
	// Starting sequence ...
	// starting first task ...
	// First task
	// first task over
	// starting second task ...
	// second task (with error)
	// second task with error
	// Warning from task 1: Error Triggered
	// starting third task ...
	// third task (with error)
	// third task with error
	// Warning from task 2: Second Error Triggered
	// starting fourth task ...
	// fourth task
	// fourth task over
	// Sequence over
}

func ExampleCreateSequence_errors_ignored_first_and_third() {
	verbosity := 1

	tasks := []task.Wrapper{
		task.Wrapper{
			BeforeMessage: []string{
				"starting first task",
			},
			AfterMessage: []string{
				"first task over",
			},
			ErrorMessage: []string{
				"first task with error",
			},
			Task: func() task.Result {
				fmt.Println("First task")
				return task.Result{}
			},
		}, task.Wrapper{
			BeforeMessage: []string{
				"starting second task",
			},
			AfterMessage: []string{
				"second task over",
			},
			ErrorMessage: []string{
				"second task with error",
			},
			Task: func() task.Result {
				fmt.Println("second task (with error)")
				return task.Result{Error: errors.New("Error Triggered")}
			},
		},
		task.Wrapper{
			BeforeMessage: []string{
				"starting third task",
			},
			AfterMessage: []string{
				"third task over",
			},
			ErrorMessage: []string{
				"third task with error",
			},
			Task: func() task.Result {
				fmt.Println("third task (with error)")
				return task.Result{Error: errors.New("Second Error Triggered")}
			},
		},
		task.Wrapper{
			BeforeMessage: []string{
				"starting fourth task",
			},
			AfterMessage: []string{
				"fourth task over",
			},
			ErrorMessage: []string{
				"fourth task with error",
			},
			Task: func() task.Result {
				fmt.Println("fourth task")
				return task.Result{}
			},
		},
	}

	sequence := task.Wrapper{
		BeforeMessage: []string{
			"Starting sequence",
		},
		AfterMessage: []string{
			"Sequence over",
		},
		ErrorMessage: []string{
			"Sequence with errors",
		},
		// Do not ignore errors for every step.
		Task: task.CreateSequence(tasks, []bool{true, false, true, false}, verbosity).Task(),
	}

	sequence.Execute(verbosity)

	// Output:
	// Starting sequence ...
	// starting first task ...
	// First task
	// first task over
	// starting second task ...
	// second task (with error)
	// second task with error
	// Warning from task 1: Error Triggered
	// starting third task ...
	// third task (with error)
	// third task with error
	// starting fourth task ...
	// fourth task
	// fourth task over
	// Sequence over
}
