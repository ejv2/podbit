package data

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dhowden/tag"
	lcss "github.com/vmarkovtsev/go-lcss"

	ev "github.com/ejv2/podbit/event"
)

// Possible cache errors.
var (
	ErrorIO             = errors.New("Error: Failed to create cache entry")
	ErrorCreation       = errors.New("Error: Download directory did not exist and could not be created")
	ErrorDownloadFailed = "Error: Failed to download from url %s"
)

// Cache is the current state of the on-disk cache and associated
// operations.
//
// This structure *is* thread safe, but ONLY is used with the
// correct methods. Use with care!
type Cache struct {
	episodes sync.Map

	downloadsMutex sync.RWMutex // Protects the below two variables
	downloads      []*Download
	ongoing        int

	hndl ev.Handler
}

// Episode represents the data extracted from a single cached episode
// media entry.
type Episode struct {
	Queued bool

	Title string
	Date  int
	Host  string
}

// Dig through newsboat stuff to guess the download dir.
// If we can't find it, just use the newsboat default and hope for the best.
func (c *Cache) guessDir(queueEntries []QueueItem) string {
	paths := make([][]byte, 0, len(queueEntries))
	for _, entry := range queueEntries {
		paths = append(paths, []byte(entry.Path))
	}

	bcom := lcss.LongestCommonSubstring(paths...)
	com := string(bcom)
	if com != "" {
		return com
	}

	ret, _ := os.UserHomeDir()
	return ret
}

// Open opens and initialises the cache.
// Should be called once and once only - further modifications
// and cache mutations happen exclusively through other methods.
func (c *Cache) Open(q *Queue, hndl ev.Handler) error {
	if q == nil {
		panic("open cache: nil queue passed")
	}
	c.hndl = hndl

	for _, elem := range q.Items {
		c.loadFile(elem.Path, true)
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
		fmt.Printf("\nError: Invalid media file %q in cache! Ignoring...\n", path)
		return
	}

	artist, albumArtist := data.Artist(), data.AlbumArtist()
	var host string
	if artist == "" {
		host = albumArtist
	} else {
		host = artist
	}

	ep := Episode{
		Queued: !startup,
		Title:  data.Title(),
		Date:   data.Year(),
		Host:   host,
	}

	c.episodes.Store(path, ep)
}

// Download starts an asynchronous download in a new goroutine. Returns the ID
// in the downloads table, which must be accessed using a mutex. Item passed
// should be locked by the caller prior to calling.
func (c *Cache) Download(item *QueueItem) (id int, err error) {
	dir := filepath.Dir(item.Path)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		dl := Download{
			mut:       new(sync.RWMutex),
			Path:      item.Path,
			File:      nil,
			Elem:      item,
			Started:   time.Now(),
			Completed: true,
			Success:   false,
			Error:     "Directory IO Error",
			Stop:      nil,
		}

		c.downloadsMutex.Lock()
		c.downloads = append(c.downloads, &dl)
		id = len(c.downloads) - 1
		c.downloadsMutex.Unlock()

		return id, ErrorIO
	}

	f, err := os.Create(item.Path)
	dl := Download{
		mut:     new(sync.RWMutex),
		Path:    item.Path,
		File:    f,
		Elem:    item,
		Started: time.Now(),
		Stop:    make(chan int),
	}

	if err != nil {
		dl = Download{
			mut:       new(sync.RWMutex),
			Path:      item.Path,
			File:      f,
			Elem:      item,
			Started:   time.Now(),
			Completed: true,
			Success:   false,
			Error:     "IO Error",
			Stop:      nil,
		}

		c.downloadsMutex.Lock()
		c.downloads = append(c.downloads, &dl)
		id = len(c.downloads) - 1
		c.downloadsMutex.Unlock()

		return id, ErrorIO
	}

	if item.Youtube {
		go dl.DownloadYoutube(c.hndl)
	} else {
		go dl.DownloadHTTP(c.hndl)
	}

	c.downloadsMutex.Lock()
	c.downloads = append(c.downloads, &dl)
	id = len(c.downloads) - 1
	c.ongoing++
	c.downloadsMutex.Unlock()

	return
}

// IsDownloading queries the download cache to check.
// if a podcast is currently downloading.
func (c *Cache) IsDownloading(path string) (bool, int) {
	c.downloadsMutex.RLock()
	defer c.downloadsMutex.RUnlock()

	for i, elem := range c.downloads {
		elem.mut.RLock()
		if elem.Path == path && !elem.Completed {
			elem.mut.RUnlock()
			return true, i
		}
		elem.mut.RUnlock()
	}

	return false, 0
}

// GetDownload returns the specified download in a thread-safely.
// This should be used to get the details of a specified download
// via the ID.
func (c *Cache) GetDownload(ind int) Download {
	c.downloads[ind].mut.RLock()
	defer c.downloads[ind].mut.RUnlock()

	return *c.downloads[ind]
}

// Ongoing returns the current number of ongoing downloads.
// The value cannot change while this function is executing.
func (c *Cache) Ongoing() int {
	c.downloadsMutex.RLock()
	defer c.downloadsMutex.RUnlock()

	return c.ongoing
}

// Downloads returns all recorded downloads at this point,
// including completed or failed downloads.
// No downloads can start or end while this function is
// executing.
func (c *Cache) Downloads() []Download {
	c.downloadsMutex.RLock()
	defer c.downloadsMutex.RUnlock()

	dls := make([]Download, len(c.downloads))

	for i, elem := range c.downloads {
		elem.mut.RLock()
		dls[i] = *elem
		elem.mut.RUnlock()
	}

	return dls
}

// Query returns cached data about an episode on disk.
func (c *Cache) Query(path string) (ep Episode, ok bool) {
	e, ok := c.episodes.Load(path)
	if e != nil {
		ep = e.(Episode)
	}

	return
}

// QueryAll returns all known data about the on-disk cache.
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
// If path lies outside the cache dir, false is returned.
func (c *Cache) EntryExists(path string) bool {
	f, err := os.Open(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	f.Close()
	return true
}
