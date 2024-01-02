package logic

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/augustogunsch/gobinet/internal/args"
	"github.com/augustogunsch/gobinet/internal/context"
)

type ProcessingFile struct {
	Input      string
	Output     string
	basePath   string
	content    string
	usesBibtex bool
	hasToc     bool
}

func (file *ProcessingFile) readInput() error {
	content, err := os.ReadFile(file.Input)

	if err != nil {
		return err
	}

	file.content = string(content)
	file.usesBibtex, _ = regexp.Match(`\\usepackage(\[.*\])?\{biblatex\}`, content)
	file.hasToc, _ = regexp.Match(`\\tableofcontents`, content)

	return nil
}

func (file *ProcessingFile) expandMacros() {
	prettyPath := strings.ReplaceAll(file.basePath, "_", " ")
	breadcrumbs := strings.ReplaceAll(prettyPath, "/", ` \textgreater\hspace{1pt} `)
	file.content = strings.ReplaceAll(file.content, `\breadcrumbs`, breadcrumbs)
	file.content = strings.ReplaceAll(file.content, `\slashcrumbs`, prettyPath)
	file.content = strings.ReplaceAll(file.content, `\outdir`, path.Dir(file.Output))
}

func (file *ProcessingFile) runBiber() (output []byte, err error) {
	cmd := exec.Command(
		"biber",
		strings.TrimSuffix(file.Output, path.Ext(file.Output)),
	)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("unable to run biber: %w", err)
	}
	return
}

func (file *ProcessingFile) runXelatex(include args.IncludeDirs) (output []byte, err error) {
	dir := path.Dir(file.Output)
	if err = os.MkdirAll(dir, os.FileMode(0755)); err != nil {
		return output, fmt.Errorf("unable to create directories `%s`: %w", dir, err)
	}

	cmd := exec.Command(
		"xelatex",
		"-jobname", path.Base(file.basePath),
		"-output-directory", path.Dir(file.Output),
		"-shell-escape",
		"-halt-on-error",
	)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TEXINPUTS="+include.String())

	stdin, _ := cmd.StdinPipe()
	stderr, _ := cmd.StderrPipe()
	stdout, _ := cmd.StdoutPipe()

	if err = cmd.Start(); err != nil {
		return output, fmt.Errorf("unable to start xelatex: %w", err)
	}

	stdin.Write([]byte(file.content))
	stdin.Close()

	output, _ = io.ReadAll(io.MultiReader(stdout, stderr))

	err = cmd.Wait()

	return
}

func (file *ProcessingFile) Generate(ctx context.Context) (success bool) {
	ctx.L.Printf("processing `%s`\n", file.Input)

	if err := file.readInput(); err != nil {
		ctx.L.Printf("failed to read file `%s`: `%s`\n", file.Input, err)
		if ctx.Args.Notify {
			ctx.N.Notify(ctx.L, "Failed to read file.")
		}
		return
	}

	file.expandMacros()

	if output, err := file.runXelatex(ctx.Args.Include); err != nil {
		ctx.L.Printf("failed to process file `%s`: `%s`\n%s", file.Input, err, output)
		if ctx.Args.Notify {
			ctx.N.Notify(ctx.L, "Failed to process file.")
		}
		return
	}

	ctx.L.Printf("processed `%s`\n", file.Input)

	if file.usesBibtex {
		ctx.L.Printf("running biber for `%s`\n", file.Output)
		if output, err := file.runBiber(); err != nil {
			ctx.L.Printf("failed to run biber for file `%s`\n%s", file.Output, output)
			if ctx.Args.Notify {
				ctx.N.Notify(ctx.L, "Failed to run Biber.")
			}
			return
		}
		ctx.L.Printf("ran bibtex for `%s`\n", file.Output)
	}

	if file.hasToc || file.usesBibtex {
		ctx.L.Printf("reprocessing `%s`\n", file.Input)

		if output, err := file.runXelatex(ctx.Args.Include); err != nil {
			ctx.L.Printf("failed to process file `%s`: `%s`\n%s", file.Input, err, output)
			if ctx.Args.Notify {
				ctx.N.Notify(ctx.L, "Failed to process file.")
			}
			return
		}

		ctx.L.Printf("reprocessed `%s`\n", file.Input)
	}

	if ctx.Args.Reload {
		cmd := exec.Command("pkill", "-HUP", "mupdf")
		if output, err := cmd.CombinedOutput(); err != nil {
			ctx.L.Printf("error reloading mupdf\n%s", output)
			if ctx.Args.Notify {
				ctx.N.Notify(ctx.L, "Failed to reload mupdf.")
			}
			return
		}
	}

	success = true
	return
}

func NewProcessingFile(args *args.ArgSet, inputFile string) ProcessingFile {
	basePath := inputFile[len(args.Input)+1:]
	basePath = strings.TrimSuffix(basePath, filepath.Ext(basePath))

	return ProcessingFile{
		Input:    inputFile,
		Output:   filepath.Join(args.Output, basePath+".pdf"),
		basePath: basePath,
	}
}

// Files starting with `_` are  ignored
func IsSourceFile(p string) bool {
	return path.Ext(p) == ".tex" && []rune(path.Base(p))[0] != '_'
}
