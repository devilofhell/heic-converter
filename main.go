package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var watchingFolders string
var timeBetweenConverting string
var keepOriginals string
var keepLivePhoto string

func main() {
	ReadEnv()

	Run(watchingFolders)
}

func Run(folder string) {
	duration, err := time.ParseDuration(timeBetweenConverting)
	if err != nil {
		log.Fatalf("parsing failed: %v, ", err)
	}
	ticker := time.NewTicker(duration)
	isConverting := false
	for {
		select {
		case <-ticker.C:
			if !isConverting {
				isConverting = true
				log.Printf("Start Converting")
				err := WalkDeleteHeic(folder)
			if err != nil {
				log.Fatalf("%v", err)
				return
			}
				isConverting = false
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

func ReadEnv() {
	watchingFolders = os.Getenv("WATCH")
	timeBetweenConverting = os.Getenv("TIME_BETWEEN")
	keepOriginals = os.Getenv("KEEP_ORIGINAL")
	keepLivePhoto = os.Getenv("KEEP_LIVE_PHOTO")

	if watchingFolders == "" {
		log.Fatalf("no folders to watch specified. set the WATCH environment variable. quit programm")
	}
	log.Printf("found WATCH folder: %s", watchingFolders)

	if timeBetweenConverting == "" {
		timeBetweenConverting = "1h"
		log.Printf("no time specified. start converting every 1 hour\n")
	}
	if keepOriginals == "" {
		keepOriginals = "false"
		log.Printf("KEEP_ORIGNAL not specified. setting default to false")
	}
	if keepLivePhoto == "" {
		keepLivePhoto = "false"
		log.Printf("KEEP_LIVE_PHOTO not specified. setting default to false")
	}
}

func WalkDeleteHeic(folder string) error {
	moduleFolder, _ := os.ReadDir(folder)

	for _, item := range moduleFolder {
		if item.IsDir() {
			err := WalkDeleteHeic(filepath.Join(folder, item.Name()))
			if err != nil {
				return err
			}
			continue
		}
		name, suffix := splitFile(item.Name()) 
		suffix = strings.ToLower(suffix)
		if suffix == "heic" {
			if IsAlreadyConverted(moduleFolder, name) {
				continue
			}

			err := Convert(folder, item, nil)
			if err != nil {
				return err
			}

			if keepLivePhoto == "false" {
				DeleteLivePhoto(moduleFolder, folder, name)
			}

			if keepOriginals == "false" {
				os.Remove(filepath.Join(folder, item.Name()))
			}
		}
}
	return nil
}

func IsAlreadyConverted(moduleFolder []fs.DirEntry, fileName string) bool {
	jpgFile := fileName + ".jpg"

	for _, item := range moduleFolder {
		if item.Name() == jpgFile {
			return true
		}
	}
	return false
}

func DeleteLivePhoto(moduleFolder []fs.DirEntry, folderPath, fileName string) error {
	livePhotoFileMOV := fileName + ".MOV"
	livePhotoFileMov := fileName + ".mov"

	for _, item := range moduleFolder {
		if item.Name() == livePhotoFileMOV {
			file := filepath.Join(folderPath, livePhotoFileMOV)
			return os.Remove(file)
		}
		if item.Name() == livePhotoFileMov {
			file := filepath.Join(folderPath, livePhotoFileMov)
			return os.Remove(file)
		}
	}
	return nil
}

func splitFile(fileName string) (string, string) {
	splitted := strings.Split(fileName, ".")
	switch len(splitted) {
	case 0: 
		return "", ""
	case 1:
		return splitted[0], "" 
	case 2:
		return splitted[0], splitted[1]
	default:
		length := len(splitted)
		return path.Join(splitted[0:length-2]...), splitted[length-1]
	}
}