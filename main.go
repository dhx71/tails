// tails project main.go
package main

import (
	"flag"
	"fmt"
	"github.com/ActiveState/tail"
	"path/filepath"
	"time"
)

var (
	outputFileName = flag.Bool("src", false, "Output source filename")
	filesTailed    = make(map[string]bool)
)

func isTailed(filename string) (ok bool) {
	_, ok = filesTailed[filename]
	return
}

func setAsTailed(filename string) {
	filesTailed[filename] = true
}

func unsetAsTailed(filename string) {
	delete(filesTailed, filename)
}

func tailFiles(pattern string, firstPass bool) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Errorf("invalid pattern %s\nError %v\n", pattern, err)
		return
	}
	for _, file := range files {
		if !isTailed(file) {
			setAsTailed(file)
			go tailFile(file, firstPass)
		}
	}
}

func scanFiles(firstPass bool) {
	for _, pattern := range flag.Args() {
		tailFiles(pattern, firstPass)
	}
}

func tailFile(filename string, fromEnd bool) {
	cfg := tail.Config{
		Follow: true,
		ReOpen: true,
		Poll:   true,
		Logger: tail.DiscardingLogger,
	}
	if fromEnd {
		cfg.Location = &tail.SeekInfo{0, 2}
	}
	t, err := tail.TailFile(filename, cfg)
	if err != nil {
		fmt.Errorf("Tail file %s error: %v\n", filename, err)
		unsetAsTailed(filename)
		return
	}
	for line := range t.Lines {
		fmt.Println(line.Text)
	}

	unsetAsTailed(filename)
}

func main() {
	flag.Parse()
	firstPass := true
	for {
		scanFiles(firstPass)
		firstPass = false
		time.Sleep(100 * time.Millisecond)
	}
}
