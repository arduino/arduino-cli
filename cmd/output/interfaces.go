package output

// ProcessResults identifies a set of generic process results.
type ProcessResults interface {
	// Results returns a set of generic results, to allow them to be modified externally.
	Results() *[]ProcessResult
}
