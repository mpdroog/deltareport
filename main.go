package main

import (
	"fmt"
	"deltareport/config"
	"flag"
	"github.com/boltdb/bolt"
	"os"
	"strconv"
	"strings"
)

func main() {
	configPath := ""
	flag.BoolVar(&config.Verbose, "v", false, "Verbose-mode (log more)")
	flag.StringVar(&configPath, "c", "./config.json", "Path to config.json")
	flag.Parse()

	if e := config.Init(configPath); e != nil {
		panic(e)
	}
	for _, fileName := range config.C.Files {
		// TODO: missing err?
	    config.DB.Update(func(tx *bolt.Tx) error {
			bucket, e := tx.CreateBucketIfNotExists([]byte("filepos"))
		    if e != nil {
		        return fmt.Errorf("create bucket: %s", e.Error())
		    }

		    start := 0
		    val := bucket.Get([]byte(fileName))
		    if val != nil {
		    	i, e := strconv.Atoi(string(val))
		    	if e != nil {
		    		return e
		    	}
		    	start = i
		    }

		    file, e := os.Open(fileName)
		    if e != nil {
		    	return e
		    }
		    stat, e := file.Stat()
		    if e != nil {
		    	return e
		    }
		    size := stat.Size()
		    if size < int64(start) {
		    	// reset as file got truncated
		    	start = 0
		    }
		    if size != int64(start) {
		    	// Write diff to stdout
		    	buf := make([]byte, size-int64(start))
		    	if _, e := file.ReadAt(buf, int64(start)); e != nil {
		    		return e
		    	}
		    	msg := string(buf)
		    	msg = strings.Trim(msg, "\r")
		    	msg = strings.Trim(msg, "\n")
		    	fmt.Println(msg)

		    	// Remember pos
		    	bucket.Put([]byte(fileName), []byte(strconv.FormatInt(size, 10)))
		    }
		    return nil
		})
	}
}