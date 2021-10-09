// Queue contents display
package ui

import (
	"github.com/ethanv2/podbit/sound"
	"github.com/ethanv2/podbit/data"
)

type Queue struct {
	test string
}

func (l *Queue) Name() string {
	return "Queue"
}

func (l *Queue) Render(x, y int) {
	for i, elem := range sound.GetQueue() {
		root.MovePrintf(y + i , 0, "Queue index %d: %s", i, data.DB.GetFriendlyName(elem.Url))
	}
}

func (l *Queue) Input(c rune) {
}
