package po

// Merge merges two message catalogs, preserving only keys present in source
func Merge(source MessageCatalog, destination MessageCatalog) MessageCatalog {
	catalog := MessageCatalog{}
	for _, k := range source.SortedKeys() {
		if k == "" {
			sourceMessage := source.Messages[k]
			destinationMessage := destination.Messages[k]

			if destinationMessage != nil {
				catalog.AddMessage(k, *destinationMessage)
			} else {
				catalog.AddMessage(k, *sourceMessage)
			}

			continue
		}

		if destination.Messages[k] != nil {
			catalog.Add(k, destination.Messages[k].Value, source.Messages[k].Comments)
		} else {
			catalog.Add(k, "", source.Messages[k].Comments)
		}
	}
	return catalog
}
