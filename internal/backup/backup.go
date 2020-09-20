package backup

import (
	"errors"
	"fmt"
	"github.com/pesoklp13/go-scripts/pkg/compression"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func CreateBackup(gzipTarHelper compression.GzipTarHelper, source string, project string, destination string) error {
	return gzipTarHelper.Compress(
		fmt.Sprintf("%s%s.%s.tar.gz",
			destination,
			project,
			time.Now().Format("20060102150405"),
		),
		source,
	)
}

func RestoreBackup(gzipTarHelper compression.GzipTarHelper, backupFile string, destination string) error {
	project := retrieveProjectName(backupFile)

	if project == nil {
		return errors.New("unable to retrieve project name")
	}

	target := destination + *project + "/"

	err := os.RemoveAll(target)

	if err != nil {
		return errors.New(fmt.Sprintf("Unable to restore backup %s. Reason: %e", backupFile, err))
	}

	return gzipTarHelper.Uncompress(backupFile, target)
}

func retrieveProjectName(backupFile string) *string {
	base := filepath.Base(backupFile)

	r := regexp.MustCompile("^(?P<project>.*)+\\.[\\d]{14}\\.tar\\.gz$")
	match := r.FindStringSubmatch(base)
	for i, name := range r.SubexpNames() {
		if i != 0 && name == "project" {
			return &match[i]
		}
	}

	return nil
}
