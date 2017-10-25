package sketches

import (
	"reflect"
	"sort"
	"strings"

	"github.com/juju/errors"
)

// Mock is a test helper that allows to persists sketches in memory while
// respecting the contract of the sketches interfaces
type Mock struct {
	Sketches map[string]Sketch
}

// Read retrieves a sketch based on the id
// A notfound error is returned if the sketch wasn't found
func (c *Mock) Read(id string) (*Sketch, error) {
	sketch, ok := c.Sketches[id]
	if !ok {
		return nil, errors.NotFoundf("with id %s", id)
	}
	return &sketch, nil
}

// Search retrieves from the storage the sketches that match the fields.
func (c *Mock) Search(fields map[string]string, skip, limit int) ([]Sketch, error) {
	list := []Sketch{}

	for _, sketch := range c.Sketches {
		add := true
		for field, value := range fields {
			r := reflect.ValueOf(sketch)
			f := reflect.Indirect(r).FieldByName(strings.Title(field))
			if f.String() != value {
				add = false
			}
		}

		if add {
			list = append(list, sketch)
		}
	}

	if skip > len(list) {
		return []Sketch{}, nil
	}

	limit = skip + limit
	if limit == 0 || limit > len(list) {
		limit = len(list)
	}

	sort.Sort(ByID(list))

	return list[skip:limit], nil
}

// Write persists the sketch in memory
func (c *Mock) Write(sketch *Sketch) error {
	old, ok := c.Sketches[sketch.ID]

	if ok {
		if sketch.Ino.Data == nil {
			sketch.Ino.Data = old.Ino.Data
		}
		for i := range sketch.Files {
			if sketch.Files[i].Data == nil {
				oldFile := old.GetFile(sketch.Files[i].Name)
				if oldFile == nil {
					continue
				}
				sketch.Files[i].Data = oldFile.Data
			}
		}
	}

	c.Sketches[sketch.ID] = *sketch
	return nil
}

// Delete removes the sketch from memory
func (c *Mock) Delete(id string) error {
	_, ok := c.Sketches[id]
	if !ok {
		return errors.NotFoundf("with id %s", id)
	}
	delete(c.Sketches, id)
	return nil
}
