package utils

import (
	"fmt"
	"path/filepath"
)

func IncrementFileName(name string, n int) string {
	ext := filepath.Ext(name)
	base := name[:len(name)-len(ext)]

	return fmt.Sprintf("%s (%d)%s", base, n, ext)
}

func IncrementFolderName(name string, n int) string {
	return fmt.Sprintf("%s (%d)", name, n)
}
