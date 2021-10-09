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
)

// Player errors
var (
	ErrorSpawnFailed = errors.New("Error: Failed to create player process")
)

const (
	// PlayerName is the name of the player program to spawn
	PlayerName = "mpv"
	// PlayerArgs are the standard arguments to use for the player
	// "%s" will be replaced with the media file to play
	PlayerArgs = "--no-video %s"
)

// Plr represents the current player instance
var Plr struct {
	proc exec.Cmd
}
