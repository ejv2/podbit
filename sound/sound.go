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
// Sound is played through mpv instances which are supervised by a nanny goroutine
// spawned upon playing each queue entry. This player is destroyed upon the next
// piece of media being requested.
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
	PlayerArgs = []string{"--no-video", "--no-config", "--idle", "--input-ipc-server=" + PlayerRPC}
	// UpdateTime is the time between queue checks and supervision updates
	UpdateTime = time.Second
)

// Player represents the current player instance
type Player struct {
	proc *exec.Cmd

	ipcc *mpv.IPCClient
	ctrl *mpv.Client

	output io.ReadCloser
	times  io.ReadCloser

	Playing  bool
	Finished bool
}

var Plr Player

func NewPlayer() (p Player, err error) {
	p.proc = exec.Command(PlayerName, PlayerArgs...)
	p.output, err = p.proc.StdoutPipe()
	p.times, err = p.proc.StderrPipe()
	p.proc.Start()

	return
}

// ConnectPlayer attempts to connect to the RPC endpoint
// Sadly, this is needed because of an exceptionally bad
// design choice in the mpv library forcing me to create
// this bad workaround. :(
func ConnectPlayer(p *Player) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error: Player connection")
		}
	}()

	p.ipcc = mpv.NewIPCClient(PlayerRPC)
	p.ctrl = mpv.NewClient(p.ipcc)

	return
}

func (p *Player) Play(q *data.QueueItem) {
	if q.State != data.StatePending {
		p.ctrl.Loadfile(q.Path, mpv.LoadFileModeReplace)
		p.Playing = true
	}
}

func (p *Player) Stop() {
	p.ctrl.SetPause(true)
	p.Playing = false
}

func (p *Player) Destroy() {
	p.proc.Process.Kill()
	p.Playing = false
}

func (p *Player) Pause() {
	// Leave playing set to true so we know not to play another episode
	p.ctrl.SetPause(true)
}

func (p *Player) Toggle() {
	paused, _ := p.ctrl.Pause()
	p.ctrl.SetPause(!paused)
}

// GetTimings returns the current time and duration
// of the ongoing player. Returns zero if we are
// not playing currently
func (p *Player) GetTimings() (float64, float64) {
	if !p.Playing {
		return 0, 0
	}

	pos, _ := p.ctrl.Position()
	dur, _ := p.ctrl.Duration()

	return pos, dur
}

func Mainloop() {
	for {
		if !Plr.Playing {

			for _, elem := range queue {
				if elem.State != data.StatePending {
					Plr.Play(PopQueue())
				} else {
					data.Caching.Download(elem)
					break
				}
			}
		} else {
		}

		time.Sleep(UpdateTime)
	}
}
