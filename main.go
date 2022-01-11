package main

import (
	"archive/zip"
	"fmt"
	flag "github.com/spf13/pflag"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var dir string
var zipFilename string
var tempFilename string
var omitName string

func main() {
	flag.Usage = func() {
		fmt.Printf("Zip Dir \n\nUSAGE:\n%s <filename> [OPTIONS]\n\nOPTIONS:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Println()
	}
	flag.StringVarP(&dir, "dir", "d", ".", "Directory be zipped")
	flag.Parse()

	if len(os.Args) <= 1 {
		flag.Usage()
		os.Exit(0)
	}
	if strings.HasSuffix(os.Args[1], "help") {
		flag.Usage()
		os.Exit(0)
	}
	if strings.HasSuffix(os.Args[1], "version") {
		fmt.Printf("tiny-zip %s (%s %s)\n", AppVersion, AppRevision, AppBuildDate)
		os.Exit(0)
	}

	file, err := ioutil.TempFile(".", "*.zip")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)
	tempFilename = file.Name()
	var index = strings.LastIndex(tempFilename, string(filepath.Separator))
	omitName = tempFilename[index+1:]
	if err := ZipDir(tempFilename, dir); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	zipFilename = os.Args[1]
	err = os.Rename(tempFilename, zipFilename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ZipDir(zipFile, dir string) error {
	// 1. Create a ZIP file and zip.Writer
	f, err := os.Create(zipFile)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(f)

	writer := zip.NewWriter(f)
	defer func(writer *zip.Writer) {
		err := writer.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(writer)

	// 2. Go through all the files of the dir
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 3. Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// set compression
		header.Method = zip.Deflate

		// 4. Set relative path of a file as the header name
		header.Name, err = filepath.Rel(filepath.Dir(dir), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		if strings.HasSuffix(path, omitName) {
			return nil
		}

		// 5. Create writer for the file header and save content of the file
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				fmt.Println(err)
			}
		}(f)

		_, err = io.Copy(headerWriter, f)
		return err
	})
}
