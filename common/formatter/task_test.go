package formatter

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bcmi-labs/arduino-cli/task"
)

func captureStdout(toCapture func()) []byte {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	toCapture()

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	fmt.Printf("Captured >>>\n%s<<<\n", out)
	return out
}
func TestCreateSequence(t *testing.T) {
	task1 := WrapTask(
		func() task.Result {
			fmt.Println("First task")
			return task.Result{}
		},
		&TaskWrapperMessages{
			BeforeMessage: "starting first task",
			AfterMessage:  "first task over",
		})
	task2 := WrapTask(
		func() task.Result {
			fmt.Println("second task")
			return task.Result{}
		},
		&TaskWrapperMessages{
			BeforeMessage: "starting second task",
			AfterMessage:  "second task over",
		})
	task3 := WrapTask(
		func() task.Result {
			fmt.Println("third task")
			return task.Result{}
		},
		&TaskWrapperMessages{
			BeforeMessage: "starting third task",
			AfterMessage:  "third task over",
		})
	tasks := []task.Task{task1, task2, task3}

	sequence := WrapTask(task.CreateSequence(tasks, []bool{true, true, true}).Task(),
		&TaskWrapperMessages{
			BeforeMessage: "Starting sequence",
			AfterMessage:  "Sequence over",
		})

	out := captureStdout(func() { sequence() })
	require.Equal(t, `Starting sequence
starting first task
First task
first task over
starting second task
second task
second task over
starting third task
third task
third task over
Sequence over
`, string(out), "Sequence output")

}

func TestCreateSequenceWithErrorsIgnored(t *testing.T) {
	task1 := WrapTask(
		func() task.Result {
			fmt.Println("First task")
			return task.Result{}
		},
		&TaskWrapperMessages{
			BeforeMessage: "starting first task",
			AfterMessage:  "first task over",
			ErrorMessage:  "first task with error",
		})
	task2 := WrapTask(
		func() task.Result {
			fmt.Println("second task (with error)")
			return task.Result{Error: errors.New("Error Triggered")}
		},
		&TaskWrapperMessages{
			BeforeMessage: "starting second task",
			AfterMessage:  "second task over",
			ErrorMessage:  "second task with error",
		})
	task3 := WrapTask(
		func() task.Result {
			fmt.Println("third task (with error)")
			return task.Result{Error: errors.New("Second Error Triggered")}
		},
		&TaskWrapperMessages{
			BeforeMessage: "starting third task",
			AfterMessage:  "third task over",
			ErrorMessage:  "third task with error",
		})
	task4 := WrapTask(
		func() task.Result {
			fmt.Println("fourth task")
			return task.Result{}
		},
		&TaskWrapperMessages{
			BeforeMessage: "starting fourth task",
			AfterMessage:  "fourth task over",
			ErrorMessage:  "fourth task with error",
		})
	tasks := []task.Task{task1, task2, task3, task4}

	// Do not ignore errors for every step.
	sequence := WrapTask(
		task.CreateSequence(tasks, []bool{true, true, true, true}).Task(),
		&TaskWrapperMessages{
			BeforeMessage: "Starting sequence",
			AfterMessage:  "Sequence over",
			ErrorMessage:  "Sequence with errors",
		})

	out := captureStdout(func() { sequence() })
	require.Equal(t, `Starting sequence
starting first task
First task
first task over
starting second task
second task (with error)
second task with error
starting third task
third task (with error)
third task with error
starting fourth task
fourth task
fourth task over
Sequence over
`, string(out), "Sequence output")
}

func TestCreateSequenceWithErrorsDisplayed(t *testing.T) {
	task1 := WrapTask(
		func() task.Result {
			fmt.Println("First task")
			return task.Result{}
		},
		&TaskWrapperMessages{
			BeforeMessage: "starting first task",
			AfterMessage:  "first task over",
			ErrorMessage:  "first task with error",
		})
	task2 := WrapTask(
		func() task.Result {
			fmt.Println("second task (with error)")
			return task.Result{Error: errors.New("Error Triggered")}
		},
		&TaskWrapperMessages{
			BeforeMessage: "starting second task",
			AfterMessage:  "second task over",
			ErrorMessage:  "second task with error",
		})
	task3 := WrapTask(
		func() task.Result {
			fmt.Println("third task (with error)")
			return task.Result{Error: errors.New("Second Error Triggered")}
		},
		&TaskWrapperMessages{
			BeforeMessage: "starting third task",
			AfterMessage:  "third task over",
			ErrorMessage:  "third task with error",
		})
	task4 := WrapTask(
		func() task.Result {
			fmt.Println("fourth task")
			return task.Result{}
		},
		&TaskWrapperMessages{
			BeforeMessage: "starting fourth task",
			AfterMessage:  "fourth task over",
			ErrorMessage:  "fourth task with error",
		})
	tasks := []task.Task{task1, task2, task3, task4}

	sequence := WrapTask(
		// Do not ignore errors for every step.
		task.CreateSequence(tasks, []bool{false, false, false, false}).Task(),
		&TaskWrapperMessages{
			BeforeMessage: "Starting sequence",
			AfterMessage:  "Sequence over",
			ErrorMessage:  "Sequence with errors",
		})

	out := captureStdout(func() { sequence() })
	require.Equal(t, `Starting sequence
starting first task
First task
first task over
starting second task
second task (with error)
second task with error
Warning from task 1: Error Triggered
starting third task
third task (with error)
third task with error
Warning from task 2: Second Error Triggered
starting fourth task
fourth task
fourth task over
Sequence over
`, string(out), "Sequence output")

}

func TestCreateSequenceWithPartialErrorsIgnored(t *testing.T) {
	task1 := WrapTask(
		func() task.Result {
			fmt.Println("First task")
			return task.Result{}
		},
		&TaskWrapperMessages{
			BeforeMessage: "starting first task",
			AfterMessage:  "first task over",
			ErrorMessage:  "first task with error",
		})
	task2 := WrapTask(
		func() task.Result {
			fmt.Println("second task (with error)")
			return task.Result{Error: errors.New("Error Triggered")}
		},
		&TaskWrapperMessages{
			BeforeMessage: "starting second task",
			AfterMessage:  "second task over",
			ErrorMessage:  "second task with error",
		})
	task3 := WrapTask(
		func() task.Result {
			fmt.Println("third task (with error)")
			return task.Result{Error: errors.New("Second Error Triggered")}
		},
		&TaskWrapperMessages{
			BeforeMessage: "starting third task",
			AfterMessage:  "third task over",
			ErrorMessage:  "third task with error",
		})
	task4 := WrapTask(
		func() task.Result {
			fmt.Println("fourth task")
			return task.Result{}
		},
		&TaskWrapperMessages{
			BeforeMessage: "starting fourth task",
			AfterMessage:  "fourth task over",
			ErrorMessage:  "fourth task with error",
		})
	tasks := []task.Task{task1, task2, task3, task4}

	sequence := WrapTask(
		// Do not ignore errors for every step.
		task.CreateSequence(tasks, []bool{true, false, true, false}).Task(),
		&TaskWrapperMessages{
			BeforeMessage: "Starting sequence",
			AfterMessage:  "Sequence over",
			ErrorMessage:  "Sequence with errors",
		})

	out := captureStdout(func() { sequence() })
	require.Equal(t, `Starting sequence
starting first task
First task
first task over
starting second task
second task (with error)
second task with error
Warning from task 1: Error Triggered
starting third task
third task (with error)
third task with error
starting fourth task
fourth task
fourth task over
Sequence over
`, string(out), "Sequence output")

}
