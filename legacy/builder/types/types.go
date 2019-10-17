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

package types

import (
	"strconv"

	"github.com/arduino/arduino-cli/arduino/libraries"
	paths "github.com/arduino/go-paths-helper"
)

type SketchFile struct {
	Name *paths.Path
}

type SketchFileSortByName []SketchFile

func (s SketchFileSortByName) Len() int {
	return len(s)
}

func (s SketchFileSortByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SketchFileSortByName) Less(i, j int) bool {
	return s[i].Name.String() < s[j].Name.String()
}

type PlatforKeysRewrite struct {
	Rewrites []PlatforKeyRewrite
}

func (p *PlatforKeysRewrite) Empty() bool {
	return len(p.Rewrites) == 0
}

type PlatforKeyRewrite struct {
	Key      string
	OldValue string
	NewValue string
}

type Prototype struct {
	FunctionName string
	File         string
	Prototype    string
	Modifiers    string
	Line         int
}

func (proto *Prototype) String() string {
	return proto.Modifiers + " " + proto.Prototype + " @ " + strconv.Itoa(proto.Line)
}

type LibraryResolutionResult struct {
	Library          *libraries.Library
	NotUsedLibraries []*libraries.Library
}

type CTag struct {
	FunctionName string
	Kind         string
	Line         int
	Code         string
	Class        string
	Struct       string
	Namespace    string
	Filename     string
	Typeref      string
	SkipMe       bool
	Signature    string

	Prototype          string
	PrototypeModifiers string
}

type Command interface {
	Run(ctx *Context) error
}
