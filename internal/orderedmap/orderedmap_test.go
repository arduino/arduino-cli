// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package orderedmap_test

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/orderedmap"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	t.Run("the key is a string", func(t *testing.T) {
		m := orderedmap.New[string, int]()

		require.Zero(t, m.Get("not-existing-key"))
		v, ok := m.GetOk("not-existing-key")
		require.Zero(t, v)
		require.False(t, ok)

		m.Set("existing-key", 1)
		require.Equal(t, 1, m.Get("existing-key"))
		v, ok = m.GetOk("existing-key")
		require.Equal(t, 1, v)
		require.True(t, ok)

		// test empty key
		m.Set("", 2)
		require.Equal(t, 2, m.Get(""))
		v, ok = m.GetOk("")
		require.Equal(t, 2, v)
		require.True(t, ok)
	})

	t.Run("the key is int", func(t *testing.T) {
		m := orderedmap.New[int, int]()

		require.Equal(t, 0, m.Get(1))
		v, ok := m.GetOk(1)
		require.Zero(t, v)
		require.False(t, ok)

		m.Set(1, 1)
		require.Equal(t, 1, m.Get(1))
		v, ok = m.GetOk(1)
		require.Equal(t, 1, v)
		require.True(t, ok)

		// test empty key
		m.Set(0, 2)
		require.Equal(t, 2, m.Get(0))
		v, ok = m.GetOk(0)
		require.Equal(t, 2, v)
		require.True(t, ok)
	})

	t.Run("the key is a pointer", func(t *testing.T) {
		m := orderedmap.New[*string, int]()

		notExistingKey := toPtr("not-existing-key")
		require.Equal(t, 0, m.Get(notExistingKey))
		v, ok := m.GetOk(notExistingKey)
		require.Zero(t, v)
		require.False(t, ok)

		existingKey := toPtr("existing-key")
		m.Set(existingKey, 1)
		require.Equal(t, 1, m.Get(existingKey))
		v, ok = m.GetOk(existingKey)
		require.Equal(t, 1, v)
		require.True(t, ok)

		// Using a different pointer with the same value returns no result
		require.Equal(t, 0, m.Get(toPtr("existing-key")))
		v, ok = m.GetOk(toPtr("existing-key"))
		require.Zero(t, v)
		require.False(t, ok)

		// test empty key
		m.Set(nil, 2)
		require.Equal(t, 2, m.Get(nil))
		v, ok = m.GetOk(nil)
		require.Equal(t, 2, v)
		require.True(t, ok)
	})

	t.Run("custom comparable key", func(t *testing.T) {
		type A struct {
			b []byte
		}
		m := orderedmap.NewWithConversionFunc[*A, int, string](
			func(a *A) string {
				if a == nil {
					return ""
				}
				return string(a.b)
			},
		)
		require.Zero(t, m.Get(&A{}))
		require.Zero(t, m.Get(nil))

		// Here we're using the conversion function to set the key, using a different
		// pointer to retreive the value works.
		m.Set(&A{b: []byte{60, 61, 62}}, 1)
		require.Equal(t, 1, m.Get(&A{b: []byte{60, 61, 62}}))

		m.Set(nil, 2)
		require.Equal(t, 2, m.Get(nil))
	})
}

func TestSet(t *testing.T) {
	t.Run("insert 3 differnt keys", func(t *testing.T) {
		m := orderedmap.New[string, int]()
		m.Set("a", 1)
		m.Set("b", 2)
		m.Set("c", 3)
		require.Equal(t, 1, m.Get("a"))
		require.Equal(t, 2, m.Get("b"))
		require.Equal(t, 3, m.Get("c"))
	})
	t.Run("insert equal keys", func(t *testing.T) {
		m := orderedmap.New[string, int]()
		m.Set("a", 1)
		m.Set("a", 2)
		require.Equal(t, 2, m.Get("a"))
	})
}

func TestSize(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		m := orderedmap.New[string, bool]()
		require.Zero(t, m.Size())
	})
	t.Run("one element", func(t *testing.T) {
		m := orderedmap.New[string, bool]()
		m.Set("a", true)
		require.Equal(t, 1, m.Size())
	})
	t.Run("three elements", func(t *testing.T) {
		m := orderedmap.New[string, bool]()
		m.Set("a", true)
		m.Set("b", true)
		m.Set("c", true)
		require.Equal(t, 3, m.Size())
	})
	t.Run("insert same keys result in size 1", func(t *testing.T) {
		m := orderedmap.New[string, bool]()
		m.Set("a", true)
		m.Set("a", false)
		m.Set("a", true)
		require.Equal(t, 1, m.Size())
	})
}

func TestKeys(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		m := orderedmap.New[string, bool]()
		require.Empty(t, m.Keys())
	})
	t.Run("one element", func(t *testing.T) {
		m := orderedmap.New[string, bool]()
		m.Set("a", true)
		require.Len(t, m.Keys(), 1)
	})
	t.Run("keys respect order of insertion", func(t *testing.T) {
		m := orderedmap.New[string, bool]()
		m.Set("a", true)
		m.Set("b", true)
		m.Set("c", true)
		require.Equal(t, []string{"a", "b", "c"}, m.Keys())
	})
	t.Run("replacing a key changes the ordering", func(t *testing.T) {
		m := orderedmap.New[string, bool]()
		m.Set("a", true)
		m.Set("b", true)
		m.Set("a", false)
		require.Equal(t, []string{"b", "a"}, m.Keys())
	})
	t.Run("delete a key", func(t *testing.T) {
		m := orderedmap.New[string, bool]()
		m.Set("a", true)
		m.Set("b", true)
		m.Remove("a")
		require.Equal(t, []string{"b"}, m.Keys())
	})
}

func TestRemove(t *testing.T) {
	t.Run("key doesn't exist", func(t *testing.T) {
		m := orderedmap.New[string, bool]()
		require.Zero(t, m.Get("not-existing-key"))
		m.Remove("not-existing-key")
		require.Zero(t, m.Get("not-existing-key"))
	})
	t.Run("key exist", func(t *testing.T) {
		m := orderedmap.New[string, bool]()
		m.Set("a", true)
		require.True(t, m.Get("a"))

		m.Remove("a")
		require.False(t, m.Get("a"))
	})
	t.Run("key deletion doesn't affect other keys", func(t *testing.T) {
		m := orderedmap.New[string, bool]()
		m.Set("a", true)
		m.Set("b", true)
		m.Remove("a")
		require.True(t, m.Get("b"))
		_, ok := m.GetOk("a")
		require.False(t, ok)
	})
}

func toPtr[V any](v V) *V {
	return &v
}
