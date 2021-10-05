package data

import (
	"fmt"
	"time"
)

const (
	RELOAD_INTERVAL = time.Duration(30) * time.Second
)

// Dependent data structures
var (
	queue Queue
	database DB
)

// Set up all data sources
func InitData() error {
	defer fmt.Print("\n")

	fmt.Print("Reading queue...")
	err := queue.Open()
	if err != nil {
		return err
	}
	fmt.Println("done")

	fmt.Print("Reading database...")
	err = database.Open()
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

	database.Save()
}

// Reload and merge data from disk into memory
func ReloadData() {
	queue.Reload()
}

// Infinite loop to continually reload the file on disk into memory
// Should allow us to hot-reload the queue file
func ReloadLoop() {
	for {
		ReloadData()
		time.Sleep(RELOAD_INTERVAL)
	}
}
