package main

import (
	"io/fs"
	"log"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

func processUpdate(inputName string, args ArgSet) {
	log.Printf("source `%s` updated", inputName)
	f := newProcessingFile(inputName, args.Input, args.Output)
	f.generate(args)
}

func watch(args ArgSet) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = filepath.WalkDir(args.Input, func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	dedupLoop(args, watcher)
}

func dedupLoop(args ArgSet, watcher *fsnotify.Watcher) {
	const waitFor = 50 * time.Millisecond
	var (
		mu     sync.Mutex
		timers = make(map[string]*time.Timer)
	)

	for {
		select {
		case e, ok := <-watcher.Events:
			if !ok {
				return
			}

			if !e.Has(fsnotify.Write) && !e.Has(fsnotify.Create) {
				continue
			}

			if path.Ext(e.Name) != ".tex" || []rune(path.Base(e.Name))[0] == '_' {
				continue
			}

			mu.Lock()
			t, ok := timers[e.Name]
			mu.Unlock()

			if !ok {
				t = time.AfterFunc(waitFor, func() {
					mu.Lock()
					delete(timers, e.Name)
					mu.Unlock()

					processUpdate(e.Name, args)
				})

				mu.Lock()
				timers[e.Name] = t
				mu.Unlock()
			} else {
				t.Reset(waitFor)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			log.Println("watcher error:", err)
		}
	}
}
