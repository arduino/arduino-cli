package schema2docs

import (
	"fmt"
	"io"
	"sort"
	"strings"

	jsonschema "github.com/swaggest/jsonschema-go"
)

type propertiesOrder struct {
	keys         []string
	keysPosition map[string]int
}

func newPropertiesOrder(keys []string, keysPosition map[string]int) *propertiesOrder {
	return &propertiesOrder{
		keys:         keys,
		keysPosition: keysPosition,
	}
}

func (po *propertiesOrder) Len() int {
	return len(po.keys)
}

func (po *propertiesOrder) Less(i, j int) bool {
	iPos := po.keysPosition[po.keys[i]]
	jPos, jOk := po.keysPosition[po.keys[j]]
	if !jOk {
		return true
	}

	return iPos < jPos
}

// Swap swaps the elements with indexes i and j.
func (po propertiesOrder) Swap(i, j int) {
	t := po.keys[i]
	po.keys[i] = po.keys[j]
	po.keys[j] = t
}

type GenerateOptions struct {
	Reader io.Reader
	Writer io.Writer
}

func Generate(opts GenerateOptions) {
	content, err := io.ReadAll(opts.Reader)
	if err != nil {
		panic(err)
	}
	schema := &jsonschema.Schema{}
	schema.UnmarshalJSON(content)
	render(opts.Writer, schema, "", 0)
}

func render(w io.Writer, schema *jsonschema.Schema, name string, level int) {
	parts := []string{}
	if level > 1 {
		parts = append(parts, strings.Repeat("  ", level-1))
	} else {
		parts = append(parts, "")
	}

	if name != "" {
		parts = append(parts, fmt.Sprintf("`%s`", name))
	}
	if schema.Description != nil && *schema.Description != "" {
		parts = append(parts, *schema.Description)
	}
	if len(parts) != 0 {
		fmt.Fprintf(w, "%s\n", strings.Join(parts, " - "))
	}

	if schema.HasType(jsonschema.String) {
		return
	}

	if schema.HasType(jsonschema.Object) {
		keysPos := map[string]int{}
		for propName, propSchema := range schema.Properties {
			if order, ok := propSchema.TypeObject.ExtraProperties["arduino.cc/docs_order"].(float64); ok {
				keysPos[propName] = int(order)
			}
		}
		keys := make([]string, 0, len(schema.Properties))
		for k := range schema.Properties {
			keys = append(keys, k)
		}

		po := newPropertiesOrder(keys, keysPos)
		sort.Sort(po)
		for _, propName := range po.keys {
			propSchema := schema.Properties[propName]
			render(w, propSchema.TypeObject, propName, level+1)
		}
		return
	}

	panic(fmt.Sprintf("Not implemented parser for %s", *schema.Type.SimpleTypes))
}
