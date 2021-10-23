package ui

import (
	"github.com/ethanv2/podbit/colors"
	"github.com/ethanv2/podbit/data"
	"github.com/ethanv2/podbit/sound"
	"github.com/ethanv2/podbit/ui/components"
)

var queueHeadings []components.Column = []components.Column{
	{
		Label: "Status",
		Width: 0.1,
		Color: colors.BackgroundRed,
	},
	{
		Label: "Name",
		Width: 0.5,
		Color: colors.BackgroundBlue,
	},
	{
		Label: "Podcast",
		Width: 0.4,
		Color: colors.BackgroundGreen,
	},
}

// Queue displays the current play queue
// Not to be confused with the current download queue, Download
type Queue struct {
}

func (l *Queue) Name() string {
	return "Queue"
}

func (l *Queue) Render(x, y int) {
	var tbl components.Table

	tbl.X, tbl.Y = x, y
	tbl.W, tbl.H = w, h
	tbl.Win = root

	tbl.Columns = queueHeadings

	for _, elem := range sound.GetQueue() {
		item := make([]string, len(queueHeadings))
		dat, ok := data.Caching.Query(elem.Path)
		pod := data.DB.GetFriendlyName(elem.URL)


		if !ok || dat.Title == "" {
			item[1] = elem.URL
		} else {
			item[1] = dat.Title
		}

		if !ok {
			// In need of download
			item[0] += "!!"
		} else if sound.Plr.NowPlaying == item[1] {
			// Currently playing
			item[0] += ">>"
		}

		item[2] = pod

		tbl.Items = append(tbl.Items, item)
	}

	tbl.Render()
}

func (l *Queue) Input(c rune) {
}
