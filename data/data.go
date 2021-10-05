package data

import (
	"fmt"
)

// Dependent data structures
var (
	queue Queue
	database DB
)

func InitData() error {
	defer fmt.Print("\n")

	fmt.Print("Reading queue...")
	err := queue.Open()
	if err != nil {
		return err
	}
	fmt.Println("done")

	return nil
}
