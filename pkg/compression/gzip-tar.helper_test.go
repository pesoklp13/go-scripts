package compression_test

import (
	"bytes"
	"fmt"
	"github.com/pesoklp13/go-scripts/internal/paths"
	"github.com/pesoklp13/go-scripts/pkg/compression"
	"os"
	"path/filepath"
	"testing"
)

func createFileMap(m map[string]int) string {
	b := new(bytes.Buffer)
	for key := range m {
		_, err := fmt.Fprintf(b, "%s\n", key)

		if err != nil {
		}
	}
	return b.String()
}

func TestCompressAndUncompress(t *testing.T) {
	err := os.RemoveAll(paths.ProjectTempFolder + "/gzip-results")
	if err != nil {
		t.Errorf("Unable to setup test case")
	}

	for _, test := range []struct {
		name           string
		pathToCompress string
		expectedFile   string
		expectedPath   string
		dirName        string
		expectedFiles  map[string]int
	}{
		{
			name:           "empty folder",
			pathToCompress: "./testdata/empty/",
			expectedFile:   paths.ProjectTempFolder + "/gzip-results/empty.tar.gz",
			expectedPath:   paths.ProjectTempFolder + "/gzip-results/empty/",
			dirName:        "empty",
			expectedFiles: map[string]int{
				paths.ProjectTempFolder + "/gzip-results":              1,
				paths.ProjectTempFolder + "/gzip-results/empty":        1,
				paths.ProjectTempFolder + "/gzip-results/empty.tar.gz": 1,
			},
		},
		{
			name:           "folder with file",
			pathToCompress: "./testdata/folder-with-file/",
			expectedFile:   paths.ProjectTempFolder + "/gzip-results/folder-with-file.tar.gz",
			expectedPath:   paths.ProjectTempFolder + "/gzip-results/folder-with-file/",
			dirName:        "folder-with-file",
			expectedFiles: map[string]int{
				paths.ProjectTempFolder + "/gzip-results":                                    1,
				paths.ProjectTempFolder + "/gzip-results/folder-with-file":                   1,
				paths.ProjectTempFolder + "/gzip-results/folder-with-file/single-file-1.txt": 1,
				paths.ProjectTempFolder + "/gzip-results/folder-with-file.tar.gz":            1,
			},
		},
		{
			name:           "folder with folder",
			pathToCompress: "./testdata/folder-with-folder/",
			expectedFile:   paths.ProjectTempFolder + "/gzip-results/folder-with-folder.tar.gz",
			expectedPath:   paths.ProjectTempFolder + "/gzip-results/folder-with-folder/",
			dirName:        "folder-with-folder",
			expectedFiles: map[string]int{
				paths.ProjectTempFolder + "/gzip-results":                                                       1,
				paths.ProjectTempFolder + "/gzip-results/folder-with-folder":                                    1,
				paths.ProjectTempFolder + "/gzip-results/folder-with-folder/folder-with-file":                   1,
				paths.ProjectTempFolder + "/gzip-results/folder-with-folder/folder-with-file/single-file-3.txt": 1,
				paths.ProjectTempFolder + "/gzip-results/folder-with-folder/single-file-2.txt":                  1,
				paths.ProjectTempFolder + "/gzip-results/folder-with-folder.tar.gz":                             1,
			},
		},
		{
			name:           "simple file",
			pathToCompress: "./testdata/single-file.txt",
			expectedFile:   paths.ProjectTempFolder + "/gzip-results/single-file.tar.gz",
			expectedPath:   paths.ProjectTempFolder + "/gzip-results/single-file.txt",
			dirName:        "",
			expectedFiles: map[string]int{
				paths.ProjectTempFolder + "/gzip-results":                    1,
				paths.ProjectTempFolder + "/gzip-results/single-file.tar.gz": 1,
				paths.ProjectTempFolder + "/gzip-results/single-file.txt":    1,
			},
		},
	} {
		helper := compression.NewGzipTarHelper()

		err := helper.Compress(test.expectedFile, test.pathToCompress)
		if err != nil {
			t.Errorf("Test %q failed. Failed to compress file %s. Error is %e", test.name, test.pathToCompress, err)
		}

		err = helper.Uncompress(test.expectedFile, test.expectedPath)
		if err != nil {
			t.Errorf("Test %q failed. Failed to uncompress file %s. Error is %e", test.name, test.expectedFile, err)
		}

		expectedFiles := &test.expectedFiles
		foundFiles := map[string]int{}

		walkerFactory := func() filepath.WalkFunc {

			return func(path string, info os.FileInfo, err error) error {
				path = filepath.ToSlash(path)

				foundFiles[path] = 1

				delete(*expectedFiles, path)

				return nil
			}
		}

		err = filepath.Walk(paths.ProjectTempFolder+"/gzip-results", walkerFactory())

		err = os.RemoveAll(paths.ProjectTempFolder + "/gzip-results")
		if err != nil {
			t.Errorf("Unable to clear test context")
		}

		if len(*expectedFiles) != 0 {
			t.Errorf("Test %s failed. Bad file structure. \n\nExpected files: \n%s\nGiven files: \n%s", test.name, createFileMap(test.expectedFiles), createFileMap(foundFiles))
		}
	}

	err = os.Remove(paths.ProjectTempFolder)
}
