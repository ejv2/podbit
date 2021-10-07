// Full screen player UI
package ui

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
