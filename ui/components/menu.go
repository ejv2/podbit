package components

import "github.com/rthornton128/goncurses"

type Menu struct {
	W, H  int
	X, Y  int
	Items []string
	Win  goncurses.Window

	scroll int
	sel int
}

func (m Menu) Render() {
	items := m.Items[m.scroll:]
	for i, elem := range items {
		var sel string
		if i == m.sel {
			sel = ">>"
		}

		var capped string
		if len(elem) > m.W {
			capped = elem[:m.W]
		} else {
			capped = elem
		}

		m.Win.MovePrintf(m.Y+i, m.X, "%s%s", sel, capped)
	}
}
