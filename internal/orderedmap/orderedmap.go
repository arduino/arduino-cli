package orderedmap

import (
	"bytes"
	"encoding/json"
	"reflect"
	"slices"
)

// Map is a container of properties
type Map[K comparable, V any] struct {
	kv map[K]V
	o  []K
}

// New returns a new Map
func New[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		kv: map[K]V{},
		o:  []K{},
	}
}

// Get retrieves the value corresponding to key
func (m *Map[K, V]) Get(key K) V {
	return m.kv[key]
}

// GetOk retrieves the value corresponding to key and returns a true/false indicator
// to check if the key is present in the map (true if the key is present)
func (m *Map[K, V]) GetOk(key K) (V, bool) {
	v, ok := m.kv[key]
	return v, ok
}

// ContainsKey returns true if the map contains the specified key
func (m *Map[K, V]) ContainsKey(key K) bool {
	_, has := m.kv[key]
	return has
}

// MarshalJSON marshal the map into json mantaining the order of the key
func (m *Map[K, V]) MarshalJSON() ([]byte, error) {
	if m.Size() == 0 {
		return []byte("{}"), nil
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	encoder := json.NewEncoder(&buf)
	for _, k := range m.o {
		if err := encoder.Encode(k); err != nil {
			return nil, err
		}
		buf.WriteByte(':')
		if err := encoder.Encode(m.kv[k]); err != nil {
			return nil, err
		}
		buf.WriteByte(',')
	}
	buf.Truncate(buf.Len() - 1) // remove last `,`
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// Set inserts or replaces an existing key-value pair in the map
func (m *Map[K, V]) Set(key K, value V) {
	if _, has := m.kv[key]; has {
		m.Remove(key)
	}
	m.kv[key] = value
	m.o = append(m.o, key)
}

// Size returns the number of elements in the map
func (m *Map[K, V]) Size() int {
	return len(m.kv)
}

// Remove removes the key from the map
func (m *Map[K, V]) Remove(key K) {
	delete(m.kv, key)
	for i, k := range m.o {
		if k == key {
			m.o = append(m.o[:i], m.o[i+1:]...)
			return
		}
	}
}

// Merge merges other Maps into this one. Each key/value of the merged Maps replaces
// the key/value present in the original Map.
func (m *Map[K, V]) Merge(sources ...*Map[K, V]) *Map[K, V] {
	for _, source := range sources {
		for _, key := range source.o {
			value := source.kv[key]
			m.Set(key, value)
		}
	}
	return m
}

// Keys returns an array of the keys contained in the Map
func (m *Map[K, V]) Keys() []K {
	keys := make([]K, len(m.o))
	copy(keys, m.o)
	return keys
}

func (m *Map[K, V]) SortKeys(f func(x, y K) int) {
	slices.SortFunc(m.o, f)
}

func (m *Map[K, V]) SortStableKeys(f func(x, y K) int) {
	slices.SortStableFunc(m.o, f)
}

// Values returns an array of the values contained in the Map. Duplicated
// values are repeated in the list accordingly.
func (m *Map[K, V]) Values() []V {
	values := make([]V, len(m.o))
	for i, key := range m.o {
		values[i] = m.kv[key]
	}
	return values
}

// AsMap returns the underlying map[string]string. This is useful if you need to
// for ... range but without the requirement of the ordered elements.
func (m *Map[K, V]) AsMap() map[K]V {
	return m.kv
}

// Clone makes a copy of the Map
func (m *Map[K, V]) Clone() *Map[K, V] {
	clone := New[K, V]()
	clone.Merge(m)
	return clone
}

// Equals returns true if the current Map contains the same key/value pairs of
// the Map passed as argument, the order of insertion does not matter.
func (m *Map[K, V]) Equals(other *Map[K, V]) bool {
	return reflect.DeepEqual(m.kv, other.kv)
}

// EqualsWithOrder returns true if the current Map contains the same key/value pairs of
// the Map passed as argument with the same order of insertion.
func (m *Map[K, V]) EqualsWithOrder(other *Map[K, V]) bool {
	return reflect.DeepEqual(m.o, other.o) && reflect.DeepEqual(m.kv, other.kv)
}
