package components

import (
	"github.com/rthornton128/goncurses"
)

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

func (m *Menu) Render() {
	items := m.Items[m.scroll:]
	c := m.scroll

	if m.prevw != m.W || m.prevh != m.H {
		m.scroll = 0
	}
	m.prevw, m.prevh = m.W, m.H

	for i, elem := range items {
		var sel string
		if c == m.sel && m.Selected {
			sel = ">>"
		}

		var capped string
		if len(elem) > m.W {
			capped = elem[:m.W]
		} else {
			capped = elem
		}

		m.Win.MovePrintf(m.Y+i, m.X, "%s%s", sel, capped)
		c++
	}
}

func (m *Menu) GetSelection() string {
	return m.Items[m.sel]
}

func (m *Menu) ChangeSelection(index int) {
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

func (m *Menu) MoveSelection(offset int) {
	off := m.sel + offset
	m.ChangeSelection(off)
}
