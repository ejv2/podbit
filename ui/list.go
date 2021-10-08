// Lists your configured/detected podcasts and available episodes
package ui

import (
	"fmt"

	"github.com/ethanv2/podbit/data"
	"github.com/ethanv2/podbit/ui/components"

	"github.com/rthornton128/goncurses"
)

type List struct {
	men [2]components.Menu

	menSel int
}

func (l *List) Name() string {
	return "Podcasts"
}

func (l *List) renderPodcasts(x, y int) {
	l.men[0].X = x
	l.men[0].Y = y

	l.men[0].W, l.men[0].H = (w/2)-1, (h - 5)
	l.men[0].Win = *root

	l.men[0].Items = make([]string, len(data.Q.Items))

	seen := make(map[string]bool)
	for i := range l.men[0].Items {
		name := data.DB.GetFriendlyName(data.Q.Items[i].Url)

		if !seen[name] {
			l.men[0].Items[i] = name
			seen[name] = true
		}
	}

	for x, elem := range l.men[0].Items {
		if elem == "" {
			l.men[0].Items = append(l.men[0].Items[:x], l.men[0].Items[x+1:]...)
		}
	}

	l.men[0].Selected = (l.menSel == 0)

	l.men[0].Render()
}

func (l *List) renderEpisodes(x, y int) {
	l.men[1].X = x
	l.men[1].Y = y

	l.men[1].W, l.men[1].H = (w/2)-1, (h - 5)
	l.men[1].Win = *root

	l.men[1].Items = l.men[1].Items[:0]

	for _, elem := range data.Q.Items {
		if data.DB.GetFriendlyName(elem.Url) == l.men[0].GetSelection() {
			l.men[1].Items = append(l.men[1].Items, fmt.Sprintf("%s", elem.Url))
		}
	}

	l.men[1].Render()
}

func (l *List) Render(x, y int) {
	l.renderPodcasts(x, y)
	root.VLine(y, w/2, goncurses.ACS_VLINE, h-2-y)
	l.renderEpisodes(w/2+1, y)
}

func (l *List) Input(c rune) {
	switch c {
	case 'j':
		l.men[l.menSel].MoveSelection(1)
	case 'k':
		l.men[l.menSel].MoveSelection(-1)
	case 'h':
		l.MoveSelection(-1)
	case 'l':
		l.MoveSelection(1)
	}
}

func (l *List) ChangeSelection(index int) {
	if index >= len(l.men) || index < 0 {
		return
	}

	l.menSel = index
}

func (l *List) MoveSelection(direction int) {
	if direction == 0 {
		return
	}

	off := l.menSel + direction
	l.ChangeSelection(off)
}
