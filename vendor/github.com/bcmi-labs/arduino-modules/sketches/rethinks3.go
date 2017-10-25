package sketches

import (
	"path/filepath"

	"github.com/bcmi-labs/arduino-modules/fs"
	"github.com/fluxio/multierror"
	"github.com/juju/errors"
	r "gopkg.in/gorethink/gorethink.v3"
)

type filesystem interface {
	fs.Reader
	fs.Writer
	fs.Remover
}

// RethinkFS saves the sketches info on rethinkdb and the file data on s3
type RethinkFS struct {
	DB    r.QueryExecutor
	Table string
	FS    filesystem
}

// Read retrieves the sketch from rethinkdb and the file contents from the filesystem
func (c *RethinkFS) Read(id string) (*Sketch, error) {
	query := r.Table(c.Table).Get(id)
	cursor, err := query.Run(c.DB)
	if err != nil {
		return nil, errors.Annotatef(err, "Run query %s", query)
	}

	if cursor.IsNil() {
		return nil, errors.NotFoundf("with id %s", id)
	}

	var sketch Sketch
	err = cursor.One(&sketch)
	if err != nil {
		return nil, errors.Annotatef(err, "Retrieve from cursor %+v", cursor)
	}

	// reading errors are returned as a notfound multierror
	var errs multierror.Accumulator

	sketch.Ino.Read(c.FS)
	if err != nil {
		errs.Push(err)
	}

	for i := range sketch.Files {
		if filepath.Ext(sketch.Files[i].Name) != ".ino" && filepath.Ext(sketch.Files[i].Name) != ".h" {
			continue
		}
		err := sketch.Files[i].Read(c.FS)
		if err != nil {
			errs.Push(err)
		}
	}

	if errs.Error() != nil {
		return &sketch, errors.NewNotFound(errs.Error(), "with id "+id)
	}
	return &sketch, nil

}

// Search retrieves from the storage the sketches that match the fields.
// It uses rethinkdb filtering to achieve it. The files in the sketches are
// without data.
func (c *RethinkFS) Search(fields map[string]string, skip, limit int) ([]Sketch, error) {
	query := r.Table(c.Table)

	if value, ok := fields["owner"]; ok {
		query = query.GetAllByIndex("owner", value)
	}

	for field, value := range fields {
		if field == "owner" {
			continue
		}
		query = query.Filter(r.Row.Field(field).Eq(value))
	}

	limit = skip + limit

	if limit != 0 {
		query = query.Slice(skip, limit)
	}

	cursor, err := query.Run(c.DB)
	if err != nil {
		return nil, errors.Annotatef(err, "Run query %s", query)
	}

	list := []Sketch{}
	err = cursor.All(&list)
	if err != nil {
		return nil, errors.Annotatef(err, "Retrieve from cursor %+v", cursor)
	}

	return list, nil
}

// Write persists the sketch on rethinkdb, saving the file contents on the filesystem
func (c *RethinkFS) Write(sketch *Sketch) error {
	old, _ := c.Read(sketch.ID)

	// save on fs
	err := c.saveFS(sketch, old)
	if err != nil {
		return err
	}

	// save on db
	options := r.InsertOpts{Conflict: "replace"}
	query := r.Table(c.Table).Insert(sketch, options)
	_, err = query.RunWrite(c.DB)
	if err != nil {
		return errors.Annotatef(err, "Runwrite query %s", query)
	}
	return nil
}

func (c *RethinkFS) saveFS(sketch, old *Sketch) error {
	var oldFile *fs.File
	if old != nil {
		oldFile = &old.Ino
	}
	err := c.saveFile(&sketch.Ino, oldFile)
	if err != nil {
		return err
	}

	for i := range sketch.Files {
		oldFile = nil
		if old != nil {
			oldFile = old.GetFile(sketch.Files[i].Name)
		}
		err := c.saveFile(&sketch.Files[i], oldFile)
		if err != nil {
			return err
		}
	}
	if old != nil {
		return c.cleanFS(sketch, old)
	}
	return nil
}

func (c *RethinkFS) saveFile(file, old *fs.File) error {
	rename := false
	if old != nil && file.Path != old.Path {
		rename = true
	}

	if rename && file.Data == nil {
		err := old.Read(c.FS)
		if err == nil {
			file.Data = old.Data
		}
	}

	if file.Data != nil {
		err := file.Write(c.FS, 0777)
		if err != nil {
			return err
		}
		if rename {
			return c.FS.Remove(old.Path)
		}
		return nil
	}

	return nil
}

func (c *RethinkFS) cleanFS(sketch, old *Sketch) error {
	for i := range old.Files {
		exists := sketch.GetFile(old.Files[i].Name)
		if exists == nil {
			err := c.FS.Remove(old.Files[i].Path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Delete removes the sketch both from rethinkdb and the filesystem
func (c *RethinkFS) Delete(id string) error {
	sketch, err := c.Read(id)
	if sketch == nil {
		return err
	}

	err = c.FS.Remove(sketch.Ino.Path)
	if err != nil && !errors.IsNotFound(err) {
		return errors.Annotatef(err, "remove file %s", sketch.Ino.Path)
	}
	for _, file := range sketch.Files {
		err = c.FS.Remove(file.Path)
		if err != nil && !errors.IsNotFound(err) {
			return errors.Annotatef(err, "remove file %s", file.Path)
		}
	}

	query := r.Table(c.Table).Get(sketch.ID).Delete()
	err = query.Exec(c.DB)
	if err != nil {
		return errors.Annotatef(err, "Exec query %s", query)
	}
	return nil
}
