package ui

import (
	"github.com/ethanv2/podbit/colors"
	"github.com/ethanv2/podbit/data"
	"github.com/ethanv2/podbit/sound"

	"github.com/rthornton128/goncurses"
)

// Player is the full screen player component
//
// Player displays the currently playing episode, the next up
// episode, progress through the episode etc.
//
// This is mostly for user convenience and visual appeal
type Player struct {
	test string
}

func (l *Player) Name() string {
	return "Now playing"
}

func (l *Player) Render(x, y int) {
	pos, dur := sound.Plr.GetTimings()
	percent := pos / dur
	p, d := data.FormatTime(pos), data.FormatTime(dur)

	root.MovePrint(h-(h/3), 0, p)
	root.MovePrint(h-(h/3), w-len(d), d)

	root.HLine(h-(h/3), len(p)+1, goncurses.ACS_HLINE, w-len(d)-len(p)-2)
	root.MovePrint(h-(h/3), len(p), "|")
	root.MovePrint(h-(h/3), len(p)+1+(w-len(d)-len(p)-2), "|")

	wid := int(float64(w-len(d)-len(p)-2) * percent)

	root.ColorOn(colors.ColorBlue)
	root.HLine(h-(h/3), len(p)+1, goncurses.ACS_HLINE, wid)
}

func (l *Player) Input(c rune) {
	switch c {
	case ' ':
		sound.Plr.Toggle()
	}
}
