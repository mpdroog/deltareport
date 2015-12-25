package main

import (
	"deltareport/config"
	"deltareport/diff"
	"deltareport/model"
	"deltareport/queue"
	"flag"
	"fmt"
)

func main() {
	configPath := ""
	flag.BoolVar(&config.Verbose, "v", false, "Verbose-mode (log more)")
	flag.StringVar(&configPath, "c", "./config.json", "Path to config.json")
	flag.Parse()

	if e := config.Init(configPath); e != nil {
		panic(e)
	}
	defer config.Close()

	// TODO: Handle toggling recurse true/false
	for path, meta := range config.C.Files {
		pos, e := model.Pos(path)
		if e != nil {
			panic(e)
		}

		lookup := make(map[string]diff.Res)
		if meta.Recurse {
			lookup, e = diff.Recurse(path, pos, meta.IncludeExt)
		} else {
			lookup[path], e = diff.File(path, pos[path])
		}
		if e != nil {
			panic(e)
		}

		// show diff
		if config.Verbose {
			fmt.Printf("%+v\n", lookup)
		}
		// report diff
		e = queue.Mail(path, meta.To, lookup)
		if e == queue.ErrNotFound {
			e = queue.Newline(path, meta.To, lookup, meta.Tags)
		}
		if e != nil {
			panic(e)
		}

		// save new file positions
		newPos := make(map[string]int64)
		for file, meta := range lookup {
			newPos[file] = meta.Pos
		}
		if e := model.SavePos(path, newPos); e != nil {
			panic(e)
		}
	}
}
