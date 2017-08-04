package main

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Files is FileInfo array
type Files []os.FileInfo

func (files Files) Len() int {
	return len(files)
}

func (files Files) Less(i, j int) bool {
	return files[i].Name() < files[j].Name()
}

func (files Files) Swap(i, j int) {
	files[i], files[j] = files[j], files[i]
}

func isImage(ext string) bool {
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
}

// ListImages list image files in the given directory.
func ListImages(dir string) ([]os.FileInfo, error) {
	var result Files
	files, err := ioutil.ReadDir(dir)

	// Failed to read directory
	if err != nil {
		return result, err
	}

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if isImage(ext) {
			result = append(result, file)
		}
	}

	sort.Sort(result)
	return result, nil
}

func GetExt(filename string) string {
	return strings.ToLower(path.Ext(filename))
}

func GetBase(filename string) string {
	base := path.Base(filename)
	return base[:len(base)-len(path.Ext(filename))]
}
