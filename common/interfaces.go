package common

// Index represents a generic index parsed from an index file.
type Index interface {
	CreateStatusContext() (StatusContext, error) // CreateStatusContext creates a status context with this index data.
}

// StatusContext represents a generic status context, created from an Index.
type StatusContext interface {
	Names() []string               // Names Returns an array with all the names of the items.
	Items() map[string]interface{} // Items Returns a map of all items with their names.
}
