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
	ErrDBSyntax = errors.New("Error: Syntax error in cache.db")
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
	db map[string]int64
}

func NewCacheDB() *CacheDB {
	return &CacheDB{
		new(sync.RWMutex),
		make(map[string]int64),
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
		// Ignore comments
		if strings.HasPrefix(elem, "#") {
			continue
		}

		fields := strings.Fields(elem)

		if len(fields) < 2 {
			return CacheSyntaxError{i, "insufficient fields (expect 2)"}
		}

		stamp, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return CacheSyntaxError{i, "parsing timestamp: " + err.Error()}
		}

		// If we have duplicates somehow take the later stamp.
		s, ok := c.db[fields[0]]
		if ok {
			if s >= stamp {
				continue
			}
		}
		c.db[fields[0]] = stamp

		i++
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

	c.db[path] = time.Now().Unix()
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
	if s < 0 {
		return ErrDBPruned
	}

	// Mark as pruned with negative timestamp
	c.db[path] = -1
}

// RawStat returns the currently recorded raw timestamp for an entry. This can
// fail if an entry does not exist, or if the entry was marked for pruning.
func (c *CacheDB) RawStat(path string) (*int64, error) {
	c.mut.RLock()
	defer c.mut.RUnlock()

	s, ok := c.db[path]
	if !ok {
		return nil, ErrDBEnoent
	}

	if s < 0 {
		return nil, ErrDBPruned
	}

	return &s, nil
}

// Stat returns the currently recorded timestamp for an entry. This can fail if
// an entry does not exist, or if the entry was marked for pruning.
func (c *CacheDB) Stat(path string) (*time.Time, error) {
	s, err := c.RawStat(path)
	if err != nil {
		return nil, err
	}

	if s == nil {
		panic("incorrect rawstat: stamp ptr and err should never both be nil")
	}

	t := time.Unix(*s, 0)
	return &t, nil
}
