package main

import (
	"fmt"
	"deltareport/config"
	"deltareport/model"
	"deltareport/diff"
	"flag"
	//"strings"
)

func main() {
	configPath := ""
	flag.BoolVar(&config.Verbose, "v", false, "Verbose-mode (log more)")
	flag.StringVar(&configPath, "c", "./config.json", "Path to config.json")
	flag.Parse()

	if e := config.Init(configPath); e != nil {
		panic(e)
	}
	// TODO: Handle toggling recurse true/false
	for path, meta := range config.C.Files {
		pos, e := model.Pos(path)
		if e != nil {
			panic(e)
		}

		lookup := make(map[string]diff.Res)
		if meta.Recurse {
			lookup, e = diff.Recurse(path, pos)
		} else {
			lookup[path], e = diff.File(path, pos[path])
		}
		if e != nil {
			panic(e)
		}

		// show diff
		fmt.Printf("%+v\n", lookup)

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