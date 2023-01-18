package ui

import (
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/ejv2/podbit/data"
	ev "github.com/ejv2/podbit/event"
	"github.com/ejv2/podbit/sound"
)

var (
	exitChan chan struct{}
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

// Exit requests that the input handler shuts down and gracefully
// exits the program via a return to the main function.
func Exit() {
	close(exitChan)
}

// InputLoop - main UI input handler
//
// Receives all key inputs serially, one character at a time
// If there is no global keybinding for this key, we pass it
// to the UI subsystem, which can deal with it from there.
//
// Any and all key inputs causes an immediate and full UI redraw.
func InputLoop(exit chan struct{}) {
	exitChan = exit

	var c rune
	var char = make(chan rune)
	var err = make(chan error, 1)

	for {
		go getInput(char, err)

		select {
		case c = <-char:
			if <-err != nil {
				return
			}

			switch c {

			case '1':
				ActivateMenu(PlayerMenu)
			case '2':
				ActivateMenu(QueueMenu)
			case '3':
				ActivateMenu(DownloadMenu)
			case '4':
				ActivateMenu(LibraryMenu)
			case 'r':
				data.Q.Reload()
				go StatusMessage("Queue file reloaded")
			case 'p':
				sound.Plr.Toggle()
			case 's':
				sound.Plr.Stop()
			case 'c':
				sound.ClearQueue()
			case 'a':
				pending := data.Q.GetByStatus(data.StatePending)
				for _, elem := range pending {
					go data.Downloads.Download(elem)
				}

				msg := fmt.Sprintf("Downloading %d episodes in parallel", len(pending))
				go StatusMessage(msg)
			case ']':
				sound.Plr.Seek(5)
			case '[':
				sound.Plr.Seek(-5)
			case '}':
				sound.Plr.Seek(60)
			case '{':
				sound.Plr.Seek(-60)
			case '\f': // Control-L
				root.Clear()
				UpdateDimensions(root)
			case 'q':
				if data.Downloads.Ongoing() == 0 {
					return
				}

				StatusMessage("Error: Cannot quit with ongoing downloads")
			default:
				PassKeystroke(c)
			}

			events <- ev.Keystroke
		case <-exit:
			return
		}
	}
}
