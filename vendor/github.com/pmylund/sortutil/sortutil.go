package sortutil

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"time"
)

// Ordering decides the order in which the specified data is sorted.
type Ordering int

func (o Ordering) String() string {
	return orderings[o]
}

// A runtime panic will occur if case-insensitive is used when not sorting by
// a string type.
const (
	Ascending Ordering = iota
	Descending
	CaseInsensitiveAscending
	CaseInsensitiveDescending
)

var orderings = []string{
	"Ascending",
	"Descending",
	"CaseInsensitiveAscending",
	"CaseInsensitiveDescending",
}

// Recognized non-standard types
var (
	t_time = reflect.TypeOf(time.Time{})
)

// A reflecting sort.Interface adapter.
type Sorter struct {
	Slice    reflect.Value
	Getter   Getter
	Ordering Ordering
	itemType reflect.Type    // Type of items being sorted
	vals     []reflect.Value // Nested/child values that we're sorting by
	valKind  reflect.Kind
	valType  reflect.Type
}

// Sort the values in s.Slice by retrieving comparison items using
// s.Getter(s.Slice). A runtime panic will occur if s.Getter is not applicable
// to s.Slice, or if the values retrieved by s.Getter can't be compared, i.e.
// are unrecognized types.
func (s *Sorter) Sort() {
	if s.Slice.Len() < 2 {
		// Nothing to sort
		return
	}
	if s.Getter == nil {
		s.Getter = SimpleGetter()
	}
	s.itemType = s.Slice.Index(0).Type()
	s.vals = s.Getter(s.Slice)
	one := s.vals[0]
	s.valType = one.Type()
	s.valKind = one.Kind()
	switch s.valKind {
	// If the value isn't a standard kind, find a known type to sort by
	default:
		switch s.valType {
		default:
			panic(fmt.Sprintf("Cannot sort by type %v", s.valType))
		case t_time:
			switch s.Ordering {
			default:
				panic(fmt.Sprintf("Invalid ordering %v for time.Time", s.Ordering))
			case Ascending:
				sort.Sort(timeAscending{s})
			case Descending:
				sort.Sort(timeDescending{s})
			}
		}
	// Strings
	case reflect.String:
		switch s.Ordering {
		default:
			panic(fmt.Sprintf("Invalid ordering %v for strings", s.Ordering))
		case Ascending:
			sort.Sort(stringAscending{s})
		case Descending:
			sort.Sort(stringDescending{s})
		case CaseInsensitiveAscending:
			sort.Sort(stringInsensitiveAscending{s})
		case CaseInsensitiveDescending:
			sort.Sort(stringInsensitiveDescending{s})
		}
	// Booleans
	case reflect.Bool:
		switch s.Ordering {
		default:
			panic(fmt.Sprintf("Invalid ordering %v for booleans", s.Ordering))
		case Ascending:
			sort.Sort(boolAscending{s})
		case Descending:
			sort.Sort(boolDescending{s})
		}
	// Ints
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch s.Ordering {
		default:
			panic(fmt.Sprintf("Invalid ordering %v for ints", s.Ordering))
		case Ascending:
			sort.Sort(intAscending{s})
		case Descending:
			sort.Sort(intDescending{s})
		}
	// Uints
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch s.Ordering {
		default:
			panic(fmt.Sprintf("Invalid ordering %v for uints", s.Ordering))
		case Ascending:
			sort.Sort(uintAscending{s})
		case Descending:
			sort.Sort(uintDescending{s})
		}
	// Floats
	case reflect.Float32, reflect.Float64:
		switch s.Ordering {
		default:
			panic(fmt.Sprintf("Invalid ordering %v for floats", s.Ordering))
		case Ascending:
			sort.Sort(floatAscending{s})
		case Descending:
			sort.Sort(floatDescending{s})
		}
	}
}

// Returns the length of the slice being sorted.
func (s *Sorter) Len() int {
	return len(s.vals)
}

// Swaps two indices in the slice being sorted.
func (s *Sorter) Swap(i, j int) {
	x := s.Slice.Index(i)
	y := s.Slice.Index(j)
	tmp := reflect.New(s.itemType).Elem()
	tmp.Set(x)
	x.Set(y)
	y.Set(tmp)
}

// *cough* typedef *cough*
type stringAscending struct{ *Sorter }
type stringDescending struct{ *Sorter }
type stringInsensitiveAscending struct{ *Sorter }
type stringInsensitiveDescending struct{ *Sorter }
type boolAscending struct{ *Sorter }
type boolDescending struct{ *Sorter }
type intAscending struct{ *Sorter }
type intDescending struct{ *Sorter }
type uintAscending struct{ *Sorter }
type uintDescending struct{ *Sorter }
type floatAscending struct{ *Sorter }
type floatDescending struct{ *Sorter }
type timeAscending struct{ *Sorter }
type timeDescending struct{ *Sorter }
type reverser struct{ *Sorter }

func (s stringAscending) Less(i, j int) bool {
	return s.Sorter.vals[i].String() < s.Sorter.vals[j].String()
}

func (s stringDescending) Less(i, j int) bool {
	return s.Sorter.vals[i].String() > s.Sorter.vals[j].String()
}

func (s stringInsensitiveAscending) Less(i, j int) bool {
	return strings.ToLower(s.Sorter.vals[i].String()) < strings.ToLower(s.Sorter.vals[j].String())
}

func (s stringInsensitiveDescending) Less(i, j int) bool {
	return strings.ToLower(s.Sorter.vals[i].String()) > strings.ToLower(s.Sorter.vals[j].String())
}

func (s boolAscending) Less(i, j int) bool {
	return !s.Sorter.vals[i].Bool() && s.Sorter.vals[j].Bool()
}
func (s boolDescending) Less(i, j int) bool {
	return s.Sorter.vals[i].Bool() && !s.Sorter.vals[j].Bool()
}

func (s intAscending) Less(i, j int) bool   { return s.Sorter.vals[i].Int() < s.Sorter.vals[j].Int() }
func (s intDescending) Less(i, j int) bool  { return s.Sorter.vals[i].Int() > s.Sorter.vals[j].Int() }
func (s uintAscending) Less(i, j int) bool  { return s.Sorter.vals[i].Uint() < s.Sorter.vals[j].Uint() }
func (s uintDescending) Less(i, j int) bool { return s.Sorter.vals[i].Uint() > s.Sorter.vals[j].Uint() }

func (s floatAscending) Less(i, j int) bool {
	a := s.Sorter.vals[i].Float()
	b := s.Sorter.vals[j].Float()
	return a < b || math.IsNaN(a) && !math.IsNaN(b)
}

func (s floatDescending) Less(i, j int) bool {
	a := s.Sorter.vals[i].Float()
	b := s.Sorter.vals[j].Float()
	return a > b || !math.IsNaN(a) && math.IsNaN(b)
}

func (s timeAscending) Less(i, j int) bool {
	return s.Sorter.vals[i].Interface().(time.Time).Before(s.Sorter.vals[j].Interface().(time.Time))
}

func (s timeDescending) Less(i, j int) bool {
	return s.Sorter.vals[i].Interface().(time.Time).After(s.Sorter.vals[j].Interface().(time.Time))
}

func (s reverser) Len() int {
	return s.Sorter.Slice.Len()
}

// Unused--only to satisfy sort.Interface
func (s reverser) Less(i, j int) bool {
	return i < j
}

// Returns a Sorter for a slice which will sort according to the
// items retrieved by getter, in the given ordering.
func New(slice interface{}, getter Getter, ordering Ordering) *Sorter {
	v := reflect.ValueOf(slice)
	return &Sorter{
		Slice:    v,
		Getter:   getter,
		Ordering: ordering,
	}
}

// Sort a slice using a Getter in the order specified by Ordering. getter
// may be nil if sorting a slice of a basic type where identifying a
// parent struct field or slice index isn't necessary, e.g. if sorting an
// []int, []string or []time.Time. A runtime panic will occur if getter is
// not applicable to the given data slice, or if the values retrieved by g
// cannot be compared.
func Sort(slice interface{}, getter Getter, ordering Ordering) {
	New(slice, getter, ordering).Sort()
}

// Sort a slice in ascending order.
func Asc(slice interface{}) {
	New(slice, nil, Ascending).Sort()
}

// Sort a slice in descending order.
func Desc(slice interface{}) {
	New(slice, nil, Descending).Sort()
}

// Sort a slice in case-insensitive ascending order.
func CiAsc(slice interface{}) {
	New(slice, nil, CaseInsensitiveAscending).Sort()
}

// Sort a slice in case-insensitive descending order.
func CiDesc(slice interface{}) {
	New(slice, nil, CaseInsensitiveDescending).Sort()
}

// Sort a slice in ascending order by a field name.
func AscByField(slice interface{}, name string) {
	New(slice, FieldGetter(name), Ascending).Sort()
}

// Sort a slice in descending order by a field name.
func DescByField(slice interface{}, name string) {
	New(slice, FieldGetter(name), Descending).Sort()
}

// Sort a slice in case-insensitive ascending order by a field name.
// (Valid for string types.)
func CiAscByField(slice interface{}, name string) {
	New(slice, FieldGetter(name), CaseInsensitiveAscending).Sort()
}

// Sort a slice in case-insensitive descending order by a field name.
// (Valid for string types.)
func CiDescByField(slice interface{}, name string) {
	New(slice, FieldGetter(name), CaseInsensitiveDescending).Sort()
}

// Sort a slice in ascending order by a list of nested field indices, e.g. 1, 2,
// 3 to sort by the third field from the struct in the second field of the struct
// in the first field of each struct in the slice.
func AscByFieldIndex(slice interface{}, index []int) {
	New(slice, FieldByIndexGetter(index), Ascending).Sort()
}

// Sort a slice in descending order by a list of nested field indices, e.g. 1, 2,
// 3 to sort by the third field from the struct in the second field of the struct
// in the first field of each struct in the slice.
func DescByFieldIndex(slice interface{}, index []int) {
	New(slice, FieldByIndexGetter(index), Descending).Sort()
}

// Sort a slice in case-insensitive ascending order by a list of nested field
// indices, e.g. 1, 2, 3 to sort by the third field from the struct in the
// second field of the struct in the first field of each struct in the slice.
// (Valid for string types.)
func CiAscByFieldIndex(slice interface{}, index []int) {
	New(slice, FieldByIndexGetter(index), CaseInsensitiveAscending).Sort()
}

// Sort a slice in case-insensitive descending order by a list of nested field
// indices, e.g. 1, 2, 3 to sort by the third field from the struct in the
// second field of the struct in the first field of each struct in the slice.
// (Valid for string types.)
func CiDescByFieldIndex(slice interface{}, index []int) {
	New(slice, FieldByIndexGetter(index), CaseInsensitiveDescending).Sort()
}

// Sort a slice in ascending order by an index in a child slice.
func AscByIndex(slice interface{}, index int) {
	New(slice, IndexGetter(index), Ascending).Sort()
}

// Sort a slice in descending order by an index in a child slice.
func DescByIndex(slice interface{}, index int) {
	New(slice, IndexGetter(index), Descending).Sort()
}

// Sort a slice in case-insensitive ascending order by an index in a child
// slice. (Valid for string types.)
func CiAscByIndex(slice interface{}, index int) {
	New(slice, IndexGetter(index), CaseInsensitiveAscending).Sort()
}

// Sort a slice in case-insensitive descending order by an index in a child
// slice. (Valid for string types.)
func CiDescByIndex(slice interface{}, index int) {
	New(slice, IndexGetter(index), CaseInsensitiveDescending).Sort()
}

// Reverse a slice.
func Reverse(slice interface{}) {
	s := reverser{New(slice, nil, 0)}
	if s.Len() < 2 {
		return
	}
	s.itemType = s.Slice.Index(0).Type()
	ReverseInterface(s)
}

// Reverse a type which implements sort.Interface.
func ReverseInterface(s sort.Interface) {
	for i, j := 0, s.Len()-1; i < j; i, j = i+1, j-1 {
		s.Swap(i, j)
	}
}

// Sort a type using its existing sort.Interface, then reverse it. For a
// slice with a a "normal" sort interface (where Less returns true if i
// is less than j), this causes the slice to be sorted in descending order.
func SortReverseInterface(s sort.Interface) {
	sort.Sort(s)
	ReverseInterface(s)
}
