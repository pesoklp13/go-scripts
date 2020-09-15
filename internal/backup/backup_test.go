package backup_test

import (
	"errors"
	"fmt"
	"github.com/pesoklp13/go-scripts/internal/backup"
	"regexp"
	"testing"
)

type GzipTarHelperMock struct {
	compressCallback func(destination string, source string) error
}

func (helper *GzipTarHelperMock) Compress(destination string, source string) error {
	return helper.compressCallback(destination, source)
}

func (helper *GzipTarHelperMock) Uncompress(string, string) error {
	return nil
}

func TestCreateBackup(t *testing.T) {
	const SOURCE = "source-dir"
	const PROJECT = "test-project"
	const DESTINATION = "/destination/path/"

	err := backup.CreateBackup(&GzipTarHelperMock{
		compressCallback: func(destination string, source string) error {
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
}
