package data

import (
	"fmt"
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
// TODO: Reload on-disk data before saving to disk
func SaveData() {
	database.Save()
}
