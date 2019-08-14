/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package output

import (
	"fmt"
	"math"
)

// Table FIXMEDOC
type Table struct {
	hasHeader        bool
	columnsCount     int
	columnsWidthMode []TableColumnWidthMode
	rows             []*TableRow
}

// TableRow FIXMEDOC
type TableRow struct {
	cells []TextBox
}

// NewTable FIXMEDOC
func NewTable() *Table {
	return &Table{
		rows: []*TableRow{},
	}
}

// TableColumnWidthMode FIXMEDOC
type TableColumnWidthMode int

const (
	// Minimum FIXMEDOC
	Minimum TableColumnWidthMode = iota
	// Average FIXMEDOC
	Average
)

// SetColumnWidthMode FIXMEDOC
func (t *Table) SetColumnWidthMode(x int, mode TableColumnWidthMode) {
	for len(t.columnsWidthMode) <= x {
		t.columnsWidthMode = append(t.columnsWidthMode, Minimum)
	}
	t.columnsWidthMode[x] = mode
}

func (t *Table) makeTableRow(columns ...interface{}) *TableRow {
	columnsCount := len(columns)
	if t.columnsCount < columnsCount {
		t.columnsCount = columnsCount
	}
	cells := make([]TextBox, columnsCount)
	for i, col := range columns {
		switch text := col.(type) {
		case TextBox:
			cells[i] = text
		case string:
			cells[i] = sprintf("%s", text)
		case fmt.Stringer:
			cells[i] = sprintf("%s", text.String())
		default:
			panic(fmt.Sprintf("invalid column argument type: %t", col))
		}
	}
	return &TableRow{cells: cells}
}

// SetHeader FIXMEDOC
func (t *Table) SetHeader(columns ...interface{}) {
	row := t.makeTableRow(columns...)
	if t.hasHeader {
		t.rows[0] = row
	} else {
		t.rows = append([]*TableRow{row}, t.rows...)
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
		separator := ""
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
			res += separator
			res += cell.Pad(selectedWidth)
			separator = " "
		}
		res += "\n"
	}
	return res
}
