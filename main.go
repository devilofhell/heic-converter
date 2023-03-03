package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// func main() {
// 	fmt.Printf("Hello world")
// 	args := []string{"convert", "IMG_0620.HEIC", "IMG_0620.jpg"}
// 	imagick.ConvertImageCommand(args)
// 	folderPath := "./"

// 	cmd := exec.Command("magick")
// 	cmd.Dir = folderPath
// }

func main() {

	folder := "./testdata/01"

	ConvertFolder(folder)
	//Run(folder)
}

func Run(folder string) {
	duration, _ := time.ParseDuration("5s")
	ticker := time.NewTicker(duration)
	for {
		select {
		case <-ticker.C:
			err := ConvertFolder(folder)
			if err != nil {
				log.Fatalf("%v", err)
				return
			}
		}
	}
}

func ConvertFolder(folder string) error {
	//files, _ := ioutil.ReadDir(folder)

	filepath.WalkDir(folder, Convert)

	return nil
}

func Convert(path string, d fs.DirEntry, err error) error {
	if d.IsDir() {
		return nil
	}
	splitted := strings.Split(path, "/")
	var buf bytes.Buffer

	splittedFileName := strings.Split(d.Name(), ".")
	newFileName := fmt.Sprintf("%s.jpg", splittedFileName[0])

	var currentFolder string

	for i := 0; i < len(splitted); i++ {
		if i == len(splitted)-1 {
			currentFolder = buf.String()
			buf.WriteString(newFileName)
			break
		}
		buf.WriteString(splitted[i])
		buf.WriteString("/")
	}
	fmt.Printf("CurrentFile: ./%s\n", path)
	fmt.Printf("NewFile    : %s\n", newFileName)

	fmt.Printf("Path       : %s\n", path)
	fmt.Printf("CurrentFold: %s\n", currentFolder)
	fmt.Printf("FileName   : %s\n", d.Name())

	baseDir, err := os.Getwd()

	fulDir := filepath.Join(baseDir, currentFolder)

	fmt.Printf("fulDir     : %s\n", fulDir)

	cmd := exec.Command("convert")
	cmd.Dir = fulDir
	cmd.Args = append(cmd.Args, d.Name(), newFileName)
	fmt.Printf("%v", cmd.Args)
	stdErr, cerr := cmd.StderrPipe()
	go io.Copy(os.Stderr, stdErr)
	stdOut, cerr := cmd.StdoutPipe()
	go io.Copy(os.Stdout, stdOut)
	cerr = cmd.Run()
	if cerr != nil {
		fmt.Printf("ERROR: %v\n", cerr)
		return nil
	}

	return nil
}
