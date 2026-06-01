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

package orderedmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
)

// Map is a map that keeps ordering insertion.
type Map[K any, V any] interface {
	Get(K) V
	GetOk(key K) (V, bool)
	Set(K, V)
	Size() int
	ContainsKey(key K) bool
	Keys() []K
	Merge(...Map[K, V]) Map[K, V]
	SortKeys(f func(x, y K) int)
	SortStableKeys(f func(x, y K) int)
	Values() []V
	Clone() Map[K, V]
	Remove(key K)
	MarshalJSON() ([]byte, error)
}

type scalars interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 |
		~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~string
}

// NewWithConversionFunc creates a map using the given conversion function
// to convert non-comparable key type to comparable items.
// The conversion function must be bijective.
func NewWithConversionFunc[K any, V any, C scalars](conv func(K) C) Map[K, V] {
	return &mapImpl[K, V, C]{
		conv: conv,
		kv:   map[C]V{},
		o:    []K{},
	}
}

// New creates a map
func New[K scalars, V any]() Map[K, V] {
	return &mapImpl[K, V, K]{
		conv: func(in K) K { return in }, // identity
		kv:   map[K]V{},
		o:    []K{},
	}
}

type mapImpl[K any, V any, C scalars] struct {
	conv func(K) C
	kv   map[C]V
	o    []K
}

// Get retrieves the value corresponding to key
func (m *mapImpl[K, V, C]) Get(key K) V {
	return m.kv[m.conv(key)]
}

// GetOk retrieves the value corresponding to key and returns a true/false indicator
// to check if the key is present in the map (true if the key is present)
func (m *mapImpl[K, V, C]) GetOk(key K) (V, bool) {
	v, ok := m.kv[m.conv(key)]
	return v, ok
}

// ContainsKey returns true if the map contains the specified key
func (m *mapImpl[K, V, C]) ContainsKey(key K) bool {
	_, has := m.kv[m.conv(key)]
	return has
}

// MarshalJSON marshal the map into json mantaining the order of the key
func (m *mapImpl[K, V, C]) MarshalJSON() ([]byte, error) {
	if m.Size() == 0 {
		return []byte("{}"), nil
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	encoder := json.NewEncoder(&buf)
	for _, k := range m.o {
		// Here we convert non string keys in string.
		if err := encoder.Encode(fmt.Sprintf("%v", m.conv(k))); err != nil {
			return nil, err
		}
		buf.WriteByte(':')
		if err := encoder.Encode(m.kv[m.conv(k)]); err != nil {
			return nil, err
		}
		buf.WriteByte(',')
	}
	buf.Truncate(buf.Len() - 1) // remove last `,`
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// Set inserts or replaces an existing key-value pair in the map
func (m *mapImpl[K, V, C]) Set(key K, value V) {
	compKey := m.conv(key)
	if _, has := m.kv[compKey]; has {
		m.Remove(key)
	}
	m.kv[compKey] = value
	m.o = append(m.o, key)
}

// Size returns the number of elements in the map
func (m *mapImpl[K, V, C]) Size() int {
	return len(m.kv)
}

// Remove removes the key from the map
func (m *mapImpl[K, V, C]) Remove(key K) {
	compKey := m.conv(key)
	delete(m.kv, compKey)
	for i, k := range m.o {
		if m.conv(k) == compKey {
			m.o = append(m.o[:i], m.o[i+1:]...)
			return
		}
	}
}

// Merge merges other Maps into this one. Each key/value of the merged Maps replaces
// the key/value present in the original Map.
func (m *mapImpl[K, V, C]) Merge(sources ...Map[K, V]) Map[K, V] {
	for _, source := range sources {
		for _, key := range source.Keys() {
			value := source.Get(key)
			m.Set(key, value)
		}
	}
	return m
}

// Keys returns an array of the keys contained in the Map
func (m *mapImpl[K, V, C]) Keys() []K {
	keys := make([]K, len(m.o))
	copy(keys, m.o)
	return keys
}

func (m *mapImpl[K, V, C]) SortKeys(f func(x, y K) int) {
	slices.SortFunc(m.o, f)
}

func (m *mapImpl[K, V, C]) SortStableKeys(f func(x, y K) int) {
	slices.SortStableFunc(m.o, f)
}

// Values returns an array of the values contained in the Map. Duplicated
// values are repeated in the list accordingly.
func (m *mapImpl[K, V, C]) Values() []V {
	values := make([]V, len(m.o))
	for i, key := range m.o {
		values[i] = m.kv[m.conv(key)]
	}
	return values
}

// Clone makes a copy of the Map
func (m *mapImpl[K, V, C]) Clone() Map[K, V] {
	clone := NewWithConversionFunc[K, V, C](m.conv)
	clone.Merge(m)
	return clone
}
