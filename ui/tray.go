package ui

import (
	"fmt"
	"sync"
	"time"
	"math"

	"github.com/ethanv2/podbit/data"
	"github.com/ethanv2/podbit/sound"

	"github.com/rthornton128/goncurses"
)

const (
	// MessageTime is the time a message will show for
	MessageTime time.Duration = 2 * time.Second
)

var (
	msgMutex      sync.Mutex
	statusMessage string
)

func trayWatcher() {
	for {
		time.Sleep(500 * time.Millisecond)

		if MenuActive(PlayerMenu) {
			Redraw(RedrawAll)
		} else {
			Redraw(RedrawTray)
		}
	}
}

// RenderTray renders the statusbar tray at the bottom of the screen
// Tray takes up two vertical cells and the entirety of the width
// The top cell is a horizontal line denoting a player status bar
// The bottom cell is the status text
func RenderTray(scr *goncurses.Window, w, h int) {
	scr.HLine(h-2, 0, goncurses.ACS_HLINE, w)

	pos, dur := sound.Plr.GetTimings()
	percent := pos / dur
	scr.HLine(h-2, 0, '=', int(percent * float64(w)))

	head := int(math.Max((percent * float64(w)) - 1, 0))
	scr.MovePrint(h-2, head, ">")

	if statusMessage != "" {
		scr.MovePrint(h-1, 0, statusMessage)
	} else {
		pos, dur := sound.Plr.GetTimings()
		p, d := data.FormatTime(pos), data.FormatTime(dur)
		code := fmt.Sprintf("[%s/%s]", p, d)

		if sound.Plr.Playing {
			// TODO: Show now playing name here
			var status string
			if sound.Plr.Paused {
				status = "Paused"
			} else {
				status = "Playing"
			}

			scr.MovePrintf(h-1, 0, "%s: %s - %s", status, sound.Plr.NowPlaying, sound.Plr.NowPodcast)
			scr.MovePrintf(h-1, w-len(code), "%s", code)
		} else {
			scr.MovePrint(h-1, 0, "Not playing")
		}
	}
}

// StatusMessage sends a status message
//
// Blocks until the message has completed displaying
// Will wait for the previous user to unlock the message bar first
// Every message can be guaranteed MSG_TIME display time
func StatusMessage(msg string) {
	msgMutex.Lock()

	statusMessage = msg
	time.Sleep(MessageTime)
	statusMessage = ""

	msgMutex.Unlock()
}
