package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/juju/fslock"
	"github.com/rthornton128/goncurses"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	homebase = "podbit"
	pidfile  = "podbit.lock"
)

var (
	homedir string
	confdir string

	w, h int
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

func UpdateDimensions(scr *goncurses.Window) {
	var err error
	w, h, err = terminal.GetSize(int(os.Stdin.Fd()))

	if err != nil {
		w, h = 72, 90
	}

	scr.Resize(h, w)
}

func watchResize(sig chan os.Signal, scr *goncurses.Window) {
	for {
		<-sig
		UpdateDimensions(scr)
	}
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

	scr, err := goncurses.Init()
	if err != nil {
		fmt.Printf("Error: Failed to initialize UI: %s\n", err)
		os.Exit(1)
	}
	initTTY()
	defer goncurses.End()

	resizeChan := make(chan os.Signal, 1)
	signal.Notify(resizeChan, syscall.SIGWINCH)

	UpdateDimensions(scr)
	go watchResize(resizeChan, scr)
}
