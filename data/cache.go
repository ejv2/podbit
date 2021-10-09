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
	CacheIOError        error  = errors.New("Error: Failed to create cache entry")
	CacheDownloadFailed string = "Error: Failed to download from url %s"
)

type Cache struct {
	dir string

	episodes sync.Map

	downloadsMutex sync.Mutex // Protects the below two variables
	Downloads      []Download
	ongoing        int
}

type Episode struct {
	Queued bool

	Title string
	Date  int
	Host  string
}

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

// Open and initialise the cache
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

// Start a download and return its ID in the downloads table
// This can be used to retrieve information about said download
//
// WARNING: DO NOT modify this table without taking out the mutex
// This code is NOT THREAD SAFE witout this mutex being used
//
// Returns as soon as the download has been initialised
// Does not block until completion
func (c *Cache) Download(item *QueueItem) (id int, err error) {

	f, err := os.Create(item.Path)
	if err != nil {
		return 0, CacheIOError
	}

	resp, err := http.Get(item.Url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf(CacheDownloadFailed, item.Url)
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

func (c *Cache) Query(path string) (ep Episode, ok bool) {
	e, ok := c.episodes.Load(path)
	if e != nil {
		ep = e.(Episode)
	}

	return
}

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
