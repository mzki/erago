package csv

import (
	"github.com/mzki/erago/filesystem"
)

// check whether file exists. It wraps filesystem.Exist()
// so that user need not to import filesystem package explicitly.
func FileExists(file string) bool {
	return filesystem.Exist(file)
}
