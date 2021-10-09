// In charge of managing input and its associated thread
package input

import (
	"os"
	"unicode/utf8"

	"github.com/ethanv2/podbit/ui"
)

func getInput(out chan rune, errc chan error) {
	var buf [4]byte
	var err error

	_, err = os.Stdin.Read(buf[:])
	if err != nil {
		errc <- err
		out <- 0x0
	}

	c, _ := utf8.DecodeRune(buf[:])

	errc <- nil
	out <- c
}

// Main input loop
//
// Recieves all key inputs serially, one character at a time
// If there is no global keybinding for this key, we pass it
// to the UI subsystem, which can deal with it from there.
//
// Any and all key inputs causes an immediate and full UI redraw
func InputLoop(exit chan int) {
	var c rune
	var char chan rune = make(chan rune)
	var err chan error = make(chan error, 1)

	for {
		go getInput(char, err)

		select {
		case c = <-char:
			if <-err != nil {
				return
			}

			switch c {

			case '1':
				if ui.MenuActive(ui.PlayerMenu) {
					ui.ActivateMenu(ui.RawPlayerMenu)
				} else {
					ui.ActivateMenu(ui.PlayerMenu)
				}
			case '4':
				ui.ActivateMenu(ui.ListMenu)
			case 'q':
				return
			default:
				ui.PassKeystroke(c)
			}

			ui.Redraw(ui.RD_ALL)
		case <-exit:
			return
		}
	}
}
