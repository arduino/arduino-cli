// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package cli_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestNoDirectOutputToStdOut(t *testing.T) {
	dirs, err := paths.New(".").ReadDirRecursiveFiltered(
		paths.FilterOutNames("testdata"), // skip all testdata folders
		paths.AndFilter(
			paths.FilterDirectories(),        // analyze only packages
			paths.FilterOutNames("feedback"), // skip feedback package
		))
	require.NoError(t, err)
	dirs.Add(paths.New("."))

	for _, dir := range dirs {
		testDir(t, dir)
	}
}

func testDir(t *testing.T, dir *paths.Path) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir.String(), nil, parser.ParseComments)
	require.NoError(t, err)

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			// Do not analyze test files
			if strings.HasSuffix(expr(file.Name), "_test") {
				continue
			}

			ast.Inspect(file, func(n ast.Node) bool {
				return inspect(t, fset, n)
			})
		}
	}
}

func inspect(t *testing.T, fset *token.FileSet, node ast.Node) bool {
	switch n := node.(type) {
	case *ast.CallExpr:
		name := expr(n.Fun)
		if strings.HasPrefix(name, "fmt.P") {
			fmt.Printf("%s: function `%s` should not be used in this package (use `feedback.*` instead)\n", fset.Position(n.Pos()), name)
			t.Fail()
		}
	case *ast.SelectorExpr:
		wanted := map[string]string{
			"os.Stdout": "%s: object `%s` should not be used in this package (use `feedback.*` instead)\n",
			"os.Stderr": "%s: object `%s` should not be used in this package (use `feedback.*` instead)\n",
			"os.Stdin":  "%s: object `%s` should not be used in this package (use `feedback.*` instead)\n",
			"os.Exit":   "%s: function `%s` should not be used in this package (use `return` or `feedback.FatalError` instead)\n",
		}
		name := expr(n)
		if msg, banned := wanted[name]; banned {
			fmt.Printf(msg, fset.Position(n.Pos()), name)
			t.Fail()
		}
	}
	return true
}

// expr returns the string representation of an expression, it doesn't expand function arguments or array index.
func expr(_e ast.Expr) string {
	switch e := _e.(type) {
	case *ast.ArrayType:
		return "[...]" + expr(e.Elt)
	case *ast.CallExpr:
		return expr(e.Fun) + "(...)"
	case *ast.FuncLit:
		return "func(...) {...}"
	case *ast.SelectorExpr:
		return expr(e.X) + "." + e.Sel.String()
	case *ast.IndexExpr:
		return expr(e.X) + "[...]"
	case *ast.Ident:
		return e.String()
	default:
		msg := fmt.Sprintf("UNKWOWN: %T", e)
		panic(msg)
	}
}
