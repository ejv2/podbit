package ui

import (
	"github.com/ethanv2/podbit/data"
	"github.com/ethanv2/podbit/sound"
)

// Queue displays the current play queue
// Not to be confused with the current download queue, Download
type Queue struct {
	test string
}

func (l *Queue) Name() string {
	return "Queue"
}

func (l *Queue) Render(x, y int) {
	for i, elem := range sound.GetQueue() {
		root.MovePrintf(y+i, 0, "Queue index %d: %s", i, data.DB.GetFriendlyName(elem.URL))
	}
}

func (l *Queue) Input(c rune) {
}
