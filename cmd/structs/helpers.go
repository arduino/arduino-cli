package structs

// LibResultsFromMap returns a LibProcessResults struct from a specified map of results.
func LibResultsFromMap(resMap map[string]string) LibProcessResults {
	results := LibProcessResults{
		Libraries: make([]libProcessResult, len(resMap)),
	}
	i := 0
	for libName, libResult := range resMap {
		results.Libraries[i] = libProcessResult{
			LibraryName: libName,
			Result:      libResult,
		}
		i++
	}
	return results
}
