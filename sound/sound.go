// Sound subsystem: responsible for playing audio and managing mpv instances
package sound

import (
	"errors"
	"os/exec"
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
var Plr struct {
	proc exec.Cmd
}
