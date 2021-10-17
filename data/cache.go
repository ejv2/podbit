package data

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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
	Downloads      []Download
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
// You should treat this struct as read only - including the contained
// file handle, despite write permissions being granted.
//
// Contained fields have no guarantees of thread safety: this struct is
// merely an approximation and a convenience.
//
// Path is the absolute path of the destination
// File is the current live handle of the download file. DO NOT TOUCH!
// Size is the size of the download to be completed
// Done is the size of the download which *has* completed
// Started is the timestamp of the download start
// Completed will be true when the nanny goroutine terminates
// Success will be true if the download completed with no errors
type Download struct {
	Path string
	File *os.File

	Percentage float64

	Size int64
	Done int64

	Started time.Time

	Completed bool
	Success   bool
}

func (c *Cache) progressWatcher(watch *Download, stop chan int) {
	for {
		c.downloadsMutex.Lock()

		fi, err := watch.File.Stat()
		if err != nil {
			return
		}
		watch.Done = fi.Size()
		watch.Percentage = float64(watch.Done) / float64(watch.Size)

		c.downloadsMutex.Unlock()

		select {
		case <-stop:
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
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
// This can be used to retrieve information about said download
//
// WARNING: DO NOT modify this table without taking out the mutex!
// This code is NOT THREAD SAFE witout this mutex being used. In most
// situations, use the entry as a read-only reference *only*.
//
// Returns as soon as the download has been initialised - which could be
// significant. We recommend calling this function in a goroutine.
// Does not block until completion, but spawns two goroutines to
// complete the work as efficiently as possible.
func (c *Cache) Download(item *QueueItem) (id int, err error) {

	f, err := os.Create(item.Path)
	if err != nil {
		return 0, ErrorIO
	}

	resp, err := http.Get(item.URL)
	if err != nil || resp.StatusCode != http.StatusOK {
		os.Remove(item.Path)
		return 0, fmt.Errorf(ErrorDownloadFailed, item.URL)
	}

	size, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	stop := make(chan int)

	c.downloadsMutex.Lock()
	var dl Download = Download{
		Path:    item.Path,
		File:    f,
		Size:    size,
		Started: time.Now(),
	}
	c.ongoing++
	c.Downloads = append(c.Downloads, dl)

	id = len(c.Downloads) - 1
	go c.progressWatcher(&c.Downloads[id], stop)
	c.downloadsMutex.Unlock()

	go func() {
		var err error
		var count int64
		var read int
		var buf []byte = make([]byte, 32*1024) // 32kb
		for err == nil {
			read, err = resp.Body.Read(buf)
			f.WriteAt(buf, count)
			count += int64(read)

			if c.ongoing > 1 {
				runtime.Gosched() // Give the other threads a turn
			}
		}

		if err != io.EOF {
			return
		}

		stop <- 1

		c.downloadsMutex.Lock()
		c.Downloads[id].Completed = true
		item.State = StateReady
		if err != nil {
			c.Downloads[id].Success = false
		} else {
			c.Downloads[id].Success = true
		}

		c.loadFile(c.Downloads[id].Path, false)
		c.ongoing--

		c.downloadsMutex.Unlock()

		resp.Body.Close()
		f.Close()
	}()

	return
}

// IsDownloading queries the download cache to check
// if a podcast is currently downloading
func (c *Cache) IsDownloading(path string) (bool, int) {
	c.downloadsMutex.Lock()
	defer c.downloadsMutex.Unlock()

	for i, elem := range c.Downloads {
		if elem.Path == path && !elem.Completed {
			return true, i
		}
	}

	return false, 0
}

// Ongoing returns the current number of ongoing downloads
// This value may be out of date after returned, but is
// thread safe.
func (c *Cache) Ongoing() int {
	c.downloadsMutex.Lock()
	defer c.downloadsMutex.Unlock()

	return c.ongoing
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
