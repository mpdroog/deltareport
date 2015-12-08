package model

import (
	"bytes"
	"deltareport/config"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
)

func val(bucket string, key string) ([]byte, error) {
	var val []byte
	e := config.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		val = b.Get([]byte(key))
		return nil
	})
	return val, e
}

func save(bucket string, key string, val []byte) error {
	e := config.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return b.Put([]byte(key), val)
	})
	return e
}

func Pos(path string) (map[string]int64, error) {
	out := make(map[string]int64)
	raw, e := val("filepos", path)
	if e != nil {
		return out, e
	}
	if len(raw) == 0 {
		// default to zero on nothing
		out[path] = 0
		return out, nil
	}

	r := bytes.NewReader(raw)
	e = json.NewDecoder(r).Decode(&out)
	return out, e
}

func SavePos(path string, vals map[string]int64) error {
	w := new(bytes.Buffer)
	enc := json.NewEncoder(w)
	e := enc.Encode(vals)
	if e != nil {
		return e
	}
	return save("filepos", path, w.Bytes())
}
