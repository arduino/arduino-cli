// This file is part of arduino-cli.
//
// Copyright 2019 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.

package types

import (
	"strconv"
)

// PlatforKeyRewrite keeps track of Key values
type PlatforKeyRewrite struct {
	Key      string
	OldValue string
	NewValue string
}

// PlatforKeysRewrite is a slice of PlatforKeyRewrite
type PlatforKeysRewrite struct {
	Rewrites []PlatforKeyRewrite
}

// Empty returns whether the slice is empty or not
func (p *PlatforKeysRewrite) Empty() bool {
	return len(p.Rewrites) == 0
}

// Prototype FIXMEDOC
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

// CTag represents a ctag
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

// Command is the interface that every command needs to implement
type Command interface {
	Run(ctx *Context) error
}
