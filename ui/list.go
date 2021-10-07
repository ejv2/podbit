// Lists your configured/detected podcasts and available episodes
package ui

import (
	"github.com/ethanv2/podbit/data"
	"github.com/ethanv2/podbit/ui/components"
)

type List struct {
	men components.Menu
}

func (l *List) Name() string {
	return "Podcasts"
}

func (l *List) Render(x, y int) {
	l.men.X = x
	l.men.Y = y

	l.men.W, l.men.H = w, (h - 5)
	l.men.Win = *root

	l.men.Items = make([]string, len(data.Q.Items))
	for i := range l.men.Items {
		l.men.Items[i] = data.DB.GetFriendlyName(data.Q.Items[i].Url)
	}

	l.men.Render()
}

func (l *List) Input(c rune) {
	switch c {
	case 'j':
		l.men.MoveSelection(1)
	case 'k':
		l.men.MoveSelection(-1)
	}
}
