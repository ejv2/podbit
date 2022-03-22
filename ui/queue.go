package ui

import (
	"github.com/ethanv2/podbit/colors"
	"github.com/ethanv2/podbit/data"
	"github.com/ethanv2/podbit/sound"
	"github.com/ethanv2/podbit/ui/components"
)

var queueHeadings []components.Column = []components.Column{
	{
		Label: "",
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
	tbl components.Table
}

func (q *Queue) Name() string {
	return "Queue"
}

func (q *Queue) Render(x, y int) {
	q.tbl.X, q.tbl.Y = x, y
	q.tbl.W, q.tbl.H = w, h-5
	q.tbl.Win = root

	q.tbl.Columns = queueHeadings

	q.tbl.Items = nil
	for _, elem := range sound.GetQueue() {
		item := make([]string, len(queueHeadings))
		dat, ok := data.Downloads.Query(elem.Path)
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

		q.tbl.Items = append(q.tbl.Items, item)
	}

	q.tbl.Render()
}

func (q *Queue) Input(c rune) {
	switch c {
	case 'j':
		q.tbl.MoveSelection(1)
	case 'k':
		q.tbl.MoveSelection(-1)
	case 'g':
		q.tbl.ChangeSelection(0)
	case 'G':
		q.tbl.ChangeSelection(len(q.tbl.Items) - 1)
	case 'd':
		i, _ := q.tbl.GetSelection()
		sound.Dequeue(i)
	case 13: // Enter key - Jump to this position
		i, _ := q.tbl.GetSelection()
		sound.JumpTo(i)
	}
}
