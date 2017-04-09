package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const FILE_PATH = "./duokan/"

type FileDesc struct {
	Id   int    `json:"id"`
	Size int64  `json:"size"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func init() {
	err := initFilePath(FILE_PATH)
	if err != nil {
		fmt.Println(err)
	}
}
func main() {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("www"))
	mux.Handle("/", fs)
	mh := http.HandlerFunc(handleRequest)
	mux.Handle("/files/", mh)
	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}

	fmt.Println("Now listening on: http://0.0.0.0:8080")
	fmt.Println("Application started. Press Ctrl+C to shut down.")
	server.ListenAndServe()
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var err error
	switch r.Method {
	case "GET":
		err = handleGet(w, r)
	case "POST":
		err = handlePost(w, r)
	case "DELETE":
		err = handleDelete(w, r)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) (err error) {
	fmt.Println("GET: " + r.URL.Path)
	name := path.Base(r.URL.Path)
	if strings.Contains(name, ".") {
		fmt.Println("download: " + name)
		file := FILE_PATH + name
		if exist, _ := exists(file); !exist {
			http.NotFound(w, r)
		}
		http.ServeFile(w, r, file)
		return
	} else {
		_dirpath, err1 := os.Open(FILE_PATH)
		if err1 != nil {
			err = err1
			return
		}
		_dir, err1 := _dirpath.Readdir(0)
		if err1 != nil {
			err = err1
			return
		}
		files := make([]FileDesc, len(_dir))
		for i, f := range _dir {
			files[i].Id = i + 1
			files[i].Name = f.Name()
			files[i].Path = ""
			files[i].Size = f.Size()

		}
		output, err1 := json.MarshalIndent(&files, "", "\t\t")
		if err1 != nil {
			err = err1
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(output)
		return
	}

}

func handlePost(w http.ResponseWriter, r *http.Request) (err error) {
	fmt.Println("POST: " + r.URL.Path)
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("newfile")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Fprintf(w, "%v", handler.Header)
	f, err := os.OpenFile(FILE_PATH+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)
	return
}
func handleDelete(w http.ResponseWriter, r *http.Request) (err error) {
	fmt.Println("DELETE: " + r.URL.Path)
	name := path.Base(r.URL.Path)
	fmt.Println("DELETE: " + name)
	err = os.Remove(FILE_PATH + name)
	if err != nil {
		fmt.Println(err)
	}
	w.WriteHeader(200)
	return
}
func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}
func initFilePath(filePath string) (err error) {
	if _, err = exists(filePath); err != nil {
		return
	}
	err = os.Mkdir(filePath, 0777)
	return
}
func exists(path string) (bool, error) {

	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}
func check(name string) bool {
	ext := []string{".gif", ".pdf", ".png", ".jpg", ".epub", ".txt"}

	for _, v := range ext {
		if v == name {
			return false
		}
	}
	return true
}
