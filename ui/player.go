package ui

import (
	"math"

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

	maxwt := int(math.Min(float64(w), float64(len(sound.Plr.NowPlaying))))
	maxwp := int(math.Min(float64(w), float64(len(sound.Plr.NowPodcast))))

	minxt := int(math.Max(0, float64((w-len(sound.Plr.NowPlaying))/2)))
	minxp := int(math.Max(0, float64((w-len(sound.Plr.NowPodcast))/2)))

	var stat, play, div, pod string
	if sound.Plr.IsPlaying() {
		stat = "Now Playing: "
		play = sound.Plr.NowPlaying[:maxwt]
		div = "by"
		pod = sound.Plr.NowPodcast[:maxwp]
	} else {
		stat = "Not playing"
		play = ""
		div = ""
		pod = ""
	}

	// Now playing
	root.ColorOn(colors.ColorRed)
	root.MovePrint(x+4, (w-len(stat))/2, stat)
	root.ColorOff(colors.ColorRed)

	// [episode name]
	root.AttrOn(goncurses.A_BOLD)
	root.MovePrint(x+6, minxt, play)
	root.AttrOff(goncurses.A_BOLD)

	// by
	root.ColorOn(colors.ColorGreen)
	root.MovePrint(x+8, (w-len(div))/2, div)
	root.ColorOff(colors.ColorGreen)

	// [podcast name]
	root.ColorOn(colors.ColorYellow)
	root.MovePrint(x+10, minxp, pod)
	root.ColorOff(colors.ColorGreen)

	root.MovePrint(h-(h/3), 0, p)
	root.MovePrint(h-(h/3), w-len(d), d)

	root.HLine(h-(h/3), len(p)+1, goncurses.ACS_HLINE, w-len(d)-len(p)-2)
	root.MovePrint(h-(h/3), len(p), "|")
	root.MovePrint(h-(h/3), len(p)+1+(w-len(d)-len(p)-2), "|")

	wid := int(float64(w-len(d)-len(p)-2) * percent)

	root.ColorOn(colors.ColorBlue)
	root.HLine(h-(h/3), len(p)+1, goncurses.ACS_HLINE, wid)
	root.ColorOff(colors.ColorBlue)
}

func (l *Player) Input(c rune) {
	switch c {
	case ' ':
		sound.Plr.Toggle()
	}
}
