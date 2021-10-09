package data

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"regexp"
)

const (
	DB_BASENAME = "podbit"
	DB_FILENAME = "db"
)

// Error values
var (
	DatabaseIOFailed    error  = errors.New("Error: IO error while reading from database file")
	DatabaseSyntaxError string = "Error: Malformed database: Syntax Error on Line %d"
)

// Human-provided podcast info
type Podcast struct {
	RegexPattern string
	FriendlyName string
}

type Database struct {
	path string
	podcasts []Podcast
}

func initDatabase(db *Database) error {
	var err error
	file, err := os.Open(db.path)
	if err != nil {
		file, err = os.Create(db.path)
		if err != nil {
			return err
		}
	} else {
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		i := 1
		for scanner.Scan() {
			var p Podcast

			if scanner.Err() != nil {
				return DatabaseIOFailed
			}

			elem := scanner.Text()
			fields := strings.Split(elem, " ")
			num := len(fields)

			if num < 2 {
				return fmt.Errorf(DatabaseSyntaxError, i)
			}

			p.RegexPattern = fields[0]

			for i := 1; i < len(fields); i++ {
				p.FriendlyName += fields[i] + " "
			}

			db.podcasts = append(db.podcasts, p)

			i++
		}
	}

	return nil
}

func (db *Database) Open() error {
	data := os.Getenv("XDG_DATA_HOME")
	db.path = filepath.Join(data, DB_BASENAME, DB_FILENAME)

	// Ensure the database exists and is initialised
	err := initDatabase(db)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) Save() {
	file, err := os.OpenFile(db.path, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		fmt.Printf("WARNING: failed to write database: %s\n", DatabaseIOFailed)
	}

	for _, elem := range db.podcasts {
		fmt.Fprintf(file, "%s %s\n", elem.RegexPattern, elem.FriendlyName)
	}

	file.Close()
}

func (db *Database) GetFriendlyName(url string) string {
	for _, elem := range db.podcasts {
		matched, _ := regexp.MatchString(elem.RegexPattern, url)
		if matched {
			return elem.FriendlyName
		}
	}

	return url
}

func (db *Database) GetRegex(friendly string) string {
	for _, elem := range db.podcasts {
		if elem.FriendlyName == friendly {
			return elem.RegexPattern
		}
	}

	return ""
}
