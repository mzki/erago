package csv

import (
	"local/erago/util"
)

// check whether file exists. It wraps util.FileExists()
// so that user need not to import util package explicitly.
func FileExists(file string) bool {
	return util.FileExists(file)
}
