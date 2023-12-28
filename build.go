package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type processingFile struct {
	input      string
	output     string
	basePath   string
	content    string
	usesBibtex bool
	hasToc     bool
}

func (file *processingFile) readInput() error {
	content, err := os.ReadFile(file.input)

	if err != nil {
		return err
	}

	file.content = string(content)
	file.usesBibtex, _ = regexp.Match(`\\usepackage(\[.*\])?\{biblatex\}`, content)
	file.hasToc, _ = regexp.Match(`\\tableofcontents`, content)

	return nil
}

func (file *processingFile) expandMacros() {
	prettyPath := strings.ReplaceAll(file.basePath, "_", " ")
	breadcrumbs := strings.ReplaceAll(prettyPath, "/", ` \textgreater\hspace{1pt} `)
	file.content = strings.ReplaceAll(file.content, `\breadcrumbs`, breadcrumbs)
	file.content = strings.ReplaceAll(file.content, `\slashcrumbs`, prettyPath)
	file.content = strings.ReplaceAll(file.content, `\outdir`, path.Dir(file.output))
}

func (file *processingFile) runBiber() (string, error) {
	cmd := exec.Command(
		"biber",
		strings.TrimSuffix(file.output, path.Ext(file.output)),
	)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (file *processingFile) runXelatex(include IncludeDirs) (string, error) {
	dir := path.Dir(file.output)
	err := os.MkdirAll(dir, os.FileMode(0755))

	if err != nil {
		return "", fmt.Errorf("unable to create directories `%s`: %w", dir, err)
	}

	xelatex, err := exec.LookPath("xelatex")

	if err != nil {
		return "", errors.New("LaTeX doesn't seem to be installed")
	}

	cmd := exec.Command(
		xelatex,
		"-jobname", path.Base(file.basePath),
		"-output-directory", path.Dir(file.output),
		"-shell-escape",
		"-halt-on-error",
	)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TEXINPUTS="+include.String())

	stdin, _ := cmd.StdinPipe()
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("unable to start xelatex: %w", err)
	}

	stdin.Write([]byte(file.content))
	stdin.Close()

	output, _ := io.ReadAll(stdout)
	errors, _ := io.ReadAll(stderr)
	stdout.Close()
	stderr.Close()
	combinedOutput := string(output) + string(errors)

	if err := cmd.Wait(); err != nil {
		return combinedOutput, err
	}

	return combinedOutput, nil
}

func notify(msg string) {
	cmd := exec.Command("notify-send", "Gobinet error", msg)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("error sending notification:\n%s", output)
		return
	}
}

func (file *processingFile) generate(args ArgSet) {
	log.Printf("processing `%s`\n", file.input)

	if err := file.readInput(); err != nil {
		log.Printf("failed to read file `%s`: `%s`\n", file.input, err)
		if args.Notify {
			notify("Failed to read file.")
		}
		return
	}

	file.expandMacros()

	if output, err := file.runXelatex(args.Include); err != nil {
		log.Printf("failed to process file `%s`: `%s`\n%s", file.input, err, output)
		if args.Notify {
			notify("Failed to process file.")
		}
		return
	}

	log.Printf("processed `%s`\n", file.input)

	if file.usesBibtex {
		log.Printf("running bibtex for `%s`\n", file.output)
		if output, err := file.runBiber(); err != nil {
			log.Printf("failed to run biber for file `%s`\n%s", file.output, output)
			if args.Notify {
				notify("Failed to run Biber.")
			}
			return
		}
		log.Printf("ran bibtex for `%s`\n", file.output)
	}

	if file.hasToc || file.usesBibtex {
		log.Printf("reprocessing `%s`\n", file.input)

		if output, err := file.runXelatex(args.Include); err != nil {
			log.Printf("failed to process file `%s`: `%s`\n%s", file.input, err, output)
			if args.Notify {
				notify("Failed to process file.")
			}
			return
		}

		log.Printf("reprocessed `%s`\n", file.input)
	}

	if args.Reload {
		cmd := exec.Command("pkill", "-HUP", "mupdf")
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Printf("error reloading mupdf\n%s", output)
			if args.Notify {
				notify("Failed to reload mupdf.")
			}
			return
		}
	}
}

func getModTime(file string) time.Time {
	var modTime time.Time
	stat, err := os.Stat(file)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	} else {
		modTime = stat.ModTime()
	}
	return modTime
}

func newProcessingFile(inputFile, inputDir, outputDir string) processingFile {
	basePath := inputFile[len(inputDir)+1:]
	basePath = strings.TrimSuffix(basePath, filepath.Ext(basePath))

	return processingFile{
		input:    inputFile,
		output:   filepath.Join(outputDir, basePath+".pdf"),
		basePath: basePath,
	}
}

func getProcessingFiles(inputDir, outputDir string) []processingFile {
	var sourceFiles []string

	filepath.WalkDir(inputDir, func(p string, d fs.DirEntry, err error) error {
		if !d.IsDir() && path.Ext(p) == ".tex" && []rune(path.Base(p))[0] != '_' {
			sourceFiles = append(sourceFiles, p)
		}
		return nil
	})

	files := make([]processingFile, 0, len(sourceFiles))

	for _, file := range sourceFiles {
		f := newProcessingFile(file, inputDir, outputDir)

		inputTime := getModTime(f.input)
		outputTime := getModTime(f.output)
		upToDate := inputTime.Before(outputTime)

		if !upToDate {
			files = append(files, f)
		}
	}

	return files
}

func build(args ArgSet) {
	var (
		wg    sync.WaitGroup
		files = getProcessingFiles(args.Input, args.Output)
	)
	wg.Add(len(files))

	log.Println("starting build")
	for _, file := range files {
		f := file
		go func() {
			defer wg.Done()
			f.generate(args)
		}()
	}

	wg.Wait()
	log.Println("finished build")
}
