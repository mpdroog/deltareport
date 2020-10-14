package queue

import (
	"bytes"
	"encoding/json"
	"github.com/mpdroog/beanstalkd"
	"github.com/mpdroog/deltareport/config"
	"github.com/mpdroog/deltareport/diff"
	"time"
)

type Email struct {
	From      string // Key that MUST match From in config
	To        []string
	Subject   string
	Html      string
	Text      string
	HtmlEmbed map[string]string // file.png => base64(bytes)
}

func Mail(path string, key string, diff map[string]diff.Res) error {
	q, ok := config.C.Queues.Mail[key]
	if !ok {
		return ErrNotFound
	}

	txt := ""
	counter := 0
	for file, meta := range diff {
		if len(meta.Diff) == 0 {
			continue
		}
		counter += len(meta.Diff)
		txt += file + "\n===============================\n\n" + meta.Diff + "\n\n"
	}
	if counter == 0 {
		// no diff
		return nil
	}

	m := Email{
		From:    q.From,
		To:      q.To,
		Subject: q.Subject + config.Hostname + " " + path,
		Text:    txt,
	}

	w := new(bytes.Buffer)
	enc := json.NewEncoder(w)
	if e := enc.Encode(m); e != nil {
		return e
	}

	queue, err := beanstalkd.Dial(q.Beanstalkd)
	if err != nil {
		return err
	}
	defer queue.Quit()
	queue.Use("email")

	_, e := queue.Put(
		1, 0*time.Second, 5*time.Second,
		w.Bytes(),
	)
	return e
}
