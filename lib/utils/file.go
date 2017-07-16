package utils

import (
	"os"
)

func FileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil {
		return !os.IsNotExist(err)
	}
	return true
}