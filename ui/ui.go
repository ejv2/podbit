package ui

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rthornton128/goncurses"
	"golang.org/x/crypto/ssh/terminal"
)

// Represents a window in the current state
type Menu interface {
	Name() string
	Render(x, y int)
}

// Redraw types
const (
	RD_ALL  = iota // Redraw everything
	RD_MENU        // Redraw just the menu
	RD_TRAY        // Redraw just the tray
)

var (
	root *goncurses.Window
	w, h int

	currentMenu Menu

	redraw chan int
)

// Menu singletons
var (
	ListMenu List
)

// Watch the terminal for resizes and redraw when needed
func watchResize(sig chan os.Signal, scr *goncurses.Window) {
	for {
		<-sig
		UpdateDimensions(scr, true)
	}
}

// Initialise the UI subsystem
func InitUI(scr *goncurses.Window, initialMenu Menu, r chan int) {
	redraw = r
	root = scr
	currentMenu = initialMenu

	resizeChan := make(chan os.Signal, 1)
	signal.Notify(resizeChan, syscall.SIGWINCH)
	go watchResize(resizeChan, scr)

	UpdateDimensions(scr, false)
}

// Change the dimensions of the terminal
func UpdateDimensions(scr *goncurses.Window, shouldRedraw bool) {
	var err error
	w, h, err = terminal.GetSize(int(os.Stdin.Fd()))

	if err != nil {
		w, h = 72, 90
	}

	goncurses.ResizeTerm(h, w)

	if shouldRedraw {
		redraw <- RD_ALL
	}
}

func renderMenu() {
	if currentMenu == nil {
		return
	}

	// Title Text
	root.Printf("%s", currentMenu.Name())
	root.HLine(1, 0, goncurses.ACS_HLINE, w)

	// Actually render menu
	currentMenu.Render(0, 3)
}

func renderTray() {
}

// Main render loop. Calls specific renderers when required
func RenderLoop() {
	for {
		toRedraw := <-redraw

		root.Clear()
		switch toRedraw {
		case RD_ALL:
			renderMenu()
			renderTray()
		case RD_MENU:
			renderMenu()
		case RD_TRAY:
			renderTray()
		default:
			goncurses.Flash()
		}

		root.Refresh()
	}
}
