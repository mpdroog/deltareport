package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/jinzhu/configor"
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
	// ContainerName name of docker container
	ContainerName string
	// ContainerNameStrict Search fo exact name
	ContainerNameStrict bool
}

type Config struct {
	Confdir   string
	Files     []File
	Scriptdir string

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
	Debug    bool
	DB       *bolt.DB
	Hostname string

	ScriptEngine *otto.Otto
	Scripts      map[string]*otto.Script
)

func Init(f string) error {
	var e error
	if e := configor.Load(&C, f); e != nil {
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

			if strings.HasSuffix(path, ".toml") || strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
				var f File
				if e := configor.Load(&f, path); e != nil {
					return fmt.Errorf("loading config (%s): %s", path, e)
				}

				// Build path for containers
				if len(f.ContainerName) > 0 {
					commandargs := fmt.Sprintf("name=%s", f.ContainerName)
					if f.ContainerNameStrict {
						commandargs = fmt.Sprintf("name=^/%s$", f.ContainerName)
					}
					id, e := exec.Command("docker", "ps", "--format='{{.ID}}'", "-f", commandargs).Output()
					// Cleanup of id
					container_id := strings.ReplaceAll(string(id), "\n", "")
					container_id = strings.ReplaceAll(container_id, "'", "")

					if e != nil || len(container_id) < 1 {
						if Verbose {
							fmt.Printf("can't find container %s\n", f.ContainerName)
						}
						return nil
					}

					lp, e := exec.Command("docker", "inspect", "--format='{{.LogPath}}'", container_id).Output()
					logpath := strings.ReplaceAll(string(lp), "\n", "")
					logpath = strings.ReplaceAll(logpath, "'", "")
					if e != nil || len(string(logpath)) < 1 {
						if Verbose {
							fmt.Printf("can't inspect container %s\n", f.ContainerName)
						}
						return nil
					}
					if len(logpath) > 0 {
						f.Path = string(logpath)
					}
				}
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
