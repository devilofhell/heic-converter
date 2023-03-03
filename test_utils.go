package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"
)

func PrepareTestEnvironment() error {
	testFolders := []string{"01", "02"}
	SetupTestFolder(testFolders)

	files, err := ioutil.ReadDir("./testdata/origin")
	if err != nil {
		return err
	}
	var counter int
	for _, f := range files {
		splittedFile := strings.Split(f.Name(), ".")
		if len(splittedFile) != 2 {
			return fmt.Errorf("File contains more than '.' Could not parse")
		}
		fileName := splittedFile[0]
		fileFormat := splittedFile[1]

		if fileFormat == "MOV" || fileFormat == "mov" {
			continue
		}
		if fileFormat == "HEIC" || fileFormat == "heic" {
			CopyToTestFolder(fileName, files, testFolders[counter % len(testFolders)])
			counter++
		}
	}
	return nil
}

func CopyToTestFolder(fileToCopy string, srcFolder []fs.FileInfo, targetFolder string) error {
	for _, f := range srcFolder {
		if strings.HasPrefix(f.Name(), fileToCopy) {
			b, _ := os.ReadFile(f.Name())
			target := fmt.Sprintf("./testdata/%s/%s", targetFolder, f.Name())
			os.WriteFile(target, b, 777)
		}
	}
	return nil
}

func SetupTestFolder(testFolder []string) {
	for _, folderName := range testFolder {
		if !ContainsFolder(folderName) {
			target := fmt.Sprintf("./testdata/%s", folderName)
			os.Mkdir(target, 777)
		}
	}
}

func ContainsFolder(folderName string) bool {
	testdata, _ := ioutil.ReadDir("./testdata")
	for _, f := range testdata {
		if f.IsDir() && f.Name() == folderName {
			return true
		}
	}
	return false
}