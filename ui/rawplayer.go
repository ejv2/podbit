// Shows the raw output from the spawned player and allows raw inputs to it
package ui

type RawPlayer struct {
	test string
}

func (p RawPlayer) Name() string {
	return "Player - Raw View"
}

func (p RawPlayer) Render(x, y int) {
	root.MovePrint(y, x, "The raw player will be here")
}
