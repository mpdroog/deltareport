package config

// Read config.json
import (
	"github.com/boltdb/bolt"
	"encoding/json"
	"os"
)

type Config struct {
	Files map[string]string
	Queues struct {
		Mail map[string]struct {
			From string
			To []string
			Subject string
		}
		Newline map[string]struct {
			To string
			Queue string
		}
	}
}

var (
	C           Config
	Verbose     bool
	DB          *bolt.DB
)

func Init(f string) error {
	r, e := os.Open(f)
	if e != nil {
		return e
	}
	if e := json.NewDecoder(r).Decode(&C); e != nil {
		return e
	}

	DB, e = bolt.Open("delta.db", 0600, nil)
    if e != nil {
        return e
    }
	return nil
}

func Close() error {
    return DB.Close()
}