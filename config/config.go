package config

// Read config.json
import (
	"github.com/boltdb/bolt"
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	Files map[string]struct {
		To string
		Recurse bool
		IncludeExt []string
	}
	Queues struct {
		Mail map[string]struct {
			Beanstalkd string
			From string
			To []string
			Subject string
		}
		Newline map[string]struct {
			Beanstalkd string
			Queue string
		}
	}
}

var (
	C           Config
	Verbose     bool
	DB          *bolt.DB
	Hostname    string
)

func Init(f string) error {
	r, e := os.Open(f)
	if e != nil {
		return e
	}
	if e := json.NewDecoder(r).Decode(&C); e != nil {
		return e
	}

	Hostname, e = os.Hostname()
	if e != nil {
		return e
	}

	DB, e = bolt.Open("delta.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
    if e != nil {
        return e
    }
	return nil
}

func Close() error {
    return DB.Close()
}