package commands

import (
	"io/ioutil"
	"strings"
)

func Ls(dirName string, showHidden bool) ([]string, error) {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		return nil, err
	}

	dirs := []string{}
	dirFiles := []string{}

	for _, file := range files {
		if !showHidden && strings.HasPrefix(file.Name(), ".") {
			continue
		}
		if file.IsDir() {
			dirs = append(dirs, file.Name())
			continue
		}

		dirFiles = append(dirFiles, file.Name())
	}
	dirs = append(dirs, dirFiles...)

	return dirs, err
}
