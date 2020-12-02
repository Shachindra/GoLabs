package main

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/hpcloud/tail"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	fileName = "C:/Users/myuser/Documents/testing/text.txt"
	text1    = "Lorem ipsum dolor."
	text2    = "Lorem dolor sit amet."
)

// main
func main() {

	t, err := tail.TailFile(fileName, tail.Config{Follow: true})
	for line := range t.Lines {
		fmt.Println(line.Text)
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(text1, text2, false)
	fmt.Println(dmp.DiffPrettyText(diffs))

	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR", err)
	}
	defer watcher.Close()

	//
	done := make(chan bool)

	//
	go func() {
		for {
			select {
			// watch for events
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				fmt.Printf("EVENT! %#v\n", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}

				// watch for errors
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("ERROR", err)
			}
		}
	}()

	// out of the box fsnotify can watch a single file, or a single directory
	if err := watcher.Add(fileName); err != nil {
		fmt.Println("ERROR", err)
	}

	<-done
}
