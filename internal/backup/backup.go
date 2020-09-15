package backup

import (
	"fmt"
	"github.com/pesoklp13/go-scripts/pkg/compression"
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
