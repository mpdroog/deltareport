package main

import (
	"deltareport/config"
	"deltareport/diff"
	"deltareport/model"
	"deltareport/queue"
	"flag"
	"fmt"
	"os"
)

func main() {
	configPath := ""
	flag.BoolVar(&config.Verbose, "v", false, "Verbose-mode (log more)")
	flag.StringVar(&configPath, "c", "./config.toml", "Path to config.toml")
	flag.Parse()

	if e := config.Init(configPath); e != nil {
		panic(e)
	}
	defer config.Close()

	if config.Verbose {
		fmt.Printf("%+v\n", config.C)
	}

	// TODO: Handle toggling recurse true/false
	for _, meta := range config.C.Files {
		path := meta.Path
		pos, e := model.Pos(meta.To, path)
		if e != nil {
			panic(e)
		}

		lookup := make(map[string]diff.Res)
		if meta.Recurse {
			lookup, e = diff.Recurse(path, pos, meta.IncludeExt, meta.Regexp)
		} else {
			lookup[path], e = diff.File(path, pos[path])
		}
		if e != nil {
			if os.IsNotExist(e) {
				fmt.Fprintf(os.Stderr, "WARN: %s\n", e.Error())
				continue
			}
			panic(e)
		}

		// show diff
		if config.Verbose {
			fmt.Printf("%+v\n", lookup)
		}
		// report diff
		if e = queue.Mail(path, meta.To, lookup); e != nil {
			panic(e)
		}

		// save new file positions
		newPos := make(map[string]int64)
		for file, meta := range lookup {
			newPos[file] = meta.Pos
		}
		if e := model.SavePos(meta.To, path, newPos); e != nil {
			panic(e)
		}
		if config.Verbose {
			fmt.Printf("Updated file positions\n")
		}
	}
}
