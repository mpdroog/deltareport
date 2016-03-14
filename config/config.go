package config

// Read config.json
import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"os"
	"time"
	"regexp"
)

type File struct {
	Path       string
	To         string
	Tags       []string
	Recurse    bool
	IncludeExt []string
	Regex      string
	Regexp     *regexp.Regexp
	Linediff   bool
}

type Config struct {
	Files []File
	Queues struct {
		Mail map[string]struct {
			Beanstalkd string
			From       string
			To         []string
			Subject    string
		}
		Newline map[string]struct {
			Beanstalkd string
			Queue      string
		}
	}
	Db string
}

var (
	C        Config
	Verbose  bool
	DB       *bolt.DB
	Hostname string
)

func Init(f string) error {
	r, e := os.Open(f)
	if e != nil {
		return e
	}
	if e := json.NewDecoder(r).Decode(&C); e != nil {
		return e
	}
	if e := prepareRegexp(); e != nil {
		return e
	}

	Hostname, e = os.Hostname()
	if e != nil {
		return e
	}

	DB, e = bolt.Open(C.Db, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if e != nil {
		return e
	}
	return nil
}

func prepareRegexp() error {
	var e error
	for idx, file := range C.Files {
		if len(file.Regex) > 0 {
			file.Regexp, e = regexp.Compile(file.Regex)
			if e != nil {
				return e
			}
			C.Files[idx] = file
		}
	}

	return nil
}

func Close() error {
	return DB.Close()
}
