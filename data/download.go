package data

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// YouTube downloading constants
const (
	YoutubeDL    string = "youtube-dl"
	YoutubeDLP   string = "yt-dlp"
	YoutubeFlags string = "--add-metadata --newline --no-colors -f bestaudio --extract-audio --audio-format mp3"
)

// Download represents the statistics of a specific ongoing download.
// Once the associated download is complete, the watcher goroutine
// terminates.
//
// This struct may only be modified through associated methods
type Download struct {
	// Protects this download instance
	mut sync.RWMutex

	// Path is the absolute path of the download destination
	Path string
	// File is the live file handle of the download
	// Will be closed once Completed == true
	File *os.File
	// Elem is the associated queue element
	Elem *QueueItem

	// Percentage is the calculated percentage currently completed
	Percentage float64

	// Size is the total size to download
	Size int64
	// Done is the currently downloaded size present on disk
	Done int64

	// Started is the timestamp of the download commencing
	Started time.Time

	// Completed == true once the operations has either finished or failed
	Completed bool
	// Success == true if the full download completed successfully
	Success bool
	// Error is the error which caused the download to fail
	// Empty if the download did not fail
	Error string

	// Stop will cause the download to cease immediately
	// Will be closed once the download completes
	Stop chan int
}

// DownloadYoutube selects an appropriate downloader (yt-dlp or
// youtube-dl) and begins a YouTube download on the calling thread
// (synchronously).
//
// Used internally by cache; avoid calling directly
func (d *Download) DownloadYoutube() {
	if !d.Elem.Youtube {
		panic("download: downloading non-youtube with youtube-dl")
	}

	// Work around "already downloaded" errors from youtube-dl
	d.File.Close()
	os.Remove(d.File.Name())

	defer func() {
		Downloads.downloadsMutex.Lock()
		Downloads.ongoing--
		Downloads.downloadsMutex.Unlock()
	}()
	defer close(d.Stop)

	// Determine downloader program - use yt-dlp if available, else use ytdl
	loader := ""
	if _, err := exec.LookPath(YoutubeDLP); err == nil {
		loader = YoutubeDLP
	} else if _, err := exec.LookPath(YoutubeDL); err == nil {
		loader = YoutubeDL
	} else {
		d.mut.Lock()
		d.Completed = true
		d.Success = false
		d.Error = "No YouTube downloader"
		d.mut.Unlock()

		return
	}

	h, _ := os.UserHomeDir()
	tmppath := filepath.Join(h, "podbit-ytdl"+strconv.FormatInt(time.Now().UnixMicro(), 10))
	flags := append(strings.Split(YoutubeFlags, " "), "-o", tmppath+".%(ext)s", d.Elem.URL)

	proc := exec.Command(loader, flags...)
	r, err := proc.StdoutPipe()
	defer r.Close()
	if err != nil {
		d.mut.Lock()
		d.Completed = true
		d.Success = false
		d.Error = "Downloader IO Error"
		d.mut.Unlock()

		return
	}
	proc.Start()

	// NOTE: Do not need to runtime.Gosched here to prevent contention, because IO waiting
	// occurs on this thread and the runtime will implicitly call it for us when there's no
	// pending data. Manually calling causes net performance loss

	buf := make([]byte, 4096)
	for err == nil {
		_, err = r.Read(buf)
		line := string(buf)
		fields := strings.Fields(line)

		if fields[0] != "[download]" || len(fields) < 2 {
			continue
		}

		d.mut.Lock()
		d.Percentage, _ = strconv.ParseFloat(fields[1][:len(fields[1])-1], 64)
		d.Percentage /= 100
		d.mut.Unlock()

		select {
		case <-d.Stop:
			d.mut.Lock()
			d.Error = "Cancelled"
			d.Completed = true
			d.Success = false
			d.mut.Unlock()

			proc.Process.Kill()
			return
		default:
		}
	}

	if err != nil && err.Error() != "EOF" {
		d.mut.Lock()
		d.Completed = true
		d.Success = false
		d.Error = "Downloader IO Error"
		d.mut.Unlock()
		return
	}

	// Move from temp location
	os.Rename(tmppath+".mp3", d.Path)

	// Final clean up
	Q.mutex.Lock()
	d.Elem.State = StateReady
	Q.mutex.Unlock()

	d.mut.Lock()
	d.Completed = true
	d.Success = true

	Downloads.downloadsMutex.Lock()
	Downloads.loadFile(d.Elem.Path, false)
	Downloads.downloadsMutex.Unlock()

	d.mut.Unlock()
	return
}

// DownloadHTTP connects to the URL of the specified download
// and downloads to download path on the calling thread
// (synchronously)
//
// Used internally by cache; avoid calling directly
func (d *Download) DownloadHTTP() {
	resp, err := http.Get(d.Elem.URL)
	if err != nil || resp.StatusCode != http.StatusOK {
		d.mut.Lock()
		d.Completed = true
		d.Success = false
		d.Error = "Download failed"
		d.mut.Unlock()

		Downloads.downloadsMutex.Lock()
		Downloads.ongoing--
		Downloads.downloadsMutex.Unlock()

		os.Remove(d.Elem.Path)
		return
	}

	d.mut.Lock()
	size, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	d.Size = size
	d.mut.Unlock()

	Q.mutex.Lock()
	d.Elem.State = StatePending
	Q.mutex.Unlock()

	var count int64
	var read int
	dlerr := ""
	buf := make([]byte, 32*1024) // 32kb

outer:
	for err == nil {
		read, err = resp.Body.Read(buf)
		d.File.WriteAt(buf, count)
		count += int64(read)

		d.mut.Lock()
		d.Done = count
		d.Percentage = float64(d.Done) / float64(d.Size)
		d.mut.Unlock()

		if Downloads.Ongoing() > 1 {
			runtime.Gosched() // Give the other threads a turn
		}

		select {
		case <-d.Stop:
			dlerr = "Cancelled"
			break outer
		default:
		}
	}

	d.mut.Lock()
	d.Completed = true
	if (err != nil && err.Error() != "EOF") || dlerr != "" {
		d.Success = false
		d.Error = dlerr
	} else {
		d.Success = true

		Q.mutex.Lock()
		d.Elem.State = StateReady
		Q.mutex.Unlock()
	}
	d.mut.Unlock()

	Downloads.downloadsMutex.Lock()
	Downloads.loadFile(d.Elem.Path, false)
	Downloads.ongoing--
	Downloads.downloadsMutex.Unlock()

	resp.Body.Close()
	d.File.Close()

	close(d.Stop)
	return
}
