package main

import (
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
				log.Printf("Finished Converting")
				isConverting = false
			}
		}
	}
}

// func ConvertFolder(folder string) error {
// 	//files, _ := ioutil.ReadDir(folder)

// 	filepath.WalkDir(folder, Convert)

// 	return nil
// }

func ConvertHeic(path string, srcFileName string, d fs.DirEntry, err error) error {
	if d.IsDir() {
		return nil
	}
	newFileName := fmt.Sprintf("%s.jpg", srcFileName)

	// fmt.Printf("CurrentFile: %s\n", filepath.Join(path, d.Name()))
	// fmt.Printf("NewFile    : %s\n\n", filepath.Join(path, newFileName))

	// fmt.Printf("splitPath  : %v\n", currentFolder)
	// fmt.Printf("cmdDir     : %s\n", path)

	cmd := exec.Command("convert")
	cmd.Dir = path
	cmd.Args = append(cmd.Args, d.Name(), newFileName)
	stdErr, cerr := cmd.StderrPipe()
	go io.Copy(os.Stderr, stdErr)
	stdOut, cerr := cmd.StdoutPipe()
	go io.Copy(os.Stdout, stdOut)
	cerr = cmd.Run()
	if cerr != nil {
		fmt.Printf("ERROR: %v\n", cerr)
		return cerr
	}
	log.Printf("successful converting of [%s  ->  %s]\n", d.Name(), newFileName)
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
	log.Printf("start converting every: %s", timeBetweenConverting)
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

			err := ConvertHeic(folder, name, item, nil)
			if err != nil {
				log.Printf("error while converting with current file (%s). dont delete original picture or live-photo: %s", item.Name(), err)
				continue
			}

			if keepLivePhoto == "false" {
				DeleteLivePhoto(moduleFolder, folder, name)
			}

			if keepOriginals == "false" {
				srcFile := filepath.Join(folder, item.Name())
				log.Printf("delete original file: %s", srcFile)
				os.Remove(srcFile)
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
			log.Printf("delete live-photo file: %s", file)
			return os.Remove(file)
		}
		if item.Name() == livePhotoFileMov {
			file := filepath.Join(folderPath, livePhotoFileMov)
			log.Printf("delete live-photo file: %s", file)
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
		return path.Join(splitted[0 : length-2]...), splitted[length-1]
	}
}