// Package sound is responsible for playing audio and managing mpv instances
//
// This package is usually interacted with by the front-end UI code to add
// items to the queue. The rest happens automatically using multi-threaded
// player logic. All that needs to be maintained from the interface side is
// the queue and the rest is run automatically.
//
// The queue is a simple FIFO structure formed from a slice of QueueEntries.
// Entries that require downloading are handled gracefully and with as little
// user impact as possible. Usually, the user won't even notice anything happened.
//
// Sound is played through an idle MPV instance which sits in the background and
// recieves media to play when appropriate
package sound

import (
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/ethanv2/podbit/data"

	"github.com/blang/mpv"
)

// Useful player vars
var (
	// PlayerName is the name of the player program to spawn
	PlayerName = "mpv"
	// The path to the RPC endpoint
	PlayerRPC = "/tmp/podbit-mpv"
	// PlayerArgs are the standard arguments to use for the player
	// These are not the final configs of the player, but just used
	// to idle mpv ready to recieve instructions
	PlayerArgs = []string{"--no-video", "--input-ipc-server=" + PlayerRPC}
	// UpdateTime is the time between queue checks and supervision updates
	UpdateTime = 200 * time.Millisecond
)

// Internal: Types of actions
const (
	actPause = iota
	actUnpause
	actToggle
	actStop

	reqPaused
	reqPlaying
	reqTimings
)

// WaitFunc is the function to call waiting between each update
type WaitFunc func(u chan int)

// Player represents the current player instance
type Player struct {
	proc *exec.Cmd

	act chan int
	dat chan interface{}

	exit      chan int
	end       chan int
	watchStop chan int

	ipcc *mpv.IPCClient
	ctrl *mpv.Client

	output io.ReadCloser
	times  io.ReadCloser

	waiting  bool
	download *data.Download

	playing bool

	NowPlaying string
	NowPodcast string
}

var (
	Plr Player
)

func updateWait(u chan int) {
	time.Sleep(UpdateTime)
	u <- 1
}

func endWait(u chan int) {
	Plr.Wait()
	time.Sleep(time.Second)

	Plr.playing = false

	u <- 1
}

func downloadWait(u chan int) {
	for !Plr.download.Completed {
	}

	Plr.waiting = false
	head--

	u <- 1
}

func waitNew(u chan int) {
	for head >= len(queue) {
	}

	Plr.waiting = false
	u <- 1
}

// NewPlayer constructs a new player. This does not yet
// launch any processes or play any media
func NewPlayer(exit chan int) (p Player, err error) {
	p.exit = exit

	p.act = make(chan int)
	p.dat = make(chan interface{})

	p.watchStop = make(chan int)
	p.end = make(chan int)

	return
}

// ConnectPlayer attempts to connect to the RPC endpoint
// Sadly, this is needed because of an exceptionally bad
// design choice in the mpv library forcing me to create
// this bad workaround. :(
func (p *Player) connect() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error: Player connection")
		}
	}()

	p.ipcc = mpv.NewIPCClient(PlayerRPC)
	p.ctrl = mpv.NewClient(p.ipcc)

	return
}

func (p *Player) start(filename string) (err error) {
	p.proc = exec.Command(PlayerName, append(PlayerArgs, filename)...)
	p.output, err = p.proc.StdoutPipe()
	p.times, err = p.proc.StderrPipe()
	p.proc.Start()

	for err = p.connect(); err != nil; {
		err = p.connect()
	}

	return
}

func (p *Player) play(q *data.QueueItem) {
	p.start(q.Path)

	if q.State != data.StatePending {
		now, ok := data.Caching.Query(q.Path)
		if !ok {
			p.NowPlaying = ""
			p.NowPodcast = ""
		}
		p.NowPlaying = now.Title
		p.NowPodcast = data.DB.GetFriendlyName(q.URL)

		p.playing = true
	}
}

func (p *Player) Stop() {
	p.act <- actStop
}

func (p *Player) stop() {
	if !p.playing {
		return
	}

	p.proc.Process.Kill()
	p.playing = false
}

func (p *Player) Destroy() {
	p.end <- 1
}

func (p *Player) IsPaused() bool {
	p.act <- reqPaused

	r := <-p.dat
	return r.(bool)
}

func (p *Player) isPaused() bool {
	paused, _ := p.ctrl.Pause()
	return paused
}

func (p *Player) IsPlaying() bool {
	p.act <- reqPlaying

	r := <-p.dat
	return r.(bool)
}

func (p *Player) isPlaying() bool {
	return p.playing
}

func (p *Player) Pause() {
	p.act <- actPause
}

func (p *Player) pause() {
	if !p.playing {
		return
	}

	// Leave playing set to true so we know not to play another episode
	p.ctrl.SetPause(true)
}

func (p *Player) Unpause() {
	p.act <- actUnpause
}

func (p *Player) unpause() {
	if !p.playing {
		return
	}

	// Leave playing set to true so we know not to play another episode
	p.ctrl.SetPause(false)
}

func (p *Player) Toggle() {
	p.act <- actToggle
}

func (p *Player) toggle() {
	paused, _ := p.ctrl.Pause()
	p.ctrl.SetPause(!paused)
}

// GetTimings returns the current time and duration
// of the ongoing player. Returns zero if we are
// not playing currently.
//
// This function is thread safe but may block until
// data is available
func (p *Player) GetTimings() (float64, float64) {
	p.act <- reqTimings

	var dat [2]float64 = (<-p.dat).([2]float64)
	return dat[0], dat[1]
}

func (p *Player) getTimings() (float64, float64) {
	if !p.playing {
		return 0, 0
	}

	pos, _ := p.ctrl.Position()
	dur, _ := p.ctrl.Duration()

	return pos, dur
}

// Wait for the current episode to complete
func (p *Player) Wait() {
	if !p.playing {
		return
	}

	p.proc.Wait()
}

func Mainloop() {
	var wait WaitFunc = updateWait
	var elem *data.QueueItem

	for {
		elem, Plr.waiting = PopHead()

		if !Plr.playing && !Plr.waiting && len(queue) > 0 {
			if elem.State != data.StatePending && data.Caching.EntryExists(elem.Path) {
				Plr.play(elem)
				wait = endWait
			} else {
				Plr.waiting = true

				if y, dow := data.Caching.IsDownloading(elem.Path); y {
					Plr.download = &data.Caching.Downloads[dow]
				} else {

					id, err := data.Caching.Download(elem)
					if err != nil {
						continue
					}

					Plr.download = &data.Caching.Downloads[id]
				}

				wait = downloadWait
			}
		}

		u := make(chan int)
		go wait(u)
		keepWaiting := true
		for keepWaiting {
			select {
			case <-Plr.end:
				Plr.proc.Process.Kill()
				Plr.playing = false
				return
			case <-u:
				keepWaiting = false

			case action := <-Plr.act:
				switch action {
				case actStop:
					Plr.stop()
				case actPause:
					Plr.pause()
				case actUnpause:
					Plr.unpause()
				case actToggle:
					Plr.toggle()

				case reqPaused:
					Plr.dat <- Plr.isPaused()
				case reqPlaying:
					Plr.dat <- Plr.isPlaying()
				case reqTimings:
					d, p := Plr.getTimings()
					arr := [2]float64{d, p}

					Plr.dat <- arr
				}
			}
		}
	}
}
