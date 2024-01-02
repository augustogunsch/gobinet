package cmds

import (
	"io/fs"
	"path/filepath"
	"sync"
	"time"

	"github.com/augustogunsch/gobinet/internal/context"
	"github.com/augustogunsch/gobinet/internal/logic"
	"github.com/fsnotify/fsnotify"
)

func handleUpdate(ctx context.Context, inputName string) {
	ctx.L.Printf("source `%s` updated", inputName)
	f := logic.NewProcessingFile(ctx.Args, inputName)
	f.Generate(ctx)
}

func Watch(ctx context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		ctx.L.Fatal(err)
	}
	defer watcher.Close()

	err = filepath.WalkDir(ctx.Args.Input, func(path string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		ctx.L.Fatal(err)
	}

	dedupLoop(ctx, watcher)
}

func dedupLoop(ctx context.Context, watcher *fsnotify.Watcher) {
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

			if !logic.IsSourceFile(e.Name) {
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

					handleUpdate(ctx, e.Name)
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

			ctx.L.Println("watcher error:", err)
		}
	}
}
