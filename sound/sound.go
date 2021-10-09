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
	"errors"
	"os/exec"
	"time"

	"github.com/ethanv2/podbit/data"
)

// Player errors
var (
	ErrorSpawnFailed = errors.New("Error: Failed to create player process")
)

// Useful player vars
var (
	// PlayerName is the name of the player program to spawn
	PlayerName = "mpv"
	// PlayerArgs are the standard arguments to use for the player
	// The media file to play will be appended to this on each play
	PlayerArgs = []string{"--no-video"}
	// UpdateTime is the time between queue checks and supervision updates
	UpdateTime = time.Second
)

// Player represents the current player instance
type Player struct {
	proc *exec.Cmd

	Playing  bool
	Finished bool
}

var Plr Player

func (p *Player) Play(q *data.QueueItem) {
	args := append(PlayerArgs, q.Path)
	p.proc = exec.Command(PlayerName, args...)
	p.Playing = true

	p.proc.Start()
}

func (p *Player) Stop() {
	p.proc.Process.Kill()
	p.Playing = false
}

func Mainloop() {
	for {
		if !Plr.Playing {

			for _, elem := range queue {
				if elem.State != data.StatePending {
					Plr.Play(queue[0])
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
