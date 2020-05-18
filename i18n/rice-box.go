package i18n

import (
	"time"

	"github.com/GeertJohan/go.rice/embedded"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    ".gitkeep",
		FileModTime: time.Unix(1589335267, 0),

		Content: string(""),
	}
	file3 := &embedded.EmbeddedFile{
		Filename:    "en.po",
		FileModTime: time.Unix(1589846460, 0),

		Content: string("msgid \"\"\nmsgstr \"\"\n\n#: ./cli/usage.go:31\nmsgid \"Additional help topics:\"\nmsgstr \"Additional help topics:\"\n\n#: ./cli/usage.go:26\nmsgid \"Aliases:\"\nmsgstr \"Aliases:\"\n\n#: ./cli/usage.go:28\nmsgid \"Available Commands:\"\nmsgstr \"Available Commands:\"\n\n#: ./cli/usage.go:27\nmsgid \"Examples:\"\nmsgstr \"Examples:\"\n\n#: ./cli/usage.go:29\nmsgid \"Flags:\"\nmsgstr \"Flags:\"\n\n#: ./cli/usage.go:30\nmsgid \"Global Flags:\"\nmsgstr \"Global Flags:\"\n\n#: ./cli/usage.go:25\nmsgid \"Usage:\"\nmsgstr \"Usage:\"\n\n#: ./cli/usage.go:32\nmsgid \"Use %s for more information about a command.\"\nmsgstr \"Use %s for more information about a command.\"\n\n"),
	}
	file4 := &embedded.EmbeddedFile{
		Filename:    "pt_BR.po",
		FileModTime: time.Unix(1589846461, 0),

		Content: string("msgid \"\"\nmsgstr \"\"\n\n#: ./cli/usage.go:31\nmsgid \"Additional help topics:\"\nmsgstr \"\"\n\n#: ./cli/usage.go:26\nmsgid \"Aliases:\"\nmsgstr \"\"\n\n#: ./cli/usage.go:28\nmsgid \"Available Commands:\"\nmsgstr \"\"\n\n#: ./cli/usage.go:27\nmsgid \"Examples:\"\nmsgstr \"\"\n\n#: ./cli/usage.go:29\nmsgid \"Flags:\"\nmsgstr \"\"\n\n#: ./cli/usage.go:30\nmsgid \"Global Flags:\"\nmsgstr \"\"\n\n#: ./cli/usage.go:25\nmsgid \"Usage:\"\nmsgstr \"\"\n\n#: ./cli/usage.go:32\nmsgid \"Use %s for more information about a command.\"\nmsgstr \"\"\n\n"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1589846461, 0),
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
		Time: time.Unix(1589846461, 0),
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
