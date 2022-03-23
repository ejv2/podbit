package components

import (
	"github.com/ethanv2/podbit/colors"
	"github.com/rthornton128/goncurses"
)

// Menu represents a vertical panel menu of cellular entries
// taking up an entire row from X to W
//
// Menu handles focus, scrolling and the managing of elements
// at each render.
//
// This component is not thread safe and should only be modified
// directly one a single thread
type Menu struct {
	W, H  int
	X, Y  int
	Items []string
	Win   goncurses.Window

	Selected bool

	scroll int
	sel    int

	prevw, prevh int
}

// Render immediately renders the menu to the specified
// fields X, Y, H and W.
func (m *Menu) Render() {
	// Reset scroll if invalid or not selected
	if len(m.Items) <= m.scroll || !m.Selected {
		m.scroll = 0
		m.sel = 0
	}

	items := m.Items[m.scroll:]
	c := m.scroll

	if m.prevw != m.W || m.prevh != m.H {
		m.scroll = 0
		m.sel = 0
	}
	m.prevw, m.prevh = m.W, m.H

	for i, elem := range items {
		if c == m.sel && m.Selected {
			m.Win.ColorOn(colors.BackgroundBlue)
		} else if i > m.H {
			break
		}

		var capped string
		capped = elem
		if len(capped) > m.W {
			capped = capped[len(capped)-m.W:]
			capped = "<" + capped
		} else {
			// Pad out to fill row
			for i := len(capped); i <= m.W; i++ {
				capped += " "
			}
		}

		m.Win.MovePrint(m.Y+i, m.X, capped)
		c++

		m.Win.ColorOff(colors.BackgroundBlue)
	}
}

// GetSelection returns the text of the currently
// selected menu element. If there are no items selected,
// GetSelection returns an empty string.
func (m *Menu) GetSelection() string {
	if len(m.Items) < 1 {
		return ""
	}

	return m.Items[m.sel]
}

// ChangeSelection changes the selection to the index specified.
// If index is out of range, no action is taken.
func (m *Menu) ChangeSelection(index int) {
	if index >= len(m.Items) || index < 0 {
		return
	}

	m.sel = index

	scrollAt := m.H + m.scroll + 1
	underscrollAt := m.scroll - 1
	if m.sel >= scrollAt {
		m.scroll += (m.sel-scrollAt)+1
	} else if m.sel <= underscrollAt {
		m.scroll -= (m.sel-underscrollAt)+1
	}
}

// MoveSelection changes the selected item relative to the current
// position. If the new selection would be out of range, no action
// is taken.
func (m *Menu) MoveSelection(offset int) {
	off := m.sel + offset
	m.ChangeSelection(off)
}
