package queue

import (
	"bytes"
	"encoding/json"
	"github.com/mpdroog/beanstalkd"
	"github.com/mpdroog/deltareport/config"
	"github.com/mpdroog/deltareport/diff"
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
			txt := meta.Diff
			lines = append(lines, LineDiff{
				Hostname: config.Hostname,
				Path:     file,
				Line:     txt,
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
