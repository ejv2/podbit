package main

import (
	"github.com/ethanv2/podbit/ui"

	"github.com/rthornton128/goncurses"
)

func InputLoop(win *goncurses.Window) {
	for {
		c := win.GetChar()
		switch c {
		case 'q':
			return
		}

		redraw <- ui.RD_ALL
	}
}
