package i18n

import (
	"time"

	"github.com/GeertJohan/go.rice/embedded"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    ".gitkeep",
		FileModTime: time.Unix(1587752852, 0),

		Content: string(""),
	}
	file3 := &embedded.EmbeddedFile{
		Filename:    "en.po",
		FileModTime: time.Unix(1587758440, 0),

		Content: string("msgid \"\"\nmsgstr \"\"\n\n"),
	}
	file4 := &embedded.EmbeddedFile{
		Filename:    "pt_BR.po",
		FileModTime: time.Unix(1587758441, 0),

		Content: string("msgid \"\"\nmsgstr \"\"\n\n"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1587758441, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // ".gitkeep"
			file3, // "en.po"
			file4, // "pt_BR.po"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`./data`, &embedded.EmbeddedBox{
		Name: `./data`,
		Time: time.Unix(1587758441, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			".gitkeep": file2,
			"en.po":    file3,
			"pt_BR.po": file4,
		},
	})
}
