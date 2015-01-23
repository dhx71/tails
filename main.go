// tails project main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Output struct {
	filename string
	text     string
}

var (
	outputFileName = flag.Bool("src", false, "Output source filename")
	filesTailed    = make(map[string]bool)
	output         = make(chan Output)
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
	defer unsetAsTailed(filename)
	oriInfo, err := os.Lstat(filename)
	if err != nil {
		return
	}
	buffer := make([]byte, 65535)
	pos := oriInfo.Size()
	if !fromEnd {
		pos = 0
	}

	for {
		info, err := os.Lstat(filename)
		if err != nil {
			return
		}
		if !os.SameFile(info, oriInfo) || pos > info.Size() {
			pos = 0
			oriInfo = info
		}
		if info.Size() > pos {
			f, err := os.Open(filename)
			if err != nil {
				return
			}
			f.Seek(pos, os.SEEK_SET)

			for pos < info.Size() {
				n, err := f.Read(buffer)
				if err != nil || n == 0 {
					break
				}
				pos += int64(n)
				output <- Output{filename, string(buffer[0:n])}

			}
			f.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func doOutput() {
	lastFilename := ""
	for {
		t := <-output
		if t.filename != lastFilename {
			lastFilename = t.filename
			fmt.Printf("\n\n**** tail of file %s ****\n\n", lastFilename)
		}
		fmt.Print(t.text)
	}
}

func main() {
	flag.Parse()
	firstPass := true
	go doOutput()
	for {
		scanFiles(firstPass)
		firstPass = false
		time.Sleep(100 * time.Millisecond)
	}
}
