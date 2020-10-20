package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/boltdb/bolt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"github.com/robertkrimen/otto"
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
	Confdir string
	Files   []File
	Scriptdir string

	Queues  struct {
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
	Debug    bool
	DB       *bolt.DB
	Hostname string

	ScriptEngine *otto.Otto
	Scripts map[string]*otto.Script
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

	Scripts = make(map[string]*otto.Script)
	if e := loadScriptDir(); e != nil {
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

func loadScriptDir() error {
	if len(C.Scriptdir) > 0 {
		ScriptEngine = otto.New()
		return filepath.Walk(C.Scriptdir, func(path string, f os.FileInfo, err error) error {
			if path == C.Scriptdir {
				// ignore root
				return nil
			}

			if strings.HasSuffix(path, ".js") {
				r, e := os.Open(path)
				if e != nil {
					return e
				}
				p, e := ScriptEngine.Compile(path, r)
				if e != nil {
					return e
				}
				r.Close()
				Scripts[path] = p
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
