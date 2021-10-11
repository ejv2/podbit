// Package ui implements podbit's main UI and front end user code
//
// This package runs mostly in a separate UI thread and is as thread-safe
// as possible
//
// Due to limitations in the C library ncurses, the render loop is
// designed to only let one thread use ncurses callbacks at a time,
// with as little loss in performance as possible. Threads will wait
// for the time to redraw using channels and modes. Usually, three
// separate threads will run at a time: the menu thread, tray thread
// and main thread. These all interact using the aforementioned channels
// to draw the screen in sync.
//
// The "redraw" chanel is the main channel around which the UI code
// revolves. It is an integer channel which recieves a "mode". This
// mode allows you to select which part of the UI to redraw. This *can*
// be all of them. The UI threads wait around for the redraw channel to
// instruct them as to when they should draw the screen.
//
// The "exit" channel simply instructs us to exit immediately. This should
// *NEVER* be used inside a render callback, least a deadlock in the UI
// code be caused
package ui

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rthornton128/goncurses"
	"golang.org/x/crypto/ssh/terminal"
)

// A Menu is a renderable UI element which takes up most of primary
// screen space and is capable of handling unhandled keybinds
type Menu interface {
	Name() string
	Render(x, y int)
	Input(c rune)
}

// Redraw types
const (
	RedrawAll  = iota // Redraw everything
	RedrawMenu        // Redraw just the menu
	RedrawTray        // Redraw just the tray
)

var (
	root *goncurses.Window
	w, h int

	currentMenu Menu

	redraw    chan int
	keystroke chan rune
)

// Menu singletons
var (
	PlayerMenu    *Player    = new(Player)    // Full screen player
	RawPlayerMenu *RawPlayer = new(RawPlayer) // Raw player output
	QueueMenu     *Queue     = new(Queue)     // Player queue display
	LibraryMenu   *Library   = new(Library)   // Library of podcasts and episodes
)

// Watch the terminal for resizes and redraw when needed
func watchResize(sig chan os.Signal, scr *goncurses.Window) {
	for {
		<-sig
		UpdateDimensions(scr, true)
	}
}

// InitUI initialises the UI subsystem
func InitUI(scr *goncurses.Window, initialMenu Menu, r chan int, k chan rune) {
	redraw = r
	keystroke = k
	root = scr
	currentMenu = initialMenu

	resizeChan := make(chan os.Signal, 1)
	signal.Notify(resizeChan, syscall.SIGWINCH)
	go watchResize(resizeChan, scr)
	go trayWatcher()

	UpdateDimensions(scr, false)
}

// UpdateDimensions changes the dimensions of the drawable area
//
// Called automatically on detected terminal resizes by the resizeLoop
// thread
func UpdateDimensions(scr *goncurses.Window, shouldRedraw bool) {
	var err error
	w, h, err = terminal.GetSize(int(os.Stdin.Fd()))

	if err != nil {
		w, h = 72, 90
	}

	if w < 10 || h < 5 {
		Exit()
	}

	goncurses.ResizeTerm(h, w)

	if shouldRedraw {
		redraw <- RedrawAll
	}
}

func renderMenu() {
	if currentMenu == nil {
		return
	}

	// Clear region
	for i := 0; i < h-2; i++ {
		root.Move(i, 0)
		root.ClearToEOL()
	}
	root.Move(0, 0)

	// Title Text
	root.AttrOn(goncurses.A_BOLD)
	root.Printf("%s", currentMenu.Name())
	root.HLine(1, 0, goncurses.ACS_HLINE, w)
	root.AttrOff(goncurses.A_BOLD)

	// Actually render menu
	currentMenu.Render(0, 2)
}

func renderTray() {
	for i := h - 2; i <= h; i++ {
		root.Move(i, 0)
		root.ClearToEOL()
	}

	RenderTray(root, w, h)
}

// Redraw signals to redraw a specific part of the UI
//
// This call *will* block if a redraw is in progress
// but will not fail
func Redraw(mode int) {
	redraw <- mode
}

// ActivateMenu sets the current menu to the requested value
// and orders a redraw of the menu area
func ActivateMenu(newMenu Menu) {
	currentMenu = newMenu

	Redraw(RedrawMenu)
}

// MenuActive returns true if the current menu claims to be of the same
// class as the passed menu
//
// "compare" does not necessarily have to be exactly the same type as
// the current menu
func MenuActive(compare Menu) bool {
	return currentMenu.Name() == compare.Name()
}

// PassKeystroke performs a keystroke passthrough for the active menu
func PassKeystroke(c rune) {
	keystroke <- c
}

// RenderLoop is the main render callback for the program
// This is intended to run in its own thread
func RenderLoop() {
	for {
		select {
		case toRedraw := <-redraw:

			switch toRedraw {
			case RedrawAll:
				renderMenu()
				renderTray()
			case RedrawMenu:
				renderMenu()
			case RedrawTray:
				renderTray()
			default:
				goncurses.Flash()
			}

			root.Refresh()
		case c := <-keystroke:
			currentMenu.Input(c)
		}
	}
}
