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
// receives media to play when appropriate
package sound

import (
	"fmt"
	"math"
	"os/exec"
	"time"

	"github.com/ejv2/podbit/data"
	ev "github.com/ejv2/podbit/event"

	"github.com/blang/mpv"
)

// Useful player vars.
var (
	// PlayerName is the name of the player program to spawn.
	PlayerName = "mpv"
	// The path to the RPC endpoint.
	PlayerRPC = "/tmp/podbit-mpv"
	// PlayerArgs are the standard arguments to use for the player.
	// These are not the final configs of the player, but just used
	// to idle mpv ready to receive instructions.
	PlayerArgs = []string{"--idle", "--no-video", "--input-ipc-server=" + PlayerRPC}
	// UpdateTime is the time between queue checks and supervision updates.
	UpdateTime = 500 * time.Millisecond
)

// Internal: Types of actions.
const (
	actPause = iota
	actUnpause
	actToggle
	actStop
	actTerm
	actSeek

	reqPaused
	reqPlaying
	reqWaiting
	reqTimings
)

// WaitFunc is the function to call waiting between each update.
type WaitFunc func(u chan int)

// Player represents the current player instance. This persists
// after the media has completed playing and becomes ineffective
// until the next call to play.
type Player struct {
	proc *exec.Cmd

	hndl   ev.Handler
	event  chan int
	dlchan chan struct{}

	act chan int
	dat chan interface{}

	ipcc *mpv.IPCClient
	ctrl *mpv.Client

	waiting  bool
	download *data.QueueItem

	exhausted  bool
	playing    bool
	manualStop bool

	Now        *data.QueueItem
	NowPlaying string
	NowPodcast string
}

// Plr is the singleton player instance.
var Plr Player

func updateWait(u chan int) {
	time.Sleep(UpdateTime)
	u <- 1
}

func endWait(u chan int) {
	Plr.Wait()
	time.Sleep(time.Second)

	Plr.playing = false
	Plr.NowPlaying = ""
	Plr.NowPodcast = ""

	// Set state to finished
	data.Q.Range(func(_ int, item *data.QueueItem) bool {
		if item.Path == Plr.Now.Path {
			if !Plr.manualStop {
				item.State = data.StateFinished
				// We just played this fime so I reckon we can ignore
				// file not found errors.
				data.Stamps.Touch(item.Path)
			}
			return false
		}

		return true
	})

	Plr.manualStop = false
	Plr.hndl.Post(ev.PlayerChanged)
	u <- 1
}

func downloadWait(u chan int) {
	var id int
	for y := true; y && DownloadAtHead(Plr.download); y, id = data.Downloads.IsDownloading(Plr.download.Path) {
		<-Plr.dlchan
	}

	Plr.waiting = false

	dl := data.Downloads.GetDownload(id)

	// Only attempt to play if
	//	a) still present
	//	b) the download succeeded
	if DownloadAtHead(Plr.download) && dl.Success {
		head--
	}

	Plr.hndl.Post(ev.PlayerChanged)
	u <- 1
}

// NewPlayer constructs a new player. This does not yet
// launch any processes or play any media.
func NewPlayer(events *ev.Handler) (p Player, err error) {
	p.act = make(chan int)
	p.dat = make(chan interface{})
	p.dlchan = make(chan struct{}, 1)

	p.hndl = *events
	p.event = events.Register()

	return
}

// ConnectPlayer attempts to connect to the RPC endpoint
// Sadly, this is needed because of an exceptionally bad
// design choice in the mpv library forcing me to create
// this bad workaround. :( .
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

func (p *Player) start() {
	p.proc = exec.Command(PlayerName, PlayerArgs...)
	p.proc.Start()

	go p.procWatcher()

	for err := p.connect(); err != nil; {
		err = p.connect()
	}
}

func (p *Player) procWatcher() {
	p.proc.Wait()
	// Should mpv exit unexpectedly, we need to close the program
	// gracefully.
	p.hndl.Post(ev.RequestShutdown)
}

func (p *Player) load(filename string, starttime int) {
	if p.proc == nil || p.ctrl == nil {
		p.start()
	}

	p.ctrl.Loadfile(filename, mpv.LoadFileModeAppendPlay)

	// Wait for track to load
	for loaded := ""; loaded != "\""+filename+"\""; loaded, _ = p.ctrl.Path() {
	}
	for {
		pos, _ := p.ctrl.PercentPosition()
		if pos != 0 {
			break
		}
	}

	p.ctrl.Seek(starttime, mpv.SeekModeAbsolute)
}

func (p *Player) play(q *data.QueueItem) {
	_, s, err := data.Stamps.Stat(q.Path)
	if err != nil {
		tmp := uint64(0)
		s = &tmp
	}
	if *s > uint64(math.MaxInt) {
		tmp := uint64(math.MaxInt)
		s = &tmp
	}

	p.load(q.Path, int(*s))

	if q.State != data.StatePending {
		Plr.Now = q

		now, ok := data.Downloads.Query(q.Path)
		if !ok {
			p.NowPlaying = ""
			p.NowPodcast = ""
		}

		if now.Title == "" {
			p.NowPlaying = q.URL
		} else {
			p.NowPlaying = now.Title
		}

		p.NowPodcast = data.DB.GetFriendlyName(q.URL)

		p.playing = true
		p.unpause()
	}
}

// Stop ends playback of the current audio track, but does not
// destroy the sound mainloop. This will usually result in the
// next podcast playing.
func (p *Player) Stop() {
	p.act <- actStop
}

func (p *Player) stop() {
	if !p.playing {
		return
	}

	Plr.NowPlaying = ""
	Plr.NowPodcast = ""
	p.manualStop = true

	pos, err := p.ctrl.Position()
	if err == nil {
		data.Stamps.Resume(p.Now.Path, uint64(pos))
	}

	p.ctrl.Exec("stop")
	p.playing = false
}

// Destroy forces the current player instance to terminate and destroys
// the sound mainloop.
// Blocks until the process is guaranteed destroyed.
func (p *Player) Destroy() {
	p.act <- actTerm
	<-p.dat
}

// IsPaused requests the sound mainloop to return if it is paused. If
// the mainloop is not playing, this function returns false. If the mainloop
// is busy or processing audio, this function blocks until the mainloop is
// ready.
func (p *Player) IsPaused() bool {
	p.act <- reqPaused

	r := <-p.dat
	return r.(bool)
}

func (p *Player) isPaused() bool {
	if !p.playing {
		return false
	}

	paused, _ := p.ctrl.Pause()
	return paused
}

// IsPlaying requests that the sound mainloop return if it is currently
// playing audio. If the mainloop is busy or processing audio, this function
// blocks until the mainloop is ready.
func (p *Player) IsPlaying() bool {
	p.act <- reqPlaying

	r := <-p.dat
	return r.(bool)
}

func (p *Player) isPlaying() bool {
	return p.playing
}

// IsWaiting checks if the mainloop is waiting on an episode to download
// before playing it.
func (p *Player) IsWaiting() bool {
	p.act <- reqWaiting

	r := <-p.dat
	return r.(bool)
}

func (p *Player) isWaiting() bool {
	return p.waiting
}

// Pause will stop audio playback, but preserve the position in the audio and
// queue. Will block until this is complete. Has no effect if not playing.
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

// Unpause will restore the position into the audio and queue from a pause
// and resume playback. Will block until this is complete. Has no effect if
// not paused or playing.
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

// Toggle pauses if the mainloop is unpaused, otherwise unpauses.
func (p *Player) Toggle() {
	p.act <- actToggle
}

func (p *Player) toggle() {
	if !p.playing {
		return
	}

	paused, _ := p.ctrl.Pause()
	p.ctrl.SetPause(!paused)
}

// GetTimings returns the current time and duration
// of the ongoing player. Returns zero if we are
// not playing currently.
//
// This function is thread safe but may block until
// data is available.
func (p *Player) GetTimings() (float64, float64) {
	p.act <- reqTimings

	dat := (<-p.dat).([2]float64)
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

// Seek moves the player head relative to the current
// position.
func (p *Player) Seek(off int) {
	p.act <- actSeek
	p.dat <- off
}

func (p *Player) seek(off int) {
	if !p.playing {
		return
	}

	p.ctrl.Seek(off, mpv.SeekModeRelative)
}

// Wait for the current episode to complete.
func (p *Player) Wait() {
	if !p.playing {
		return
	}

	now, _ := p.ctrl.Filename()
	for filename := now; filename == now; filename, _ = p.ctrl.Filename() {
		if !p.isPaused() {
			p.hndl.Post(ev.PlayerChanged)
		}
		time.Sleep(UpdateTime)
	}
}

func (p *Player) Event(e int) {
	switch e {
	case ev.DownloadChanged:
		select {
		case p.dlchan <- struct{}{}:
		default:
		}
	}
}

// Mainloop - Main sound handling thread loop
//
// Loops infinitely waiting for sound events, managing the queue, mpv
// processes and player data handling. Can be communicated with through
// a series of channels indicating different actions.
func Mainloop() {
	var wait WaitFunc
	var elem *data.QueueItem
	u := make(chan int)

	for {
		wait = updateWait
		elem, Plr.exhausted = PopHead()

		if !Plr.playing && !Plr.waiting && !Plr.exhausted && len(queue) > 0 {
			if elem.State != data.StatePending && data.Downloads.EntryExists(elem.Path) {
				Plr.play(elem)
				wait = endWait

				// Set status to played
				data.Q.Range(func(_ int, item *data.QueueItem) bool {
					if item.Path == elem.Path {
						item.State = data.StatePlayed
						data.Stamps.Touch(item.Path)
						return false
					}

					return true
				})
			} else {
				Plr.waiting = true

				if y, _ := data.Downloads.IsDownloading(elem.Path); y {
					Plr.download = elem
				} else {

					_, err := data.Downloads.Download(elem)
					if err != nil {
						continue
					}

					Plr.download = elem
				}

				wait = downloadWait
			}
		}

		go wait(u)
		keepWaiting := true
		for keepWaiting {
			select {
			case <-u:
				keepWaiting = false
			case e := <-Plr.event:
				Plr.Event(e)
			case action := <-Plr.act:
				switch action {
				case actTerm:
					if Plr.proc != nil {
						Plr.proc.Process.Kill()
					}

					Plr.playing = false

					Plr.dat <- 1
					return
				case actStop:
					Plr.stop()
				case actPause:
					Plr.pause()
				case actUnpause:
					Plr.unpause()
				case actToggle:
					Plr.toggle()
				case actSeek:
					dat := <-Plr.dat
					Plr.seek(dat.(int))

				case reqPaused:
					Plr.dat <- Plr.isPaused()
				case reqPlaying:
					Plr.dat <- Plr.isPlaying()
				case reqWaiting:
					Plr.dat <- Plr.isWaiting()
				case reqTimings:
					d, p := Plr.getTimings()
					arr := [2]float64{d, p}

					Plr.dat <- arr
				}
			}
		}
	}
}
