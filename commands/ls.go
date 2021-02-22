package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

func Ls(dirName string, showHidden bool) {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		log.Fatal(err)
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

	for _, file := range dirs {
		fmt.Println(file)
	}
}
