// In charge of managing input and its associated thread
package ui

import (
	"os"
	"unicode/utf8"
)

var (
	exitChan chan int
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

func Exit() {
	exitChan <- 1
}

// Main input loop
//
// Receives all key inputs serially, one character at a time
// If there is no global keybinding for this key, we pass it
// to the UI subsystem, which can deal with it from there.
//
// Any and all key inputs causes an immediate and full UI redraw
func InputLoop(exit chan int) {
	exitChan = exit

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
				if MenuActive(PlayerMenu) {
					ActivateMenu(RawPlayerMenu)
				} else {
					ActivateMenu(PlayerMenu)
				}
			case '2':
				ActivateMenu(QueueMenu)
			case '4':
				ActivateMenu(ListMenu)
			case 'q':
				return
			default:
				PassKeystroke(c)
			}

			Redraw(RD_ALL)
		case <-exit:
			return
		}
	}
}
