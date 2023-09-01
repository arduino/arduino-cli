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

package table

import (
	"fmt"
	"math"
)

// ColumnWidthMode is used to configure columns type
type ColumnWidthMode int

const (
	// Minimum FIXMEDOC
	Minimum ColumnWidthMode = iota
	// Average FIXMEDOC
	Average
)

// Table represent a table that can be printed on a terminal
type Table struct {
	hasHeader        bool
	columnsCount     int
	columnsWidthMode []ColumnWidthMode
	rows             []*tableRow
}

type tableRow struct {
	cells []Cell
}

// New creates an empty table
func New() *Table {
	return &Table{
		rows: []*tableRow{},
	}
}

// SetColumnWidthMode FIXMEDOC
func (t *Table) SetColumnWidthMode(x int, mode ColumnWidthMode) {
	for len(t.columnsWidthMode) <= x {
		t.columnsWidthMode = append(t.columnsWidthMode, Minimum)
	}
	t.columnsWidthMode[x] = mode
}

// SetHeader FIXMEDOC
func (t *Table) SetHeader(columns ...interface{}) {
	row := t.makeTableRow(columns...)
	if t.hasHeader {
		t.rows[0] = row
	} else {
		t.rows = append([]*tableRow{row}, t.rows...)
		t.hasHeader = true
	}
}

// AddRow FIXMEDOC
func (t *Table) AddRow(columns ...interface{}) {
	row := t.makeTableRow(columns...)
	t.rows = append(t.rows, row)
}

// Render FIXMEDOC
func (t *Table) Render() string {
	// find max width for each row
	average := make([]int, t.columnsCount)
	widths := make([]int, t.columnsCount)
	count := make([]int, t.columnsCount)
	minimum := make([]int, t.columnsCount)
	for _, row := range t.rows {
		for x, cell := range row.cells {
			l := cell.Len()
			if l == 0 {
				continue
			}
			count[x]++
			average[x] += l
			if cell.Len() > widths[x] {
				widths[x] = l
			}
		}
	}
	for x := range average {
		if count[x] > 0 {
			average[x] = average[x] / count[x]
		}
	}
	// table headers will dictate the absolute min width
	for x := range minimum {
		if t.hasHeader {
			minimum[x] = t.rows[0].cells[x].Len()
		} else {
			minimum[x] = 1
		}
	}

	variance := make([]int, t.columnsCount)
	for _, row := range t.rows {
		for x, cell := range row.cells {
			l := cell.Len()
			if l == 0 {
				continue
			}
			d := l - average[x]
			variance[x] += d * d
		}
	}
	for x := range variance {
		if count[x] > 0 {
			variance[x] = int(math.Sqrt(float64(variance[x] / count[x])))
		}
	}

	res := ""
	for _, row := range t.rows {
		for x, cell := range row.cells {
			selectedWidth := widths[x]
			if x < len(t.columnsWidthMode) {
				switch t.columnsWidthMode[x] {
				case Minimum:
					selectedWidth = widths[x]
				case Average:
					selectedWidth = average[x] + variance[x]*3
				}
			}
			if selectedWidth < minimum[x] {
				selectedWidth = minimum[x]
			}
			if x > 0 {
				line += " "
			}
			res += cell.Pad(selectedWidth)
		}
		res += "\n"
	}
	return res
}

func makeCell(format string, args ...interface{}) *Cell {
	cleanArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if text, ok := arg.(*Cell); ok {
			cleanArgs[i], args[i] = text.clean, text.raw
		} else {
			cleanArgs[i] = args[i]
		}
	}

	return &Cell{
		clean: fmt.Sprintf(format, cleanArgs...),
		raw:   fmt.Sprintf(format, args...),
	}
}

func (t *Table) makeTableRow(columns ...interface{}) *tableRow {
	columnsCount := len(columns)
	if t.columnsCount < columnsCount {
		t.columnsCount = columnsCount
	}
	cells := make([]Cell, columnsCount)
	for i, col := range columns {
		switch elem := col.(type) {
		case *Cell:
			cells[i] = *elem
		case string:
			cells[i] = *makeCell("%s", elem)
		case fmt.Stringer:
			cells[i] = *makeCell("%s", elem.String())
		default:
			panic(fmt.Sprintf("invalid column argument type: %t", col))
		}
	}
	return &tableRow{cells: cells}
}
