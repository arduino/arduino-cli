package multierror

import (
	"errors"
	"fmt"
	"testing"
)

func Works() error { return nil }
func Fails() error { return fmt.Errorf("Failed") }

func FailMulti(msg string, n int) error {
	var e Accumulator
	for i := 0; i < n; i++ {
		e.Pushf("%s%d", msg, i+1)
	}
	return e.Error()
}

func ReturnsEmptyMultiError() error {
	var e Accumulator
	e.Push(Works())
	e.Push(Works())
	return e.Error()
}

func TestEmptyMultiError(t *testing.T) {
	var e Accumulator
	e.Push(Works())
	e.Push(Works())
	if e != nil {
		t.Fatal("Empty Accumulator is not nil")
	}

	if ReturnsEmptyMultiError() != nil {
		t.Fatal("Empty Accumulator return value is not nil")
	}
}

func ReturnsNonEmptyMultiError() error {
	var e Accumulator
	e.Push(Fails())
	e.Push(Fails())
	return e.Error()
}

func TestNonEmptyMultiError(t *testing.T) {
	var e Accumulator
	e.Push(Fails())
	e.Push(Fails())
	if e == nil {
		t.Fatal("Non-empty Accumulator is nil")
	}

	if ReturnsNonEmptyMultiError() == nil {
		t.Fatal("Non-empty Accumulator return value is nil")
	}
}

func TestSingleErrorReturnedDirectly(t *testing.T) {
	var err = errors.New("some error")
	var e Accumulator
	e.Push(err)
	if e.Error() != err {
		t.Error("Expected a single error to be directly returned rather than " +
			"wrapped by an _error helper")
	}
}

func TestPushingMultiError(t *testing.T) {
	var e Accumulator
	e.Push(FailMulti("Fail", 2))
	if fmt.Sprint(e) != "2 errors:\n:   Fail1\n:   Fail2" {
		t.Errorf("Incorrect error string: %q", e)
	}

	e.Push(FailMulti("X", 3))

	// Test some Pushf/PushWithf.
	e.Pushf("Y%d", 1)
	e.PushWithf("Y %s %d %d", nil, 2, 3) // Should be ignored.
	e.PushWithf("Y %s %d %d", errors.New("Z"), 4, 5)

	if fmt.Sprint(e) != `7 errors:
:   Fail1
:   Fail2
:   X1
:   X2
:   X3
:   Y1
:   Y Z 4 5` {
		t.Errorf("Incorrect error string: %q", e)
	}
}

func TestMultiErrorStringification(t *testing.T) {
	var e Accumulator
	if fmt.Sprint(e) != `nil` {
		t.Errorf("Incorrect error string: %q ", e)
	}

	e.Push(Fails())
	if fmt.Sprint(e) != `Failed` {
		t.Errorf("Incorrect error string: %q ", e)
	}

	e.Push(Fails())
	if fmt.Sprint(e) != "2 errors:\n:   Failed\n:   Failed" {
		t.Errorf("Incorrect error string: %q ", e)
	}

	if ReturnsNonEmptyMultiError().Error() != "2 errors:\n:   Failed\n:   Failed" {
		t.Errorf("Incorrect error string: %q", ReturnsNonEmptyMultiError())
	}
}
