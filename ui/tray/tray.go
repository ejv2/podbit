package tray

import (
	"github.com/rthornton128/goncurses"
)

var (
	statusMessage string
)

// Renders the statusbar tray at the bottom of the screen
// Tray takes up two vertical cells and the entirety of the width
// The top cell is a horizontal line denoting a player status bar
// The bottom cell is the status text
func RenderTray(scr *goncurses.Window, w, h int) {
	scr.HLine(h-2, 0, goncurses.ACS_HLINE, w)

	scr.MovePrint(h-1, 0, "Status text: Status1234")
}
