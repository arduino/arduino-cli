package task_test

import (
	"errors"
	"fmt"

	"github.com/bcmi-labs/arduino-cli/task"
)

func ExampleCreateSequence() {
	tasks := []task.Wrapper{
		task.Wrapper{
			BeforeMessage: "starting first task",
			AfterMessage:  "first task over",
			Task: func() task.Result {
				fmt.Println("First task")
				return task.Result{}
			},
		}, task.Wrapper{
			BeforeMessage: "starting second task",
			AfterMessage:  "second task over",
			Task: func() task.Result {
				fmt.Println("second task")
				return task.Result{}
			},
		},
		task.Wrapper{
			BeforeMessage: "starting third task",
			AfterMessage:  "third task over",
			Task: func() task.Result {
				fmt.Println("third task")
				return task.Result{}
			},
		},
	}

	sequence := task.Wrapper{
		BeforeMessage: "Starting sequence",
		AfterMessage:  "Sequence over",
		Task:          task.CreateSequence(tasks, []bool{true, true, true}).Task(),
	}

	sequence.Execute()

	// Output:
	// Starting sequence
	// starting first task
	// First task
	// first task over
	// starting second task
	// second task
	// second task over
	// starting third task
	// third task
	// third task over
	// Sequence over
}

func ExampleCreateSequence_errors_ignored() {
	tasks := []task.Wrapper{
		task.Wrapper{
			BeforeMessage: "starting first task",
			AfterMessage:  "first task over",
			ErrorMessage:  "first task with error",
			Task: func() task.Result {
				fmt.Println("First task")
				return task.Result{}
			},
		}, task.Wrapper{
			BeforeMessage: "starting second task",
			AfterMessage:  "second task over",
			ErrorMessage:  "second task with error",
			Task: func() task.Result {
				fmt.Println("second task (with error)")
				return task.Result{Error: errors.New("Error Triggered")}
			},
		},
		task.Wrapper{
			BeforeMessage: "starting third task",
			AfterMessage:  "third task over",
			ErrorMessage:  "third task with error",
			Task: func() task.Result {
				fmt.Println("third task (with error)")
				return task.Result{Error: errors.New("Second Error Triggered")}
			},
		},
		task.Wrapper{
			BeforeMessage: "starting fourth task",
			AfterMessage:  "fourth task over",
			ErrorMessage:  "fourth task with error",
			Task: func() task.Result {
				fmt.Println("fourth task")
				return task.Result{}
			},
		},
	}

	sequence := task.Wrapper{
		BeforeMessage: "Starting sequence",
		AfterMessage:  "Sequence over",
		ErrorMessage:  "Sequence with errors",
		// Do not ignore errors for every step.
		Task: task.CreateSequence(tasks, []bool{true, true, true, true}).Task(),
	}

	sequence.Execute()

	// Output:
	// Starting sequence
	// starting first task
	// First task
	// first task over
	// starting second task
	// second task (with error)
	// second task with error
	// starting third task
	// third task (with error)
	// third task with error
	// starting fourth task
	// fourth task
	// fourth task over
	// Sequence over
}

func ExampleCreateSequence_errors_shown_all() {
	tasks := []task.Wrapper{
		task.Wrapper{
			BeforeMessage: "starting first task",
			AfterMessage:  "first task over",
			ErrorMessage:  "first task with error",
			Task: func() task.Result {
				fmt.Println("First task")
				return task.Result{}
			},
		}, task.Wrapper{
			BeforeMessage: "starting second task",
			AfterMessage:  "second task over",
			ErrorMessage:  "second task with error",
			Task: func() task.Result {
				fmt.Println("second task (with error)")
				return task.Result{Error: errors.New("Error Triggered")}
			},
		},
		task.Wrapper{
			BeforeMessage: "starting third task",
			AfterMessage:  "third task over",
			ErrorMessage:  "third task with error",
			Task: func() task.Result {
				fmt.Println("third task (with error)")
				return task.Result{Error: errors.New("Second Error Triggered")}
			},
		},
		task.Wrapper{
			BeforeMessage: "starting fourth task",
			AfterMessage:  "fourth task over",
			ErrorMessage:  "fourth task with error",
			Task: func() task.Result {
				fmt.Println("fourth task")
				return task.Result{}
			},
		},
	}

	sequence := task.Wrapper{
		BeforeMessage: "Starting sequence",
		AfterMessage:  "Sequence over",
		ErrorMessage:  "Sequence with errors",
		// Do not ignore errors for every step.
		Task: task.CreateSequence(tasks, []bool{false, false, false, false}).Task(),
	}

	sequence.Execute()

	// Output:
	// Starting sequence
	// starting first task
	// First task
	// first task over
	// starting second task
	// second task (with error)
	// second task with error
	// Warning from task 1: Error Triggered
	// starting third task
	// third task (with error)
	// third task with error
	// Warning from task 2: Second Error Triggered
	// starting fourth task
	// fourth task
	// fourth task over
	// Sequence over
}

func ExampleCreateSequence_errors_ignored_first_and_third() {
	tasks := []task.Wrapper{
		task.Wrapper{
			BeforeMessage: "starting first task",
			AfterMessage:  "first task over",
			ErrorMessage:  "first task with error",
			Task: func() task.Result {
				fmt.Println("First task")
				return task.Result{}
			},
		}, task.Wrapper{
			BeforeMessage: "starting second task",
			AfterMessage:  "second task over",
			ErrorMessage:  "second task with error",
			Task: func() task.Result {
				fmt.Println("second task (with error)")
				return task.Result{Error: errors.New("Error Triggered")}
			},
		},
		task.Wrapper{
			BeforeMessage: "starting third task",
			AfterMessage:  "third task over",
			ErrorMessage:  "third task with error",
			Task: func() task.Result {
				fmt.Println("third task (with error)")
				return task.Result{Error: errors.New("Second Error Triggered")}
			},
		},
		task.Wrapper{
			BeforeMessage: "starting fourth task",
			AfterMessage:  "fourth task over",
			ErrorMessage:  "fourth task with error",
			Task: func() task.Result {
				fmt.Println("fourth task")
				return task.Result{}
			},
		},
	}

	sequence := task.Wrapper{
		BeforeMessage: "Starting sequence",
		AfterMessage:  "Sequence over",
		ErrorMessage:  "Sequence with errors",
		// Do not ignore errors for every step.
		Task: task.CreateSequence(tasks, []bool{true, false, true, false}).Task(),
	}

	sequence.Execute()

	// Output:
	// Starting sequence
	// starting first task
	// First task
	// first task over
	// starting second task
	// second task (with error)
	// second task with error
	// Warning from task 1: Error Triggered
	// starting third task
	// third task (with error)
	// third task with error
	// starting fourth task
	// fourth task
	// fourth task over
	// Sequence over
}
