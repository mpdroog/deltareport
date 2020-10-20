package main

import (
	"flag"
	"fmt"
	"github.com/mpdroog/deltareport/config"
	"github.com/mpdroog/deltareport/diff"
	"github.com/mpdroog/deltareport/model"
	"github.com/mpdroog/deltareport/queue"
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
			fmt.Printf("diff=%+v\n", lookup)
		}
		// allow filtering by JS
		// TODO: interrupt handler to limit execution to N-seconds?
		vm := config.ScriptEngine
		for name, script := range config.Scripts {
			for k, v := range lookup {
				vm.Set("queue", meta.To)
				vm.Set("body", v.Diff)
				value, e := vm.Run(script)
				if e != nil {
					panic(e)
				}
				diff, e := value.ToString()
				if e != nil {
					panic(e)
				}
				if config.Verbose {
					fmt.Printf("script(%s) out=%s\n", name, diff)
				}

				if len(diff) == 0 {
					// strip out
					delete(lookup, k)
					continue
				}

				v.Diff = diff
				lookup[k] = v
			}
		}
		if config.Verbose {
			fmt.Printf("script.filtered diff=%+v\n", lookup)
		}

		if len(lookup) > 0 {
			// report diff
			e = queue.Mail(path, meta.To, lookup)
			if e == queue.ErrNotFound {
				e = queue.Newline(path, lookup, meta)
			}
			if e != nil {
				panic(e)
			}
		}

		// save new file positions
		newPos := make(map[string]int64)
		for file, meta := range lookup {
			newPos[file] = meta.Pos
		}
		if e := model.SavePos(meta.To, path, newPos); e != nil {
			panic(e)
		}
	}
}
