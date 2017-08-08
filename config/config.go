package config

import (
	"github.com/BurntSushi/toml"
	"github.com/boltdb/bolt"
	"os"
	"time"
	"regexp"
	"fmt"
	"strings"
	"path/filepath"
)

type File struct {
	Path       string
	To         string
	Recurse    bool
	IncludeExt []string
	Regex      string
	Regexp     *regexp.Regexp
}
type Queue struct {
	User string
	Pass string
	Host string
	Port int
	To []string
	From string
	FromName string
	Subject string
}

type Config struct {
	Confdir string
	Files []File
	Queues map[string]Queue
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
	defer r.Close()
	if _, e := toml.DecodeReader(r, &C); e != nil {
		return fmt.Errorf("TOML: %s", e)
	}
	if e := loadConfDir(); e != nil {
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

	// test config
	for _, ln := range C.Files {
		if _, ok := C.Queues[ln.To]; !ok {
			return fmt.Errorf("File(%s) has non-existing Queue(To=%s)", ln.Path, ln.To)
		}
	}
	return nil
}

func loadConfDir() error {
	if len(C.Confdir) > 0 {
		return filepath.Walk(C.Confdir, func(path string, f os.FileInfo, err error) error {
			if path == C.Confdir {
				// ignore root
				return nil
			}

			if strings.HasSuffix(path, ".toml") {
				r, e := os.Open(path)
				if e != nil {
					return e
				}
				var f File
				if _, e := toml.DecodeReader(r, &f); e != nil {
					r.Close()
					return fmt.Errorf("TOML(%s): %s", path, e)
				}
				r.Close()
				C.Files = append(C.Files, f)
			}

			return nil
		})
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
