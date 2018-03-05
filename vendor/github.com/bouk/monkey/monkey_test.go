package monkey_test

import (
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
)

func no() bool  { return false }
func yes() bool { return true }

func TestTimePatch(t *testing.T) {
	before := time.Now()
	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	})
	during := time.Now()
	assert.True(t, monkey.Unpatch(time.Now))
	after := time.Now()

	assert.Equal(t, time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC), during)
	assert.NotEqual(t, before, during)
	assert.NotEqual(t, during, after)
}

func TestGC(t *testing.T) {
	value := true
	monkey.Patch(no, func() bool {
		return value
	})
	defer monkey.UnpatchAll()
	runtime.GC()
	assert.True(t, no())
}

func TestSimple(t *testing.T) {
	assert.False(t, no())
	monkey.Patch(no, yes)
	assert.True(t, no())
	assert.True(t, monkey.Unpatch(no))
	assert.False(t, no())
	assert.False(t, monkey.Unpatch(no))
}

func TestGuard(t *testing.T) {
	var guard *monkey.PatchGuard
	guard = monkey.Patch(no, func() bool {
		guard.Unpatch()
		defer guard.Restore()

		return !no()
	})
	for i := 0; i < 100; i++ {
		assert.True(t, no())
	}
	monkey.Unpatch(no)
}

func TestUnpatchAll(t *testing.T) {
	assert.False(t, no())
	monkey.Patch(no, yes)
	assert.True(t, no())
	monkey.UnpatchAll()
	assert.False(t, no())
}

type s struct{}

func (s *s) yes() bool { return true }

func TestWithInstanceMethod(t *testing.T) {
	i := &s{}

	assert.False(t, no())
	monkey.Patch(no, i.yes)
	assert.True(t, no())
	monkey.Unpatch(no)
	assert.False(t, no())
}

type f struct{}

func (f *f) no() bool { return false }

func TestOnInstanceMethod(t *testing.T) {
	i := &f{}
	assert.False(t, i.no())
	monkey.PatchInstanceMethod(reflect.TypeOf(i), "no", func(_ *f) bool { return true })
	assert.True(t, i.no())
	assert.True(t, monkey.UnpatchInstanceMethod(reflect.TypeOf(i), "no"))
	assert.False(t, i.no())
}

func TestNotFunction(t *testing.T) {
	assert.Panics(t, func() {
		monkey.Patch(no, 1)
	})
	assert.Panics(t, func() {
		monkey.Patch(1, yes)
	})
}

func TestNotCompatible(t *testing.T) {
	assert.Panics(t, func() {
		monkey.Patch(no, func() {})
	})
}
