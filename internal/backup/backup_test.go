package backup_test

import (
	"errors"
	"fmt"
	"github.com/pesoklp13/go-scripts/internal/backup"
	"github.com/pesoklp13/go-scripts/internal/closeable"
	"github.com/pesoklp13/go-scripts/internal/paths"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

type GzipTarHelperMock struct {
	compressCallback   func(destination string, source string) error
	uncompressCallback func(source string, destination string) error
}

func (helper *GzipTarHelperMock) Compress(destination string, source string) error {
	return helper.compressCallback(destination, source)
}

func (helper *GzipTarHelperMock) Uncompress(source string, destination string) error {
	return helper.uncompressCallback(source, destination)
}

const PROJECT = "test-project"

func TestCreateBackup(t *testing.T) {
	const SOURCE = "source-dir"
	const DESTINATION = "/destination/path/"

	invoked := 0

	err := backup.CreateBackup(&GzipTarHelperMock{
		compressCallback: func(destination string, source string) error {
			invoked++

			if source != SOURCE {
				return errors.New(fmt.Sprintf("Source not match. Expected %s Given %s", SOURCE, source))
			}

			expectedValue := fmt.Sprintf("%s%s\\.[\\d]{14}\\.tar\\.gz", DESTINATION, PROJECT)
			match, err := regexp.Match(expectedValue, []byte(destination))

			if !match {
				return errors.New(fmt.Sprintf("Destination not match. Expected pattern: %s Givne %s", expectedValue, destination))
			}

			return err
		},
	}, SOURCE, PROJECT, DESTINATION)

	if err != nil {
		t.Fatalf("Test failed. %e", err)
	}

	if invoked == 0 {
		t.Fatal("Test failed. Compress was not called.")
	}
}

func TestRestoreBackup(t *testing.T) {
	// set up context
	destination := fmt.Sprintf("%s/%s/", paths.ProjectTempFolder, "nginx/path")
	err := os.MkdirAll(destination+PROJECT, os.ModeDir)

	if err != nil {
		t.Fatal("Unable to prepare context. test-project not created under tmp folder")
	}

	file, err := os.OpenFile(filepath.ToSlash(destination+"test-project/file-to-be-removed.txt"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	closeable.CloseStream(file, nil)

	if err != nil {
		t.Fatal("Unable to prepare context. Unable to create file-to-be-removed.txt under test-project")
	}

	invoked := 0

	const BackupFile = "./testdata/test-project.20060102150405.tar.gz"

	err = backup.RestoreBackup(
		&GzipTarHelperMock{
			uncompressCallback: func(source string, dest string) error {
				invoked++

				if source != BackupFile {
					return errors.New(fmt.Sprintf("Source not match. Expected %s Given %s", BackupFile, source))
				}

				expectedDestination := destination + PROJECT + "/"

				if dest != expectedDestination {
					return errors.New(fmt.Sprintf("Destination not match. Expected %s Given %s", expectedDestination, dest))
				}

				return nil
			},
		},
		BackupFile,
		destination,
	)

	if err != nil {
		t.Fatalf("Test failed. %e", err)
	}

	if invoked == 0 {
		t.Fatal("Test failed. Uncompress was not called.")
	}

	// tear down
	err = os.RemoveAll(paths.ProjectTempFolder)
}
