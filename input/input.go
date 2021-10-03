// In charge of managing input and its associated thread
package input

import (
	"os"

	"github.com/ethanv2/podbit/ui"
)

func InputLoop() {
	var buf [1]byte
	var chr rune
	for {
		_, err := os.Stdin.Read(buf[:])
		if err != nil {
			return
		}

		chr = rune(buf[0])

		switch chr {
		case 'q':
			return
		}

		ui.Redraw(ui.RD_ALL)
	}
}
