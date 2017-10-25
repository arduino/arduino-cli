package sortutil

import (
	"reflect"
)

// A Getter is a function which takes a reflect.Value for a slice, and returns a
// a slice of reflect.Value, e.g. a slice with a reflect.Value for each of the
// Name fields from a reflect.Value for a slice of a struct type. It is used by
// the sort functions to identify the elements to sort by.
type Getter func(reflect.Value) []reflect.Value

func valueSlice(l int) []reflect.Value {
	s := make([]reflect.Value, l, l)
	return s
}

// Returns a Getter which returns the values from a reflect.Value for a
// slice. This is the default Getter used if none is passed to Sort.
func SimpleGetter() Getter {
	return func(s reflect.Value) []reflect.Value {
		vals := valueSlice(s.Len())
		for i := range vals {
			vals[i] = reflect.Indirect(reflect.Indirect(s.Index(i)))
		}
		return vals
	}
}

// Returns a Getter which gets fields with name from a reflect.Value for a
// slice of a struct type, returning them as a slice of reflect.Value (one
// Value for each field in each struct.) Can be used with Sort to sort an
// []Object by e.g. Object.Name or Object.Date. A runtime panic will occur if
// the specified field isn't exported.
func FieldGetter(name string) Getter {
	return func(s reflect.Value) []reflect.Value {
		vals := valueSlice(s.Len())
		for i := range vals {
			vals[i] = reflect.Indirect(reflect.Indirect(s.Index(i)).FieldByName(name))
		}
		return vals
	}
}

// Returns a Getter which gets nested fields corresponding to e.g.
// []int{1, 2, 3} = field 3 of field 2 of field 1 of each struct from a
// reflect.Value for a slice of a struct type, returning them as a slice of
// reflect.Value (one Value for each of the indices in the structs.) Can be
// used with Sort to sort an []Object by the first field in the struct
// value of the first field of each Object. A runtime panic will occur if
// the specified field isn't exported.
func FieldByIndexGetter(index []int) Getter {
	return func(s reflect.Value) []reflect.Value {
		vals := valueSlice(s.Len())
		for i := range vals {
			vals[i] = reflect.Indirect(reflect.Indirect(s.Index(i)).FieldByIndex(index))
		}
		return vals
	}
}

// Returns a Getter which gets values with index from a reflect.Value for a
// slice. Can be used with Sort to sort an [][]int by e.g. the second element
// in each nested slice.
func IndexGetter(index int) Getter {
	return func(s reflect.Value) []reflect.Value {
		vals := valueSlice(s.Len())
		for i := range vals {
			vals[i] = reflect.Indirect(reflect.Indirect(s.Index(i)).Index(index))
		}
		return vals
	}
}
