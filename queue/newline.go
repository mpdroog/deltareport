package queue

import (
	"bytes"
	"deltareport/config"
	"deltareport/diff"
	"encoding/json"
	"github.com/mpdroog/beanstalkd"
	"strings"
	"time"
)

type LineDiff struct {
	Hostname string
	Path     string
	Line     string
	Tags     []string
}

func Newline(path string, diff map[string]diff.Res, conf config.File) error {
	q, ok := config.C.Queues.Newline[conf.To]
	if !ok {
		return ErrNotFound
	}

	var lines []LineDiff
	for file, meta := range diff {
		if len(meta.Diff) == 0 {
			continue
		}
		if conf.Linediff {
			for _, line := range strings.Split(meta.Diff, "\n") {
				lines = append(lines, LineDiff{
					Hostname: config.Hostname,
					Path:     file,
					Line:     line,
					Tags:     conf.Tags,
				})
			}
		} else {
			lines = append(lines, LineDiff{
				Hostname: config.Hostname,
				Path:     file,
				Line:     meta.Diff,
				Tags:     conf.Tags,
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
