package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var watchingFolders string
var timeBetweenConverting string
var keepOriginals string
var keepLivePhoto string
var username string

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

func ConvertHeic(path string, srcFileName string, d fs.DirEntry, err error) error {
	if d.IsDir() {
		return nil
	}
	newFileName := fmt.Sprintf("%s.jpg", srcFileName)

	cmd := exec.Command("convert")
	cmd.Dir = path
	cmd.Args = append(cmd.Args, d.Name(), newFileName)
	stdErr, cerr := cmd.StderrPipe()
	go io.Copy(os.Stderr, stdErr)
	stdOut, cerr := cmd.StdoutPipe()
	go io.Copy(os.Stdout, stdOut)
	cerr = cmd.Run()
	if cerr != nil {
		log.Printf("ERROR: %v\n", cerr)
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
	username = os.Getenv("USERNAME")

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
		keepOriginals = "true"
		log.Printf("KEEP_ORIGNAL not specified. setting default to true")
	}
	if keepLivePhoto == "" {
		keepLivePhoto = "true"
		log.Printf("KEEP_LIVE_PHOTO not specified. setting default to true")
	}
	if username == "" {
		log.Printf("USERNAME not specified. setting default to 'no change'")
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

			if username != "" {
				UpdateFileOwner(folder, name)
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

func UpdateFileOwner(path string, srcFileName string) {
	// Get the userid and groupid from the username
	var userID, groupID int
	systemUser, err := user.Lookup(username)
	if err != nil {
		log.Printf("Error: %s \n", err)
		userID = -1
		groupID = -1
	} else {
		userID, _ = strconv.Atoi(systemUser.Uid)
		groupID, err = strconv.Atoi(systemUser.Gid)
		if err != nil {
			log.Printf("Error: %s \n", err)
			groupID = -1
		}
	}
	newFile := filepath.Join(path, srcFileName + ".jpg")
	if err := os.Chown(newFile, userID, groupID); err != nil {
		log.Printf("Error: %s \n", err)
		return
	}
	log.Printf("Updated %s: UID=%v, GUID=%v \n", newFile, userID, groupID)
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
