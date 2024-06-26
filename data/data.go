// Package data implements data loading, management and serialisation.
// It maintains a set of singleton instances of the data sources which
// should be used directly by clients.
//
// Although most of the data sources are designed to be thread safe,
// some, such as the cache, cannot guarantee this in all use cases.
package data

import (
	"fmt"
	"net/url"
	"os"
	"time"

	ev "github.com/ejv2/podbit/event"
)

const (
	// QueueReloadInterval is how often the queue will be reloaded.
	QueueReloadInterval = time.Minute
	// EpisodeCacheTime is how long an episode is allowed to stay in cache in seconds.
	// Default value is three days (3 * 24 * 60 * 60).
	EpisodeCacheTime = 259200
)

// Queue reload operations.
const (
	QueueReload = iota
	QueueSave
)

// Dependent data structures.
var (
	Q         Queue
	DB        Database
	Downloads Cache
)

// InitData initialises all dependent data structures.
// The only returned errors *will* be fatal to the program.
func InitData(hndl ev.Handler) error {
	fmt.Print("Reading queue...")
	err := Q.Open()
	if err != nil {
		return err
	}
	fmt.Println("done")

	fmt.Print("Reading database...")
	err = DB.Open()
	if err != nil {
		return err
	}
	fmt.Println("done")

	fmt.Print("Initialising cache...")
	err = Downloads.Open(hndl)
	if err != nil {
		return err
	}
	fmt.Println("done")

	return nil
}

// SaveData cleans up and saves data to disk. First ensures we have
// hot-reloaded any required data. Only designed for use at startup.
func SaveData() {
	ReloadData()
	CleanData()

	Q.Save()
	DB.Save()
}

// ReloadData performs a hot-reload of any data which can/needs
// to be hot reloaded.
//
// This is called automatically on an interval by ReloadLoop
// and upon saving to ensure up-to-date data.
func ReloadData() {
	Q.Reload()
}

// CleanData cleans out the cache based on items which are both finished/played
// and with a last listen time of more than EpisodeCacheTime seconds ago
// (defaults to three days). Removed episodes are set to "pending" status (to
// be downloaded) and have their cache file removed.
func CleanData() {
	fmt.Printf("Cache cleanup...")

	now := time.Now().Unix()
	count := 0

	Q.Range(func(i int, item *QueueItem) bool {
		if item.Date == -1 || (item.State != StatePlayed && item.State != StateFinished) {
			return true
		}

		diff := now - item.Date
		if diff >= EpisodeCacheTime {
			item.Date = -1
			item.State = StatePending

			os.Remove(item.Path)
			count++
		}

		return true
	})

	fmt.Printf("done (removed %d items)\n", count)
}

// ReloadLoop is an infinite loop to continually reload the
// file on disk into memory.
//
// Should allow us to hot-reload the queue file - among other
// things.
func ReloadLoop(upchan chan int8) {
	ticker := time.NewTicker(QueueReloadInterval)
	count := 0
	defer ticker.Stop()

loop:
	for {

		select {
		case <-ticker.C:
			ReloadData()
			if count == 3 {
				Q.Save()
				count = 0
				continue
			}
			count++
		case i, ok := <-upchan:
			if !ok {
				break loop
			}

			ReloadData()
			if i == QueueSave {
				Q.Save()
			}
		}
	}
}

// IsURL returns true if a string is a valid HTTP(s) URL.
func IsURL(check string) bool {
	u, err := url.Parse(check)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// LimitString limits a UTF-8 string to max visible runes, showing the full
// string if it is shorter than max. If max is negative or zero, an empty
// string is returned.
func LimitString(in string, max int) string {
	r := []rune(in)
	if max <= 0 {
		return ""
	}
	if max > len(r) {
		return in
	}

	return string(r[:max])
}

// FormatTime formats a time measured in seconds.
func FormatTime(seconds float64) string {
	round := int(seconds)

	s := round % 60
	m := round / 60
	h := m / 60
	m = m % 60

	return fmt.Sprintf("%.2d:%.2d:%.2d", h, m, s)
}
