// Sound subsystem: responsible for playing audio and managing mpv instances
package sound

import (
	"errors"
	"os/exec"

	"github.com/ethanv2/podbit/data"
)

// Player errors
var (
	PlayerSpawnFailed = errors.New("Error: Failed to create player process")
)

const (
	PLAYER_NAME = "mpv"
	PLAYER_ARGS = "--no-video %s"
)

// Player instance
type Player struct {
	proc exec.Cmd
}

// Singleton state for the sound subsystem
var (
	Plr   Player

	queue []data.QueueItem
)

func Enqueue(item data.QueueItem) {
	queue = append(queue, item)
}

func ClearQueue() {
	queue = queue[:0]
}

func GetQueue() []data.QueueItem {
	return queue
}
