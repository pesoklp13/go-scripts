package compression

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type closeable interface {
	Close() error
}

func closeStream(stream closeable) {
	err := stream.Close()

	if err != nil {
		log.Fatal("Unable to close resource")
	}
}

type GzipTarHelper interface {
	Compress(destination string, source string) error
	Uncompress(source string, destination string) error
}

type GzipTarHelperImpl struct {
}

func (helper *GzipTarHelperImpl) Compress(destination string, source string) error {
	destinationPath, err := filepath.Abs(destination)
	destinationPath = filepath.ToSlash(destinationPath)

	sourceAbsPath, err := filepath.Abs(source)
	sourceAbsPath = filepath.ToSlash(sourceAbsPath)

	sourceDir := filepath.Dir(destinationPath)

	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		err := os.MkdirAll(sourceDir, os.ModeDir)

		if err != nil {
			return err
		}
	}

	tarFile, err := os.Create(destinationPath)

	if err != nil {
		return err
	}

	defer closeStream(tarFile)

	gzWriter := gzip.NewWriter(tarFile)
	defer closeStream(gzWriter)

	tarWriter := tar.NewWriter(gzWriter)
	defer closeStream(tarWriter)

	dir, file := filepath.Split(sourceAbsPath)

	var size int

	isFile := len(filepath.Ext(file)) > 0

	if isFile {
		size = len(dir)
	} else {
		size = len(sourceAbsPath)
	}

	walker := func(file string, fi os.FileInfo, err error) error {
		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		fileAbsPath, err := filepath.Abs(file)
		fileAbsPath = filepath.ToSlash(fileAbsPath)

		if err != nil {
			return err
		}

		// must provide real name
		// (see https://golang.org/src/archive/tar/common.go?#L626)
		if len(fileAbsPath) > size {
			header.Name = filepath.ToSlash(fileAbsPath[size+1:])
		} else {
			header.Name = ""
		}

		// write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// if not a dir, write file content
		if !fi.IsDir() {
			err = compressFile(file, tarWriter)
		}

		return err
	}

	if !isFile {
		// handle folder
		err = filepath.Walk(source, walker)

		return err
	}

	// handle single file

	// write header for file into archive
	// without this it will cause "write too long" error
	fileInfo, err := os.Stat(source)

	if err != nil {
		return err
	}

	header := &tar.Header{
		Name: "",
		Mode: 0600,
		Size: fileInfo.Size(),
	}

	err = tarWriter.WriteHeader(header)
	err = compressFile(source, tarWriter)

	return err
}

func compressFile(file string, tarWriter *tar.Writer) error {
	data, err := os.Open(file)

	defer closeStream(data)

	if err != nil {
		return err
	}

	if _, err := io.Copy(tarWriter, data); err != nil {
		return err
	}

	return nil
}

func (helper *GzipTarHelperImpl) Uncompress(source string, destination string) error {
	absPath, err := filepath.Abs(source)
	absPath = filepath.ToSlash(absPath)

	tarFile, err := os.Open(absPath)

	if err != nil {
		return err
	}

	defer closeStream(tarFile)

	gzReader, err := gzip.NewReader(tarFile)

	if err != nil {
		return err
	}

	defer closeStream(gzReader)

	tarReader := tar.NewReader(gzReader)

	absDestinationPath, err := filepath.Abs(destination)

	// untar each segment
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		// determine proper file path info
		fileInfo := header.FileInfo()
		fileName := header.Name
		absFileName := filepath.Join(absDestinationPath, fileName)

		// if a dir, create it, then go to next segment
		if fileInfo.Mode().IsDir() {
			if err := os.MkdirAll(absFileName, 0755); err != nil {
				return err
			}
			continue
		}

		// create new file with original file mode
		file, err := os.OpenFile(
			absFileName,
			os.O_RDWR|os.O_CREATE|os.O_TRUNC,
			fileInfo.Mode().Perm(),
		)

		if err != nil {
			return err
		}

		fmt.Printf("x %s\n", absFileName)
		n, cpErr := io.Copy(file, tarReader)

		if closeErr := file.Close(); closeErr != nil {
			return err
		}

		if cpErr != nil {
			return cpErr
		}

		if n != fileInfo.Size() {
			return fmt.Errorf("wrote %d, want %d", n, fileInfo.Size())
		}
	}

	return nil
}

func NewGzipTarHelper() GzipTarHelper {
	return &GzipTarHelperImpl{}
}
