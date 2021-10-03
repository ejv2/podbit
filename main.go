package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethanv2/podbit/input"
	"github.com/ethanv2/podbit/ui"
	"github.com/ethanv2/podbit/data"

	"github.com/juju/fslock"
	"github.com/rthornton128/goncurses"
)

const (
	homebase = "podbit"
	pidfile  = "podbit.lock"
)

var (
	homedir string
	confdir string

	redraw chan int = make(chan int)

	queue data.Queue
)

func banner() {
	fmt.Printf("Starting Podbit %d.%d.%d...\n", verMaj, verMin, verPatch)
}

func initDirs() {
	share := os.Getenv("XDG_DATA_HOME")
	config, _ := os.UserConfigDir()

	if share == "" {
		share = ".local/share"
	}

	homedir = filepath.Join(share, homebase)
	confdir = filepath.Join(config, homebase)

	err := os.MkdirAll(homedir, os.ModeDir|os.ModePerm)
	err = os.MkdirAll(confdir, os.ModeDir|os.ModePerm)

	if err != nil {
		fmt.Println("Error: Failed to required directory(s)")
		os.Exit(1)
	}
}

func alreadyRunning() (bool, *fslock.Lock) {
	lockpath := filepath.Join(homedir, pidfile)
	lock := fslock.New(lockpath)
	err := lock.TryLock()

	if err != nil {
		return true, nil
	}

	return false, lock
}

func initTTY() {
	goncurses.Raw(true)
	goncurses.Echo(false)
	goncurses.Cursor(0)
}

func main() {
	banner()
	initDirs()

	running, lock := alreadyRunning()
	if running {
		fmt.Println("Error: Podbit is already running")
		os.Exit(1)
	}
	defer lock.Unlock()

	fmt.Println("Reading queue...")
	err := queue.Open()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Parsing queue...")
	err = queue.Parse()
	if err != nil {
		fmt.Println(err)
		return
	}

	scr, err := goncurses.Init()
	if err != nil {
		fmt.Printf("Error: Failed to initialize UI: %s\n", err)
		os.Exit(1)
	}
	initTTY()
	defer goncurses.End()

	ui.InitUI(scr, ui.ListMenu, redraw)
	go ui.RenderLoop()

	// Initial UI draw
	redraw <- ui.RD_ALL

	// Initialisation is done; use this thread as the input loop
	input.InputLoop()
}
