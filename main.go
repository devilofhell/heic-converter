package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"strconv"
	"time"
)

var watchingFolders string
var targetFolder string
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
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cerr := cmd.Run()
	if cerr != nil {
		log.Printf("ERROR: %v\n", cerr)
		return cerr
	}

	// Get the userid and groupid from the username
	var uid, gid int
	uname, err := user.Lookup(username)
	if err != nil {
		log.Printf("Error: %s \n", err)
		uid = -1
		gid = -1
	} else {
		uid, _ = strconv.Atoi(uname.Uid)
		gid, err = strconv.Atoi(uname.Gid)
		if err != nil {
			log.Printf("Error: %s \n", err)
			gid =  -1
		}
	}
	newFile := fmt.Sprintf("%s/%s.jpg", path, srcFileName)
	if err := os.Chown(newFile, uid,  gid); err != nil {
		log.Printf("Error: %s \n", err)
	} else {
		log.Printf("Updated %s: UID=%v, GUID=%v \n", newFile, uid, gid)
	}
	
	log.Printf("successful converting of [%s  ->  %s]\n", d.Name(), newFileName)
	return nil
}

func ReadEnv() {
	watchingFolders = os.Getenv("WATCH")
	targetFolder = os.Getenv("TARGET")
	timeBetweenConverting = os.Getenv("TIME_BETWEEN")
	keepOriginals = os.Getenv("KEEP_ORIGINAL")
	keepLivePhoto = os.Getenv("KEEP_LIVE_PHOTO")
	username = os.Getenv("USERNAME")

	if watchingFolders == "" {
		log.Fatalf("no folders to watch specified. set the WATCH environment variable. quit programm")
	}
	log.Printf("found WATCH folder: %s", watchingFolders)

	if targetFolder == "" {
		log.Printf("no target specified. converted files are stored in the same folder as the watching folder\n")
	}
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

type FileIDs struct {
	UserID  int
	GroupID int
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
			if IsAlreadyConverted(folder, name) {
				continue
			}

			fileIDs := findFileIDs(item)

			err := ConvertHeic(folder, name, item, nil)
			if err != nil {
				log.Printf("error while converting with current file (%s). dont delete original picture or live-photo: %s", item.Name(), err)
				continue
			}

			if targetFolder != "" {
				moveFile(folder, name, fileIDs)
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


func moveFile(sourceFolder string, filename string, fileIDs FileIDs) {
	filename = filename + ".jpg"

	sourcePath := filepath.Join(sourceFolder, filename)
	destinationPath := filepath.Join(targetFolder, filename)

	if sourceFolder != watchingFolders {
		addedFolders, _ := strings.CutPrefix(sourceFolder, watchingFolders)
		destinationDir := filepath.Join(targetFolder, addedFolders)
		destinationPath = filepath.Join(destinationDir, filename)

		err := MkDirAllAndChown(addedFolders, fileIDs)
		if err != nil {
			log.Println(err)
			return
		}
	}

	if _, err := copy(sourcePath, destinationPath, fileIDs); err != nil {
		log.Println(err)
		return
	}
	log.Printf("moved [%s  ->  %s]\n", sourcePath, destinationPath)

	err := os.Remove(sourcePath)
	if err != nil {
		log.Println(err)
	}
}

func MkDirAllAndChown(addedFolders string, fileIDs FileIDs) error {
	addedFoldersSplitted := strings.Split(addedFolders, string(filepath.Separator))
	walkPath := targetFolder
	for _, f := range addedFoldersSplitted {
		if f == "" {
			continue
		}

		walkPath = filepath.Join(walkPath, f)
		if _, err := os.Stat(walkPath); err != nil {
			if os.IsNotExist(err) {
				// file does not exist
				err := os.Mkdir(walkPath, os.ModePerm)
				if err != nil {
					log.Println(err)
					continue
				}
				f, err := os.Open(walkPath)
				if err != nil {
					log.Println(err)
					continue
				}
				err = f.Chown(fileIDs.UserID, fileIDs.GroupID)
				if err != nil {
					log.Println(err)
					continue
				}
				log.Printf("Updated folder %s: UID=%v, GUID=%v \n", walkPath, fileIDs.UserID, fileIDs.GroupID)
			} else {
				return err
			}
		}
	}
	return nil
}

func IsAlreadyConverted(folderPath string, fileName string) bool {
	jpgFile := fileName + ".jpg"

	addedFolders, _ := strings.CutPrefix(folderPath, watchingFolders)
	destinationDir := filepath.Join(targetFolder, addedFolders)
	targetDirEntries, err := os.ReadDir(destinationDir)

	// Folder does not exist, so file does not exist either
	if err != nil {
		return false
	}

	for _, item := range targetDirEntries {
		if item.Name() == jpgFile {
			return true
		}
	}
	return false
}

func findFileIDs(item fs.DirEntry) FileIDs {
	var result FileIDs
	info, err := item.Info()
	if err != nil {
		result.UserID = -1
		result.GroupID = -1
		return result
	}

	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		result.UserID = int(stat.Uid)
		result.GroupID = int(stat.Gid)
	} else {
		// we are not in linux, this won't work anyway in windows,
		// but maybe you want to log warnings
		result.UserID = -1
		result.GroupID = -1
	}
	return result
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
