package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mpdroog/deltareport/config"
	"github.com/mpdroog/deltareport/diff"
	"github.com/mpdroog/deltareport/model"
	"github.com/mpdroog/deltareport/queue"
)

func main() {
	configPath := ""
	flag.BoolVar(&config.Verbose, "v", false, "Verbose-mode (log more)")
	flag.BoolVar(&config.Debug, "d", false, "Debug-mode (log diff msgs)")
	flag.StringVar(&configPath, "c", "./config.toml", "Path to config.toml")
	flag.Parse()

	if e := config.Init(configPath); e != nil {
		panic(e)
	}
	defer func() {
		if e := config.Close(); e != nil {
			panic(e)
		}
	}()

	if config.Verbose {
		fmt.Printf("Config=%+v\n", config.C)
	}

	// TODO: Handle toggling recurse true/false
	for _, meta := range config.C.Files {
		path := meta.Path
		pos, e := model.Pos(meta.To, path)
		if e != nil {
			panic(e)
		}
		if config.Verbose {
			fmt.Printf("File(%s) pos=%+v\n", path, pos)
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
		if config.Debug {
			fmt.Printf("diff=%+v\n", lookup)
		}

		sumbytes := 0
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
					fmt.Printf("script(%s)\n", name)
				}

				sumbytes += len(diff)
				v.Diff = diff
				lookup[k] = v
			}
		}
		if config.Verbose {
			fmt.Printf("script.filtered bytes=%d\n", sumbytes)
		}

		// No scripts found so difference in sumbytes
		if sumbytes == 0 && len(lookup) > 0 {
			sumbytes = len(lookup)
		}

		if sumbytes > 0 {
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
