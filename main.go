package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ejv2/podbit/colors"
	"github.com/ejv2/podbit/data"
	ev "github.com/ejv2/podbit/event"
	"github.com/ejv2/podbit/sound"
	"github.com/ejv2/podbit/ui"

	"github.com/juju/fslock"
	goncurses "github.com/vit1251/go-ncursesw"
)

const (
	homebase = "podbit"
	pidfile  = "podbit.lock"
)

var (
	homedir string
	confdir string

	events    *ev.Handler
	keystroke = make(chan rune)
	newMen    = make(chan ui.Menu)
	exit      = make(chan struct{})
)

func banner() {
	fmt.Printf("Starting Podbit v%d.%d.%d...\n", verMaj, verMin, verPatch)
}

func initDirs() {
	share := os.Getenv("XDG_DATA_HOME")
	config, _ := os.UserConfigDir()

	if share == "" {
		share = ".local/share"
	}

	homedir = filepath.Join(share, homebase)
	confdir = filepath.Join(config, homebase)

	herr := os.MkdirAll(homedir, os.ModeDir|os.ModePerm)
	cerr := os.MkdirAll(confdir, os.ModeDir|os.ModePerm)

	if herr != nil || cerr != nil {
		fmt.Println("Error: Failed to create required directory(s)")
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

func initColors() {
	if goncurses.HasColors() {
		goncurses.StartColor()
	}

	colors.CreateColors()
}

func initTTY() {
	goncurses.Raw(true)
	goncurses.Echo(false)
	goncurses.Cursor(0)
}

func initSignals(exit chan struct{}) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigchan
		exit <- struct{}{}
	}()
}

func main() {
	banner()
	initDirs()
	initSignals(exit)

	running, lock := alreadyRunning()
	if running {
		fmt.Println("Error: Podbit is already running")
		os.Exit(1)
	}
	defer lock.Unlock()

	events = ev.NewHandler()
	now := time.Now()
	err := data.InitData(*events)
	if err != nil {
		fmt.Println("\n" + err.Error())
		return
	}
	defer data.SaveData()
	go data.ReloadLoop()

	fmt.Print("Initialising sound system...")
	sound.Plr, err = sound.NewPlayer(events)
	if err != nil {
		fmt.Printf("\nError: Failed to initialise sound system: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Println("done")

	go sound.Mainloop()
	defer sound.Plr.Destroy()

	scr, err := goncurses.Init()
	if err != nil {
		fmt.Printf("Error: Failed to initialize UI: %s\n", err)
		os.Exit(1)
	}
	initTTY()
	initColors()
	defer goncurses.End()

	ui.InitUI(scr, ui.LibraryMenu, events, keystroke, newMen)
	go ui.RenderLoop()

	// Welcome message
	startup := time.Since(now)
	go ui.StatusMessage(fmt.Sprintf("Podbit v%d.%d.%d -- %d episodes of %d podcasts loaded in %.2fs",
		verMaj,
		verMin,
		verPatch,
		len(data.Q.Items), len(data.Q.GetPodcasts()),
		startup.Seconds()))

	// Run events handler and kickstart listeners
	go events.Run()
	events.Post(ev.Keystroke)

	// Initialisation is done; use this thread as the input loop
	ui.InputLoop(exit)
}
