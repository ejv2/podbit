package data

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	CacheDBFilename = "cache.db"
	CacheDBComment  = `# This is the podbit cache.db file
# It contains the last consumption time for the listed media to allow for cache cleanouts
# Do not modify by hand`
)

var (
	ErrDBIO     = errors.New("Error: IO error while reading from cache.db")
	ErrDBIOW    = errors.New("Error: IO error while writing to cache.db")
	ErrDBSyntax = errors.New("Error: Syntax error in cache.db")
	ErrDBExists = errors.New("entry alredy exists")
	ErrDBEnoent = errors.New("no such entry in cache.db")
	ErrDBPruned = errors.New("entry has been marked for pruning")
)

// CacheSyntaxError is a syntax error which contains a line reference.
type CacheSyntaxError struct {
	Line    uint
	Comment string
}

func (c CacheSyntaxError) Error() string {
	return ErrDBSyntax.Error() + ": line " + strconv.FormatUint(uint64(c.Line), 10) + ": " + c.Comment
}

func (c CacheSyntaxError) Unwrap() error {
	return ErrDBSyntax
}

// A CacheEntry is a single entry in the db map contained within a CacheDB. It
// is made up of a resume timecode and a last played/finished timestamp.
type CacheEntry struct {
	// unix epoch time of an entry
	// set to <0 to indicate pruned entry
	finished int64
	// number of seconds in to the media file which we should resume at
	// if zero, start from the beginning (obviously!), but also is excluded
	// from the cache.db file
	resume uint64
}

// The CacheDB contains the timestamps which specify when media was last played
// or finished. This is used to avoid the media downloads directory becoming
// bigger and bigger as more and more podcasts are downloaded.
//
// The format for the CacheDB is simply the download filepath followed by the
// timestamp which would have been stored after the last field in the queue
// file. The timestamp is simply a 64-bit unix timestamp, however negative
// values are interpreted as pruned items (i.e items cleaned via cache cleanup)
// and will be excluded from deserialization. I doubt that anybody will have
// listen times in the 1960s.
//
// The cache.db is assumed to be under the exclusive control of podbit and as
// such is not reloaded during operation and may only be
// serialized/deserialized at program entry and exit.
//
// Prior to Podbit v4.0, this was an extra field on the queue file, which broke
// compatibility with podboat, which is obviously undesirable.
type CacheDB struct {
	mut *sync.RWMutex
	// db maps episode paths to timestamps
	db map[string]CacheEntry
}

func NewCacheDB() *CacheDB {
	return &CacheDB{
		new(sync.RWMutex),
		make(map[string]CacheEntry),
	}
}

func (c *CacheDB) Open() error {
	data := os.Getenv("XDG_DATA_HOME")
	path := filepath.Join(data, DatabaseDirname, CacheDBFilename)

	f, err := os.Open(path)
	if err != nil {
		fo, err := os.Create(path)
		if err != nil {
			return err
		}

		fo.WriteString(CacheDBComment)
		fo.Close()
		return nil
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	i := uint(1)
	for scanner.Scan() {
		if scanner.Err() != nil {
			return ErrDBIO
		}

		elem := scanner.Text()
		// Ignore comments and blank lines
		if len(elem) == 0 || strings.HasPrefix(elem, "#") {
			continue
		}

		fields := strings.Fields(elem)

		if len(fields) < 2 {
			return CacheSyntaxError{i, "insufficient fields (expect 2/3)"}
		}

		stamp, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return CacheSyntaxError{i, "parsing timestamp: " + err.Error()}
		}

		resume := uint64(0)
		if len(fields) == 3 {
			r, err := strconv.ParseUint(fields[2], 10, 64)
			if err != nil {
				return CacheSyntaxError{i, "parsing resume timecode: " + err.Error()}
			}

			resume = r
		}

		// If we have duplicates somehow take the later stamp.
		s, ok := c.db[fields[0]]
		if ok {
			if s.finished >= stamp {
				continue
			}
		}
		c.db[fields[0]] = CacheEntry{stamp, resume}

		i++
	}

	return nil
}

// Save truncates the cache.db to zero bytes before zerializing the in-memory
// database to the file in the accepted format.
func (c *CacheDB) Save() error {
	data := os.Getenv("XDG_DATA_HOME")
	path := filepath.Join(data, DatabaseDirname, CacheDBFilename)

	f, err := os.Create(path)
	if err != nil {
		return ErrDBIOW
	}
	defer f.Close()

	f.WriteString(CacheDBComment + "\n\n")

	c.mut.RLock()
	defer c.mut.RUnlock()

	for url, ts := range c.db {
		// Pruned entries
		if ts.finished < 0 {
			continue
		}

		sts := strconv.FormatInt(ts.finished, 10)
		res := strconv.FormatUint(ts.resume, 10)

		if ts.resume != 0 {
			f.WriteString(url + " " + sts + " " + res + "\n")
			continue
		}
		f.WriteString(url + " " + sts + "\n")
	}

	return nil
}

// Touch updates the listen time for an episode to the current timestamp. This
// method refuses to update the timestamp for any file which does not exist, so
// this should be checked first.
func (c *CacheDB) Touch(path string) error {
	// Refuse to touch a non existent path
	_, err := os.Stat(path)
	if err != nil {
		return err
	}

	c.mut.Lock()
	defer c.mut.Unlock()

	c.db[path] = CacheEntry{time.Now().Unix(), 0}
	return nil
}

// Resume is like Touch, except it sets the resume timecode to the given value.
// Normal touches implicitly reset the resume timecode back to zero, so ensure
// that this is called after any touch calls, if they should be made.
func (c *CacheDB) Resume(path string, rt uint64) error {
	// Refuse to touch a non existent path
	_, err := os.Stat(path)
	if err != nil {
		return err
	}

	c.mut.Lock()
	defer c.mut.Unlock()

	orig := c.db[path]
	c.db[path] = CacheEntry{orig.finished, rt}
	return nil
}

// Insert inserts a new path with the given timestamp into the map. Panics if
// ts < 0.
func (c *CacheDB) Insert(path string, ts int64) error {
	if ts < 0 {
		panic("invalid timestamp (<0): cannot insert pruned item")
	}

	c.mut.Lock()
	defer c.mut.Unlock()

	_, ok := c.db[path]
	if ok {
		return ErrDBExists
	}

	c.db[path] = CacheEntry{ts, 0}
	return nil
}

// Prune marks an entry as having been pruned, which excludes it from being
// saved at the next cache.db save.
func (c *CacheDB) Prune(path string) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	s, ok := c.db[path]
	if !ok {
		return ErrDBEnoent
	}

	// Refuse to prune more than once - obviously erroneous
	if s.finished < 0 {
		return ErrDBPruned
	}

	// Mark as pruned with negative timestamp
	c.db[path] = CacheEntry{-1, 0}
	return nil
}

// RawStat returns the currently recorded raw timestamp for an entry. This can
// fail if an entry does not exist, or if the entry was marked for pruning.
func (c *CacheDB) RawStat(path string) (*int64, *uint64, error) {
	c.mut.RLock()
	defer c.mut.RUnlock()

	s, ok := c.db[path]
	if !ok {
		return nil, nil, ErrDBEnoent
	}

	if s.finished < 0 {
		return nil, nil, ErrDBPruned
	}

	return &s.finished, &s.resume, nil
}

// Stat returns the currently recorded timestamp for an entry. This can fail if
// an entry does not exist, or if the entry was marked for pruning.
func (c *CacheDB) Stat(path string) (*time.Time, *uint64, error) {
	s, r, err := c.RawStat(path)
	if err != nil {
		return nil, nil, err
	}

	if s == nil {
		panic("incorrect rawstat: stamp ptr and err should never both be nil")
	}

	t := time.Unix(*s, 0)
	return &t, r, nil
}
