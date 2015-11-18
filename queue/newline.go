package queue

import (
	"strings"
	"deltareport/config"
	"deltareport/diff"
	"github.com/mpdroog/beanstalkd"
	"bytes"
	"encoding/json"
	"time"
)

type LineDiff struct {
	Path string
	Line string
}

func Newline(path string, key string, diff map[string]diff.Res) error {
	q, ok := config.C.Queues.Newline[key]
	if !ok {
		return ErrNotFound
	}

	var lines []LineDiff
    for file, meta := range diff {
    	if len(meta.Diff) == 0 {
    		continue
    	}
    	for _, line := range strings.Split(meta.Diff, "\n") {
	    	lines = append(lines, LineDiff{
	    		Path: file,
	    		Line: line,
	    	})
	    }
    }
    if len(lines) == 0 {
    	// no diff
    	return nil
    }

    // queue lines
	queue, err := beanstalkd.Dial(q.Beanstalkd)
    if err != nil {
        return err
    }
    defer queue.Quit()
    queue.Use(q.Queue)

    w := new(bytes.Buffer)
    for _, line := range lines {
    	w.Reset()
		enc := json.NewEncoder(w)
		if e := enc.Encode(line); e != nil {
			return e
		}

	    _, e := queue.Put(
	    	1, 0*time.Second, 5*time.Second,
	    	w.Bytes(),
	    )
	    if e != nil {
	    	return e
	    }
	}
    return nil
}