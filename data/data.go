package data

import (
	"fmt"
	"net/url"
	"time"
)

const (
	RELOAD_INTERVAL = time.Duration(30) * time.Second
)

// Dependent data structures
var (
	Q       Queue
	DB      Database
	Caching Cache
)

// Set up all data sources
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

// Clean up and save data to disk
// Will first reload data to merge them
func SaveData() {
	ReloadData()

	Q.Save()
	DB.Save()
}

// Reload and merge data from disk into memory
func ReloadData() {
	Q.Reload()
}

// Infinite loop to continually reload the file on disk into memory
// Should allow us to hot-reload the queue file
func ReloadLoop() {
	for {
		ReloadData()
		time.Sleep(RELOAD_INTERVAL)
	}
}

// Returns true is a string is a valid HTTP(s) URL
func IsUrl(check string) bool {
	u, err := url.Parse(check)
	return err == nil && u.Scheme != "" && u.Host != ""
}
