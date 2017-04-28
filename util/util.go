package util

import (
	"os"
)

// return existance of file
func FileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}
