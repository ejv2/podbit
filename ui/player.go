package ui

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
	root.MovePrint(y, x, "The player will be here")
}

func (l *Player) Input(c rune) {

}
