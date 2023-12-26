package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type processingFile struct {
	input   string
	pdfOut  string
	htmlOut string
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

func getOutputBase(inputFile, inputDir, outputDir string) string {
	inputDirsCount := strings.Count(inputDir, string(filepath.Separator))
	inputSegments := strings.Split(inputFile, string(filepath.Separator))
	inputFile = filepath.Join(inputSegments[inputDirsCount+1:]...)
	return outputDir + "/" + strings.TrimSuffix(inputFile, filepath.Ext(inputFile))
}

func newProcessingFile(inputFile, inputDir, outputDir string) (file processingFile, upToDate bool) {
	file.input = inputFile

	outputBase := getOutputBase(inputFile, inputDir, outputDir)
	fmt.Println(outputBase)
	file.pdfOut = outputBase + ".pdf"
	file.htmlOut = outputBase + ".html"

	inputTime := getModTime(inputFile)
	pdfTime := getModTime(file.pdfOut)
	htmlTime := getModTime(file.htmlOut)
	upToDate = inputTime.Before(pdfTime) && inputTime.Before(htmlTime)

	return
}

func getProcessingFiles(inputDir, outputDir string) []processingFile {
	sourceFiles, _ := filepath.Glob(inputDir + "/**/*.tex")

	files := make([]processingFile, 0, len(sourceFiles))

	for _, file := range sourceFiles {
		f, upToDate := newProcessingFile(file, inputDir, outputDir)
		if !upToDate {
			files = append(files, f)
		}
	}

	return files
}

func build(args ArgSet) {
	files := getProcessingFiles(args.Input, args.Output)

	for _, file := range files {
		fmt.Println(file)
	}
}
