/*
Sortutil is a Go library which lets you sort a slice without implementing a
sort.Interface, and in different orderings: ascending, descending, or
case-insensitive ascending or descending (for slices of strings.)

Additionally, Sortutil lets you sort a slice of a custom struct by a given
struct field or index--for example, you can sort a []MyStruct by the structs'
"Name" fields, or a [][]int by the second index of each nested slice, similar
to using sorted(key=operator.itemgetter/attrgetter) in Python.
*/
package sortutil
