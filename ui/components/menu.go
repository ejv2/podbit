package components

import (
	"fmt"

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
		capped = fmt.Sprintf("%s%s", sel, elem)
		if len(capped) > m.W {
			capped = capped[:m.W]
		}

		m.Win.MovePrint(m.Y+i, m.X, capped)
		c++
	}
}

func (m *Menu) GetSelection() string {
	if len(m.Items) < 1 {
		return ""
	}

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
