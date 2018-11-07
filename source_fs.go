package main

import (
    "os"
	"io"
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
    var fullcachedirpath = filepath.Dir(fullcachedirpathandfile);

    fmt.Printf("Path pedido --> %s\n",relativepath);
    fmt.Printf("Path pedido Full edition  --> %s\n",fullpath);
    fmt.Printf("CacheDir --> %s\n\n",s.Config.CacheDirPath);
    fmt.Printf("Full cache dir --> %s\n\n",fullcachedirpath);
    fmt.Printf("Full Cache Dir and file --> %s\n\n",fullcachedirpathandfile);

    if _, err := os.Stat(fullcachedirpath); os.IsNotExist(err) {
        err = r.MkdirAll(fullcachedirpath, 0770)
		if err != nil {
			fmt.Printf("mkdir recursive operation failed %q\n", err)
		}

		nBytes, err := copy(fullpath, fullcachedirpathandfile)
		if err != nil {
			fmt.Printf("The copy operation failed %q\n", err)
		} else {
			fmt.Printf("Copied %d bytes!\n", nBytes)
		}
    }else{
        if _, err := os.Stat(fullcachedirpathandfile); !os.IsNotExist(err) {
            file = fullcachedirpathandfile
          }else{
			nBytes, err := copy(fullpath, fullcachedirpathandfile)
			if err != nil {
				fmt.Printf("The copy operation failed %q\n", err)
			} else {
				fmt.Printf("Copied %d bytes!\n", nBytes)
			}
          }
    }




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
func copy(src, dst string) (int64, error) {
        sourceFileStat, err := os.Stat(src)
        if err != nil {
                return 0, err
        }

        if !sourceFileStat.Mode().IsRegular() {
                return 0, fmt.Errorf("%s is not a regular file", src)
        }

        source, err := os.Open(src)
        if err != nil {
                return 0, err
        }
        defer source.Close()

        destination, err := os.Create(dst)
        if err != nil {
                return 0, err
        }
        defer destination.Close()
        nBytes, err := io.Copy(destination, source)
        return nBytes, err
}
