package ui

// RawPlayer represents a RawPlayer menu component
//
// The RawPlayer allows the user to see the raw output of the
// active player and send raw keystrokes to it
type RawPlayer struct {
	test string
}

func (p *RawPlayer) Name() string {
	return "Player output"
}

func (p *RawPlayer) Render(x, y int) {
	root.MovePrint(y, x, "The raw player will be here")
}

func (p *RawPlayer) Input(c rune) {

}
