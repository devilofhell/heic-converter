package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestConvertFolder(t *testing.T) {
	err := PrepareTestEnvironment()
	if err != nil {
		t.Fatalf("Could not setup testenvironment: %v", err)
	}
	convertingFolder := "./testdata/01"
	heicList, _ := FindHeicFiles(convertingFolder)

	WalkDeleteHeic(convertingFolder)

	if ContainsHeicFiles(convertingFolder) {
		t.Errorf("did not delete heic files")
	}
	if missingFiles, ok := JPGsWereCreated(convertingFolder, heicList); !ok {
		var missingFilesString string
		for _, s := range missingFiles {
			missingFilesString = fmt.Sprintf("%s%s\n", missingFilesString, s)
		}
		t.Errorf("Following Files were not converted into JPG: \n%s", missingFilesString)
	}
}

func JPGsWereCreated(convertingFolder string, fileList []string) ([]string, bool) {
	resultFiles, _ := FindJPGFiles(convertingFolder)
	var missingFiles []string

	for _, file := range fileList {
		if ContainsFile(file, resultFiles) {
			continue
		}
		missingFiles = append(missingFiles, file)
	}
	return missingFiles, len(missingFiles) == 0
}

func findFiles(folder string, possibleFileTypes []string) ([]string, error) {
	files, _ := ioutil.ReadDir(folder)
	var result []string

	for _, f := range files {
		splittedFile := strings.Split(f.Name(), ".")
		if len(splittedFile) != 2 {
			continue
		}
		fileName := splittedFile[0]
		fileFormat := splittedFile[1]
		for _, s := range possibleFileTypes {
			if fileFormat == s {
				result = append(result, fileName)
			}
		}
	}
	return result, nil
}

func FindHeicFiles(folder string) ([]string, error) {
	return findFiles(folder, []string{"HEIC", "heic"})
}

func FindJPGFiles(folder string) ([]string, error) {
	return findFiles(folder, []string{"JPG", "jpg", "JPEG", "jpeg"})
}

func ContainsHeicFiles(folder string) bool {
	if files, err := FindHeicFiles(folder); err != nil && len(files) == 0 {
		return true
	}
	return false
}

func ContainsFile(fileName string, files []string) bool {
	for _, f := range files {
		if f == fileName {
			return true
		}
	}
	return false
}
