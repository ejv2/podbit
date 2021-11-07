package data

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dhowden/tag"
)

// Possible cache errors
var (
	ErrorIO             error  = errors.New("Error: Failed to create cache entry")
	ErrorDownloadFailed string = "Error: Failed to download from url %s"
)

// Cache is the current state of the on-disk cache and associated
// operations.
//
// This structure *is* thread safe, but ONLY is used with the
// correct methods. Use with care!
type Cache struct {
	dir string

	episodes sync.Map

	downloadsMutex sync.Mutex // Protects the below two variables
	downloads      []Download
	ongoing        int
}

// Episode represents the data extracted from a single cached episode
// media entry.
type Episode struct {
	Queued bool

	Title string
	Date  int
	Host  string
}

// Download represents the statistics of a specific ongoing download.
// Once the associated download is complete, the watcher goroutine
// terminates.
//
// This struct is *not* thread safe; don't write to it
type Download struct {
	// Path is the absolute path of the download destination
	Path string
	// File is the live file handle of the download
	// Will be closed once Completed == true
	File *os.File

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

// Dig through newsboat stuff to guess the download dir
// If we can't find it, just use the newsboat default and hope for the best
func (c *Cache) guessDir() string {
	conf, _ := os.UserConfigDir()
	p := filepath.Join(conf, "newsboat/config")

	file, err := os.Open(p)
	if err != nil {
		ret, _ := os.UserHomeDir()
		return ret
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Err() != nil {
			ret, _ := os.UserHomeDir()
			return ret
		}

		line := scanner.Text()
		fields := strings.Split(line, " ")

		if len(line) < 1 || len(fields) < 2 {
			continue
		}

		if fields[0] == "download-path" {
			return fields[1]
		}
	}

	ret, _ := os.UserHomeDir()
	return ret
}

// Open opens and initialises the cache
// Should be called once and once only - further modifications
// and cache mutations happen exclusively through other methods
func (c *Cache) Open() error {
	home, _ := os.UserHomeDir()
	c.dir = strings.ReplaceAll(c.guessDir(), "~", home)
	files, _ := ioutil.ReadDir(c.dir)

	for _, elem := range files {
		path := filepath.Join(c.dir, elem.Name())
		c.loadFile(path, true)
	}

	return nil
}

func (c *Cache) loadFile(path string, startup bool) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	defer func() {
		// Prevent invalid media files from causing a panic
		if p := recover(); p != nil {
			fmt.Printf("\nInvalid media file %q in cache! Ignoring...\n", path)
			return
		}
	}()

	data, err := tag.ReadFrom(file)
	if err != nil {
		fmt.Println("Error: Invalid media file in cache")
		return
	}

	artist, albumArtist := data.Artist(), data.AlbumArtist()
	var host string
	if artist == "" {
		host = albumArtist
	} else {
		host = artist
	}

	var ep Episode = Episode{
		Queued: !startup,
		Title:  data.Title(),
		Date:   data.Year(),
		Host:   host,
	}

	c.episodes.Store(path, ep)
}

// Download Starts a download and return its ID in the downloads table
// This can be used to retrieve information about said download by passing
// to GetDownload.
//
// Returns as soon as the download has been initialised - which could be
// significant. We recommend calling this function in a goroutine.
// Does not block until completion, but spawns two goroutines to
// complete the work as efficiently as possible.
func (c *Cache) Download(item *QueueItem) (id int, err error) {
	f, err := os.Create(item.Path)
	if err != nil {
		c.downloadsMutex.Lock()
		var dl Download = Download{
			Started:   time.Now(),
			Completed: true,
			Success:   false,
			Error:     "IO Error",
		}
		c.downloads = append(c.downloads, dl)
		id = len(c.downloads) - 1
		c.downloadsMutex.Unlock()

		return id, ErrorIO
	}

	resp, err := http.Get(item.URL)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.downloadsMutex.Lock()
		var dl Download = Download{
			Path:      item.Path,
			File:      f,
			Started:   time.Now(),
			Completed: true,
			Success:   false,
			Error:     "Download failed",
		}
		c.downloads = append(c.downloads, dl)
		id = len(c.downloads) - 1

		c.downloadsMutex.Unlock()

		os.Remove(item.Path)
		return id, fmt.Errorf(ErrorDownloadFailed, item.URL)
	}

	size, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)

	c.downloadsMutex.Lock()
	var dl Download = Download{
		Path:    item.Path,
		File:    f,
		Size:    size,
		Started: time.Now(),
		Stop:    make(chan int),
	}

	Q.mutex.Lock()
	item.State = StatePending
	Q.mutex.Unlock()

	c.ongoing++
	c.downloads = append(c.downloads, dl)
	id = len(c.downloads) - 1

	c.downloadsMutex.Unlock()

	go func(entry int) {
		var err error
		var count int64
		var read int
		var buf []byte = make([]byte, 32*1024) // 32kb
		for err == nil {
			read, err = resp.Body.Read(buf)
			f.WriteAt(buf, count)
			count += int64(read)

			c.downloadsMutex.Lock()
			c.downloads[entry].Done = count
			c.downloads[entry].Percentage = float64(c.downloads[entry].Done) / float64(c.downloads[entry].Size)
			c.downloadsMutex.Unlock()

			if c.Ongoing() > 1 {
				runtime.Gosched() // Give the other threads a turn
			}

			select {
			case <-c.downloads[entry].Stop:
				err = errors.New("Cancelled")
				break
			default:
			}
		}

		Q.mutex.Lock()
		item.State = StateReady
		Q.mutex.Unlock()

		c.downloadsMutex.Lock()
		c.downloads[entry].Completed = true
		if err != nil && err.Error() != "EOF" {
			c.downloads[entry].Success = false
			c.downloads[entry].Error = err.Error()
		} else {
			c.downloads[entry].Success = true
		}
		close(c.downloads[entry].Stop)

		c.loadFile(c.downloads[entry].Path, false)
		c.ongoing--
		c.downloadsMutex.Unlock()

		resp.Body.Close()
		f.Close()
	}(id)

	return
}

// IsDownloading queries the download cache to check
// if a podcast is currently downloading
func (c *Cache) IsDownloading(path string) (bool, int) {
	c.downloadsMutex.Lock()
	defer c.downloadsMutex.Unlock()

	for i, elem := range c.downloads {
		if elem.Path == path && !elem.Completed {
			return true, i
		}
	}

	return false, 0
}

// GetDownload returns the specified download in a thread-safely
// This should be used to get the details of a specified download
// via the ID
func (c *Cache) GetDownload(ind int) Download {
	c.downloadsMutex.Lock()
	defer c.downloadsMutex.Unlock()

	return c.downloads[ind]
}

// Ongoing returns the current number of ongoing downloads.
// The value cannot change while this function is executing.
func (c *Cache) Ongoing() int {
	c.downloadsMutex.Lock()
	defer c.downloadsMutex.Unlock()

	return c.ongoing
}

// Downloads returns all recorded downloads at this point,
// including completed or failed downloads.
// in time thread safely. No downloads can start or end
// while this function is executing.
func (c *Cache) Downloads() []Download {
	c.downloadsMutex.Lock()
	defer c.downloadsMutex.Unlock()

	return c.downloads
}

// Query returns cached data about an episode on disk
func (c *Cache) Query(path string) (ep Episode, ok bool) {
	e, ok := c.episodes.Load(path)
	if e != nil {
		ep = e.(Episode)
	}

	return
}

// QueryAll returns all known data about the on-disk cache
func (c *Cache) QueryAll(allowQueued bool) (e []Episode) {
	c.episodes.Range(func(key interface{}, value interface{}) bool {
		ep := value.(Episode)
		if (!allowQueued && !ep.Queued) || allowQueued {
			e = append(e, ep)
		}

		return true
	})

	return
}

// EntryExists searches the cache to determine if the entry exists
// Path should be an absolute path
// If path lies outside the cache dir, false is returned
func (c *Cache) EntryExists(path string) bool {
	f, err := os.Open(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	f.Close()
	return true
}
