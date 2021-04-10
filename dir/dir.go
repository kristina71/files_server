package dir

import (
	"encoding/json"
	"files_server/commands"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Dir struct {
	path string
}

func New() *Dir {
	return &Dir{path: "/Users"}
}

func (currentDir *Dir) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/rm":
		currentDir.rm(w, r)
	case "/mkdir":
		currentDir.mkdir(w, r)
	case "/touch":
		currentDir.touch(w, r)
	case "/pwd":
		currentDir.pwd(w)
	case "/ls":
		currentDir.ls(w, r)
	case "/cd":
		currentDir.cd(w, r)
	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

func (currentDir *Dir) pwd(w http.ResponseWriter) {
	w.Write([]byte(currentDir.path))
}

func (currentDir *Dir) cd(w http.ResponseWriter, r *http.Request) {
	dir := r.URL.Query().Get("dir")

	if !strings.HasPrefix(dir, "/") {
		dir = filepath.Join(currentDir.path, dir)
	}
	fileInfo, err := os.Stat(dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !fileInfo.IsDir() {
		http.Error(w, "Not directory", http.StatusBadRequest)
		return
	}

	currentDir.path = dir
}

func (currentDir *Dir) ls(w http.ResponseWriter, r *http.Request) {
	hide := false
	if r.URL.Query().Get("hide") == "true" {
		hide = true
	}
	dir, err := commands.Ls(currentDir.path, hide)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res, err := json.Marshal(dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(res)
}

func (currentDir *Dir) mkdir(w http.ResponseWriter, r *http.Request) {
	dirName := r.URL.Query().Get("dirname")
	if dirName == "" {
		http.Error(w, "No dirname", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(dirName, "/") {
		dirName = filepath.Join(currentDir.path, dirName)
	}

	if dirName == "/" {
		return
	}

	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (currentDir *Dir) touch(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("filename")
	if fileName == "" {
		http.Error(w, "No filename", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(fileName, "/") {
		fileName = filepath.Join(currentDir.path, fileName)
	}

	_, err := os.Stat(fileName)
	if !os.IsNotExist(err) {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	file, err := os.Create(fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	file.Close()
}

func (currentDir *Dir) rm(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("filename")
	if fileName == "" {
		http.Error(w, "No filename", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(fileName, "/") {
		fileName = filepath.Join(currentDir.path, fileName)
	}

	if fileName == "/" {
		http.Error(w, "Can't delete root directory", http.StatusBadRequest)
		return
	}

	err := os.RemoveAll(fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
