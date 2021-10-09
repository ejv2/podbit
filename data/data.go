package data

import (
	"fmt"
	"net/url"
	"time"
)

const (
	// QueueReloadInterval is how often the queue will be reloaded
	QueueReloadInterval = time.Duration(30) * time.Second
)

// Dependent data structures
var (
	Q       Queue
	DB      Database
	Caching Cache
)

// InitData initialises all dependent data structures
// The only returned errors *will* be fatal to the program
func InitData() error {
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
	err = Caching.Open()
	if err != nil {
		return err
	}
	fmt.Println("done")

	return nil
}

// SaveData cleans up and saves data to disk
// First ensures we have hot-reloaded any required data
func SaveData() {
	ReloadData()

	Q.Save()
	DB.Save()
}

// ReloadData performs a hot-reload of any data which can/needs
// to be hot reloaded
//
// This is called automatically on an interval by ReloadLoop
// and upon saving to ensure up-to-date data
func ReloadData() {
	Q.Reload()
}

// ReloadLoop is an infinite loop to continually reload the 
// file on disk into memory.
//
// Should allow us to hot-reload the queue file - among other
// things.
func ReloadLoop() {
	for {
		ReloadData()
		time.Sleep(QueueReloadInterval)
	}
}

// IsURL returns true if a string is a valid HTTP(s) URL
func IsURL(check string) bool {
	u, err := url.Parse(check)
	return err == nil && u.Scheme != "" && u.Host != ""
}
