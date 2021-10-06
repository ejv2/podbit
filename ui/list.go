// Lists your configured/detected podcasts and available episodes
package ui

import (
	"github.com/ethanv2/podbit/data"
	"github.com/ethanv2/podbit/ui/components"
)

type List struct {
}

func (l List) Name() string {
	return "Podcasts"
}

func (l List) Render(x, y int) {
	var men components.Menu
	men.X = x
	men.Y = y

	men.W, men.H = w, h
	men.Win = *root

	men.Items = make([]string, len(data.Q.Items))
	for i := range men.Items {
		men.Items[i] = data.DB.GetFriendlyName(data.Q.Items[i].Url)
	}

	men.Render()
}
