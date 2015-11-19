package diff

import (
	"os"
	"path/filepath"
	"strings"
	"fmt"
)

type Res struct {
	Pos int64
	Diff string
}

func File(fileName string, start int64) (Res, error) {
	var r Res
	file, e := os.Open(fileName)
	if e != nil {
		return r, e
	}
	stat, e := file.Stat()
	if e != nil {
		return r, e
	}
	if stat.IsDir() {
		return r, fmt.Errorf("Given path is dir: " + fileName)
	}
	size := stat.Size()

	if size < start {
		// reset as file got truncated
		start = 0
	}
	if size != start {
		// change!
		buf := make([]byte, size-start)
    	if _, e := file.ReadAt(buf, start); e != nil {
    		return r, e
    	}
    	msg := string(buf)
    	msg = strings.Trim(msg, "\r")
    	msg = strings.Trim(msg, "\n")

		r.Diff = msg
	}

	r.Pos = size
	return r, nil
}

func Recurse(basedir string, posLookup map[string]int64, exts []string) (map[string]Res, error) {
	out := make(map[string]Res)
	e := filepath.Walk(basedir, func(path string, f os.FileInfo, err error) error {
		if path == basedir {
			// ignore root
			return nil
		}
		var e error
		ok := false
		for _, ext := range exts {
			if strings.HasSuffix(path, ext) {
				ok = true
				break
			}
		}
		if !ok {
			// Skip file, not matching pattern
			return nil
		}

		out[path], e = File(path, posLookup[path])
		return e
	})
	return out, e
}