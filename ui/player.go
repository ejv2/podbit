package ui

import (
	"github.com/ethanv2/podbit/sound"
	"github.com/ethanv2/podbit/data"
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
	return "Player"
}

func (l *Player) Render(x, y int) {
	pos, dur := sound.Plr.GetTimings()
	p, d := data.FormatTime(pos), data.FormatTime(dur)

	root.MovePrint(h-(h/3), 0, p)
	root.MovePrint(h-(h/3), w-len(d), d)
}

func (l *Player) Input(c rune) {

}
