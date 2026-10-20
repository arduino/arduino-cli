// This file is part of arduino-cli.
//
// Copyright (C) Arduino s.r.l. and/or its affiliated companies
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html

package builder

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/compilation"
	"github.com/arduino/arduino-cli/internal/arduino/builder/logger"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
)

// lastCommandArgs saves db to dbPath and returns the Arguments of the single
// recorded command. Used to inspect the expanded recipe command line without
// actually executing a compiler.
func lastCommandArgs(t *testing.T, db *compilation.Database, dbPath *paths.Path) []string {
	t.Helper()
	db.SaveToFile()
	raw, err := dbPath.ReadFile()
	require.NoError(t, err)
	var contents []compilation.Command
	require.NoError(t, json.Unmarshal(raw, &contents))
	require.Len(t, contents, 1)
	return contents[0].Arguments
}

func TestPathIsCore(t *testing.T) {
	coreBuildPath := paths.New("build", "core")
	b := &Builder{coreBuildPath: coreBuildPath}

	require.True(t, b.pathIsCore(paths.New("build", "core")))
	require.True(t, b.pathIsCore(paths.New("build", "core", "core.a")))
	require.True(t, b.pathIsCore(paths.New("build", "core", "variant", "objs.a")))
	require.False(t, b.pathIsCore(paths.New("build", "sketch")))
	require.False(t, b.pathIsCore(paths.New("build", "libraries", "Bridge.a")))
	require.False(t, b.pathIsCore(paths.New("build", "core.a")))
}

func TestCompileFileWithRecipeExcludesLibraryDiscoveryFlagsForCore(t *testing.T) {
	tmpDir := paths.New(t.TempDir())
	sourceFile := tmpDir.Join("sketch.ino.cpp")
	require.NoError(t, sourceFile.WriteFile([]byte("// test\n")))

	coreBuildPath := tmpDir.Join("core-build")
	otherBuildPath := tmpDir.Join("other-build")

	newTestBuilder := func(dbPath *paths.Path) (*Builder, *compilation.Database) {
		props := properties.NewMap()
		props.Set("recipe.o.pattern", "echo {build.library_discovery_flags}")
		props.Set("build.library_discovery_flags", "-DFOUND_TESTLIB_LIB=0x010203")
		db := compilation.NewDatabase(dbPath)
		return &Builder{
			buildProperties:               props,
			coreBuildPath:                 coreBuildPath,
			logger:                        logger.New(io.Discard, io.Discard, logger.VerbosityQuiet, "none"),
			onlyUpdateCompilationDatabase: true,
			compilationDatabase:           db,
		}, db
	}

	t.Run("CoreBuildPathExcludesFlag", func(t *testing.T) {
		dbPath := tmpDir.Join("core_compile_commands.json")
		b, db := newTestBuilder(dbPath)
		_, err := b.compileFileWithRecipe(tmpDir, sourceFile, coreBuildPath, nil, "recipe.o.pattern")
		require.NoError(t, err)
		args := lastCommandArgs(t, db, dbPath)
		require.Contains(t, args, "{build.library_discovery_flags}")
		require.NotContains(t, args, "-DFOUND_TESTLIB_LIB=0x010203")
	})

	t.Run("NonCoreBuildPathKeepsFlag", func(t *testing.T) {
		dbPath := tmpDir.Join("other_compile_commands.json")
		b, db := newTestBuilder(dbPath)
		_, err := b.compileFileWithRecipe(tmpDir, sourceFile, otherBuildPath, nil, "recipe.o.pattern")
		require.NoError(t, err)
		args := lastCommandArgs(t, db, dbPath)
		require.Contains(t, args, "-DFOUND_TESTLIB_LIB=0x010203")
	})
}
