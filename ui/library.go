package ui

import (
	"fmt"

	"github.com/ethanv2/podbit/data"
	"github.com/ethanv2/podbit/sound"
	"github.com/ethanv2/podbit/ui/components"

	"github.com/rthornton128/goncurses"
)

// Library represents the list menu type and state
//
// Library displays all detected and configured podcasts, along
// with associated episodes sorted into said podcasts
type Library struct {
	men [2]components.Menu

	menSel int
}

func (l *Library) Name() string {
	return "Library"
}

func (l *Library) renderPodcasts(x, y int) {
	l.men[0].X = x
	l.men[0].Y = y

	l.men[0].W, l.men[0].H = (w/2)-1, (h - 5)
	l.men[0].Win = *root

	l.men[0].Items = l.men[0].Items[:0]

	seen := make(map[string]bool)
	for i := range data.Q.Items {
		name := data.DB.GetFriendlyName(data.Q.Items[i].URL)

		if !seen[name] {
			l.men[0].Items = append(l.men[0].Items, name)
			seen[name] = true
		}
	}

	l.men[0].Selected = true

	l.men[0].Render()
}

func (l *Library) renderEpisodes(x, y int) {
	if len(l.men[0].Items) < 1 {
		return
	}

	l.men[1].X = x
	l.men[1].Y = y

	l.men[1].W, l.men[1].H = (w/2)-1, (h - 5)
	l.men[1].Win = *root

	l.men[1].Items = l.men[1].Items[:0]

	for i := len(data.Q.Items) - 1; i >= 0; i-- {
		elem := data.Q.Items[i]

		if data.DB.GetFriendlyName(elem.URL) == l.men[0].GetSelection() {
			var text string
			entry, ok := data.Caching.Query(elem.Path)
			title := entry.Title
			if !ok || title == "" {
				text = elem.URL
			} else {
				text = title
			}

			l.men[1].Items = append(l.men[1].Items, text)
		}
	}

	l.men[1].Selected = (l.menSel == 1)

	l.men[1].Render()
}

func (l *Library) Render(x, y int) {
	l.renderPodcasts(x, y)

	root.AttrOn(goncurses.A_BOLD)
	root.VLine(y, w/2, goncurses.ACS_VLINE, h-2-y)
	root.AttrOff(goncurses.A_BOLD)

	l.renderEpisodes(w/2+1, y)
}

func (l *Library) Input(c rune) {
	switch c {
	case 'j':
		l.men[l.menSel].MoveSelection(1)
	case 'k':
		l.men[l.menSel].MoveSelection(-1)
	case 'h':
		l.MoveSelection(-1)
	case 'l':
		l.MoveSelection(1)
	case ' ':
		l.StartDownload()
	case 13:
		l.StartPlaying() // Enter key
	}
}

func (l *Library) ChangeSelection(index int) {
	if index >= len(l.men) || index < 0 {
		return
	}

	l.menSel = index
}

func (l *Library) MoveSelection(direction int) {
	if direction == 0 {
		return
	}

	off := l.menSel + direction
	l.ChangeSelection(off)
}

// StartDownload downloads the currently focused library entry
func (l *Library) StartDownload() {
	if len(l.men[0].Items) < 1 || len(l.men[1].Items) < 1 {
		return
	}

	targets := l.men[1].Items
	if l.menSel == 1 {
		for i, elem := range data.Q.Items {
			if elem.URL == l.men[1].GetSelection() {
				go data.Caching.Download(&data.Q.Items[i])
				go StatusMessage(fmt.Sprintf("Download of %s started...", elem.URL))

				return
			}
		}
	} else {
		for _, elem := range targets {
			if data.IsURL(elem) {
				for i, q := range data.Q.Items {
					if q.URL == elem {
						go data.Caching.Download(&data.Q.Items[i])
					}
				}
			}
		}

		go StatusMessage("Download of multiple episodes started...")

	}

}

// StartPlaying begins playing the currently focused element
// If the current focus requires downloading (and enough information
// is known to oblige) it will first be downloaded
func (l *Library) StartPlaying() {
	if len(l.men[0].Items) < 1 || len(l.men[1].Items) < 1 {
		return
	}

	if l.menSel == 1 {
		entry := l.men[1].GetSelection()
		if data.IsURL(entry) {
			sound.EnqueueByURL(entry)
		} else {
			sound.EnqueueByTitle(entry)
		}

		go StatusMessage(fmt.Sprintf("Enqueued episode %q to play", entry))
	} else {
		entry := l.men[0].GetSelection()

		sound.EnqueueByPodcast(entry)

		go StatusMessage("Multiple episodes enqueued...")
	}
}
