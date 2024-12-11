// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
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

package ast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"

	"github.com/arduino/arduino-cli/internal/locales/cmd/po"
)

// GenerateCatalog generates the i18n message catalog for the go source files
func GenerateCatalog(files []string) po.MessageCatalog {
	fset := token.NewFileSet()
	catalog := po.MessageCatalog{}

	for _, file := range files {
		node, err := parser.ParseFile(fset, file, nil, parser.AllErrors)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		doFile(fset, node, &catalog)
	}

	catalog.Add("", "", nil)

	return catalog
}

func doFile(fset *token.FileSet, file *ast.File, catalog *po.MessageCatalog) {
	ast.Inspect(file, func(node ast.Node) bool {
		funcCall, ok := node.(*ast.CallExpr)

		if !ok {
			return true
		}

		if functionName(funcCall) != "i18n.Tr" && functionName(funcCall) != "tr" {
			return true
		}

		pos := fset.Position(funcCall.Pos())
		firstArg, ok := funcCall.Args[0].(*ast.BasicLit)
		if !ok {
			fmt.Fprintf(os.Stderr, "%s:%d\n", pos.Filename, pos.Line)
			fmt.Fprintln(os.Stderr, "argument to i18n.Tr must be a literal string")
			return true
		}

		msg, err := strconv.Unquote(firstArg.Value)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s:%d\n", pos.Filename, pos.Line)
			fmt.Fprintln(os.Stderr, err.Error())
			return true
		}

		catalog.Add(msg, msg, []string{fmt.Sprintf("#: %s:%d", filepath.ToSlash(pos.Filename), pos.Line)})

		return true
	})
}

func functionName(callExpr *ast.CallExpr) string {

	if iden, ok := callExpr.Fun.(*ast.Ident); ok {
		return iden.Name
	}

	if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
		if iden, ok := sel.X.(*ast.Ident); ok {
			return iden.Name + "." + sel.Sel.Name
		}
	}

	return ""
}
