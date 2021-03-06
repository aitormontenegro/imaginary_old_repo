package main

import (
    "os"
	"io/ioutil"
	"net/http"
	"path"
    "fmt"
	"strings"
	"path/filepath"

//	"gopkg.in/h2non/bimg.v1"
//	"gopkg.in/h2non/filetype.v0"
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

	fmt.Printf("file ====> %s\n",file)

	file, cach, err := s.buildPath_orig(file)
	if err != nil {
		return nil, err
	}

	fmt.Printf("cach = %s\n",cach)

	if cach != "" {
//		fmt.Printf("Caching file...\n")
		c := make(chan int64)
		go defercache(file,cach,c)
	}

	//TODO: forzar caso extremo que falle escritura + full disk
	//TODO: cambiar la funciona para que haga un defer
	//TODO: test de estres

	fmt.Printf("file = %s\n",file)

	return s.read(file)
}

func (s *FileSystemImageSource) buildPath_orig(file string) (string, string, error) {
	// first --> return original file or cached file
	// second -> "" if cached file, string if file has to be cached
	// third --> error

    var fullcachedirpathandfile = s.Config.CacheDirPath + file
	file = path.Clean(path.Join(s.Config.MountPath, file))

	cach := ""

	if _, err := os.Stat(fullcachedirpathandfile); os.IsNotExist(err) {
//		fmt.Printf("Return original file path\n")
		cach = fullcachedirpathandfile
	}else{
//		fmt.Printf("Return cached file path\n")
		file = fullcachedirpathandfile
	}

    fmt.Printf("\nReturn file --> %s\n", file);
    fmt.Printf("\nReturn file --> %+v\n", s);
		if strings.HasPrefix(file, s.Config.MountPath) == false && strings.HasPrefix(file,s.Config.CacheDirPath) == false {
			return "","", ErrInvalidFilePath
		}

	return file, cach, nil

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
func dofilecache(src, dst string) (int64, error) {

	 var fullcachedirpath = filepath.Dir(dst);

	 if _, err := os.Stat(fullcachedirpath); os.IsNotExist(err) {
		err = os.MkdirAll(fullcachedirpath, 0770)
		if err != nil {
			fmt.Printf("mkdir recursive operation failed %q\n", err)
		}
	}

        sourceFileStat, err := os.Stat(src)
        if err != nil {
                return 0, err
        }
        if !sourceFileStat.Mode().IsRegular() {
                return 0, fmt.Errorf("%s is not a regular file", src)
        }

		source, err := ioutil.ReadFile(src)
        if err != nil {
                return 0, err
        }

		var o ImageOptions;
		o.Width = 1200;
		o.Height = 840;
		o.Quality = 80;
		o.Colorspace = 22;
		o.StripMetadata = true

		fmt.Printf("1. Saved quality = %d\n", o.Quality)

		image, err := Fit(source, o)

		var destinationFile = dst
		err = ioutil.WriteFile(destinationFile, image.Body, 0774)



		if err != nil {
			fmt.Println("Error creating file %s", destinationFile)
			fmt.Println(err)
			return 0, err
		}

		return int64(len(image.Body)), err

}
func defercache(src, dst string, c chan int64) () {
	nBytes, err := dofilecache(src, dst)
	if err != nil || nBytes == 0 {
		fmt.Printf("Copy operation to cache failed %q\n", err)
		err := os.Remove(dst)
		if err != nil {
			  fmt.Println(err)
			    return
		}
		//delete file
	} else {
		fmt.Printf("File cached!! (Image Generated: %d bytes, path: %s)\n", nBytes, dst)
	}
	c <- nBytes
	close(c)
}
