package ui

import (
	"sync"
	"time"

	"github.com/rthornton128/goncurses"
)

const (
	MSG_TIME time.Duration = 2 * time.Second
)

var (
	msgMutex      sync.Mutex
	statusMessage string
)

// Renders the statusbar tray at the bottom of the screen
// Tray takes up two vertical cells and the entirety of the width
// The top cell is a horizontal line denoting a player status bar
// The bottom cell is the status text
func RenderTray(scr *goncurses.Window, w, h int) {
	scr.HLine(h-2, 0, goncurses.ACS_HLINE, w)

	if statusMessage != "" {
		scr.MovePrint(h-1, 0, statusMessage)
	} else {
		scr.MovePrint(h-1, 0, "Status text: Status1234")
	}
}

// Send a status message
// Blocks until the message has completed displaying
// Will wait for the previous user to unlock the message bar first
func StatusMessage(msg string) {
	msgMutex.Lock()

	statusMessage = msg
	time.Sleep(MSG_TIME)
	statusMessage = ""

	msgMutex.Unlock()
}
