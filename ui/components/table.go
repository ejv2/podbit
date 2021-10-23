package components

import (
	"github.com/rthornton128/goncurses"
)

// A Column is a vertical section of a table, defined by a
// pair of both a "width" and a "label".
//
// The width is the fraction of the table width the column
// should take up (defined between zero and one). Table's
// render method will panic if any column has a "width"
// of less than zero or greater than one, or if an impossible
// combination of widths was requested.
//
// The width is the label which will be displayed at the top of the table
// The color is the ncurses color to activate when selected
type Column struct {
	Label string
	Width float64
	Color int16
}

// Table represents a vertical, headed table structure, useful
// for displaying a slice of structs
type Table struct {
	X, Y int
	W, H int
	Win  *goncurses.Window

	// See column struct for docs
	Columns []Column
	// Each sub-slice represents each column entry (eg, 0 = first column)
	Items [][]string

	scroll int
	sel    int

	prevw, prevh int
}

func (t *Table) Render() {
	items := t.Items[t.scroll:]
	if t.prevw != t.W || t.prevh != t.H {
		t.scroll = 0
	}
	t.prevw, t.prevh = t.W, t.H

	t.Win.AttrOn(goncurses.A_BOLD)
	t.Win.HLine(t.Y+1, t.X, goncurses.ACS_HLINE, t.W)
	t.Win.AttrOff(goncurses.A_BOLD)

	off := 0
	for i, elem := range t.Columns {
		colw := int(float64(t.W) * elem.Width)
		if off > t.W || colw > t.W {
			panic("invalid table header: impossible column width")
		}

		if len(elem.Label) > colw {
			elem.Label = elem.Label[:colw]
		}

		t.Win.AttrOn(goncurses.A_BOLD)
		t.Win.MovePrint(t.Y, t.X+off, elem.Label)
		t.Win.AttrOff(goncurses.A_BOLD)

		c := t.scroll
		for j, entry := range items {
			if i >= len(entry) {
				panic("invalid table entry: missing fields")
			}

			capped := entry[i]
			sel := c == t.sel

			if len(capped) > colw {
				// Trim to fit row
				capped = capped[:colw]
			} else {
				// Pad out to fill row
				for i := len(capped); i <= colw; i++ {
					capped += " "
				}
			}

			if sel {
				t.Win.ColorOn(elem.Color)
			}

			t.Win.MovePrint(t.Y+j+2, t.X+off, capped)

			if sel {
				t.Win.ColorOff(elem.Color)
			}

			c++
		}

		off += colw
	}
}

// GetSelection returns the text of the currently
// selected menu element. If there are no items selected,
// GetSelection returns an empty slice.
func (m *Table) GetSelection() []string {
	if len(m.Items) < 1 {
		return []string{}
	}

	return m.Items[m.sel]
}

// ChangeSelection changes the selection to the index specified.
// If index is out of range, no action is taken.
func (m *Table) ChangeSelection(index int) {
	if index >= len(m.Items) || index < 0 {
		return
	}

	m.sel = index

	scrollAt := m.H + m.scroll + 1
	underscrollAt := m.scroll - 1
	if m.sel == scrollAt {
		m.scroll++
	} else if m.sel == underscrollAt {
		m.scroll--
	}
}

// MoveSelection changes the selected item relative to the current
// position. If the new selection would be out of range, no action
// is taken.
func (m *Table) MoveSelection(offset int) {
	off := m.sel + offset
	m.ChangeSelection(off)
}
