# Gobinet
Gobinet is the simplest personal knowledge base. It recursively generates PDF
files to a destination directory from an input directory of LaTeX source files.
It will generate twice when the source uses BibTeX (in which case it also
generates bibliography) or has a table of contents. It ignores up-to-date
files and files starting with and underscore (`_`).

## Instalation
```
go install github.com/augustogunsch/gobinet@v1.0.3
```

## Requirements
A standard LaTeX installation with `xelatex`.

## Macros
These get expanded when processing the source files.
 - `\outdir`: File's output directory.
 - `\breadcrumbs`: Pretty ">"-separated path to file.
 - `\slashcrumbs`: Slash separated path to file.

## Usage
For a concrete example, please look under `example/`.
```
usage: gobinet [--help] [--include DIR] [--reload] [--notify] <build|watch> INPUT OUTPUT
  -help
        Show this help message and exit.
  -include value
        Include this directory. May be passed multiple times.
  -notify
        Send a desktop notification when compilation fails.
  -reload
        Reload MuPDF by sending a HUP signal when files are updated.
```
