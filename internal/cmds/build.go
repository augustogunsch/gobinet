package cmds

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/augustogunsch/gobinet/internal/args"
	"github.com/augustogunsch/gobinet/internal/context"
	"github.com/augustogunsch/gobinet/internal/logic"
)

func getModTime(file string) (time.Time, error) {
	var modTime time.Time
	stat, err := os.Stat(file)
	if err != nil {
		if !os.IsNotExist(err) {
			return modTime, err
		}
	} else {
		modTime = stat.ModTime()
	}
	return modTime, nil
}

func getOutdatedFiles(args *args.ArgSet) ([]logic.ProcessingFile, error) {
	var sourceFiles []string

	err := filepath.WalkDir(args.Input, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && logic.IsSourceFile(p) {
			sourceFiles = append(sourceFiles, p)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	files := make([]logic.ProcessingFile, 0, len(sourceFiles))

	for _, file := range sourceFiles {
		f := logic.NewProcessingFile(args, file)

		inputTime, err := getModTime(f.Input)
		if err != nil {
			return files, err
		}
		outputTime, err := getModTime(f.Output)
		if err != nil {
			return files, err
		}

		if inputTime.After(outputTime) {
			files = append(files, f)
		}
	}

	return files, nil
}

func Build(ctx context.Context) {
	var (
		wg    sync.WaitGroup
		files []logic.ProcessingFile
	)

	files, err := getOutdatedFiles(ctx.Args)
	if err != nil {
		ctx.L.Fatal(err)
	}

	wg.Add(len(files))

	ctx.L.Println("starting build")
	for _, file := range files {
		f := file
		go func() {
			defer wg.Done()
			f.Generate(ctx)
		}()
	}

	wg.Wait()
	ctx.L.Println("finished build")
}
