// In charge of managing input and its associated thread
package input

import (
	"os"
	"unicode/utf8"

	"github.com/ethanv2/podbit/ui"
)

// Main input loop
//
// Recieves all key inputs serially, one character at a time
// If there is no global keybinding for this key, we pass it
// to the UI subsystem, which can deal with it from there.
//
// Any and all key inputs causes an immediate and full UI redraw
func InputLoop() {
	var buf [4]byte
	var err error

	for err == nil {
		_, err = os.Stdin.Read(buf[:])
		c, _ := utf8.DecodeRune(buf[:])

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
	}
}
