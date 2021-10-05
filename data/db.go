package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	DB_BASENAME = "podbit"
	DB_FILENAME = "db.json"
)

// Error values
var (
	DatabaseIOFailed    error  = errors.New("Error: IO error while reading from queue file")
	DatabaseSyntaxError string = "Error: Malformed database: %v"
)

// Human-provided podcast info
type Podcast struct {
	FriendlyName string
}

type DB struct {
	path string
	file *os.File

	podcasts map[string]Podcast
}

func initDatabase(db *DB) error {
	var err error
	file, err := os.OpenFile(db.path, os.O_RDWR, os.ModePerm)
	if err != nil {
		file, err = os.Create(db.path)
		if err != nil {
			return err
		}
	} else {

		buf, err := io.ReadAll(file)
		if err != nil {
			return DatabaseIOFailed
		}

		// Ignore null databases
		if len(buf) > 0 {
			err = json.Unmarshal(buf, &db.podcasts)
			if err != nil {
				return fmt.Errorf(DatabaseSyntaxError, err)
			}
		}
	}

	return nil
}

func (db *DB) Open() error {
	db.podcasts = make(map[string] Podcast)

	data := os.Getenv("XDG_DATA_HOME")
	db.path = filepath.Join(data, DB_BASENAME, DB_FILENAME)

	// Ensure the database exists and is initialised
	err := initDatabase(db)
	if err != nil {
		return err
	}

	db.file, err = os.OpenFile(db.path, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return DatabaseIOFailed
	}

	return nil
}

func (db *DB) Save() {
	data, err := json.Marshal(db.podcasts)
	if err != nil {
		fmt.Println("WARNING: Failed to encode data for saving. Bailing out...")
		return
	}

	db.file.Write(data)
	db.file.Close()
}
