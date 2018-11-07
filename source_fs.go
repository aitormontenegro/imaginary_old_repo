package main

import (
	"io/ioutil"
	"net/http"
	"path"
	"strings"
    "fmt"
    "path/filepath"
)

const ImageSourceTypeFileSystem ImageSourceType = "fs"

type FileSystemImageSource struct {
	Config *SourceConfig
}

func NewFileSystemImageSource(config *SourceConfig) ImageSource {
	return &FileSystemImageSource{config}
}

func (s *FileSystemImageSource) Matches(r *http.Request) bool {
	return r.Method == "GET" && s.getFileParam(r) != ""
}

func (s *FileSystemImageSource) GetImage(r *http.Request) ([]byte, error) {

	file := s.getFileParam(r)
	if file == "" {
		return nil, ErrMissingParamFile
	}

	file, err := s.buildPath(file)
	if err != nil {
		return nil, err
	}

	return s.read(file)
}

func (s *FileSystemImageSource) buildPath(file string) (string, error) {
    var relativepath = file
	file = path.Clean(path.Join(s.Config.MountPath, file))
    var fullpath = file
    var fullcachedirpathandfile = s.Config.CacheDirPath + relativepath

    fmt.Printf("File --> %s\n",relativepath);
    fmt.Printf("File --> %s\n",fullpath);
    fmt.Printf("CacheDir --> %s\n\n",s.Config.CacheDirPath);
    fmt.Printf("Full Cache Dir and file --> %s\n\n",fullcachedirpathandfile);
    var fullcachedirpath = filepath.Dir(fullcachedirpathandfile);
    fmt.Printf("Full cache dir --> %s\n\n",fullcachedirpath);

	if strings.HasPrefix(file, s.Config.MountPath) == false {
		return "", ErrInvalidFilePath
	}
	return file, nil
}

func (s *FileSystemImageSource) read(file string) ([]byte, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, ErrInvalidFilePath
	}
	return buf, nil
}

func (s *FileSystemImageSource) getFileParam(r *http.Request) string {
	return r.URL.Query().Get("file")
}

func init() {
	RegisterSource(ImageSourceTypeFileSystem, NewFileSystemImageSource)
}
