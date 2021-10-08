package data

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"os"
	"io"
	"path/filepath"
	"strings"
)

// Possible cache errors
var (
	CacheIOError        error  = errors.New("Error: Failed to create cache entry")
	CacheDownloadFailed string = "Error: Failed to download from url %s"
)

type Cache struct {
	dir string

	episodes map[string]Episode
}

type Episode struct {
	entry *QueueItem

	title string
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
	c.dir = c.guessDir()

	return nil
}

func (c *Cache) Download(item *QueueItem) error {
	f, err := os.Create(item.Path)
	if err != nil {
		return CacheIOError
	}
	defer f.Close()

	resp, err := http.Get(item.Url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf(CacheDownloadFailed, item.Url)
	}
	defer resp.Body.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return CacheIOError
	}

	c.episodes[item.Path] = Episode{
		entry: item,
	}

	return nil
}
