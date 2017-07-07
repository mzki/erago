package model

import (
	"local/erago/width"
)

// textwidth.go expose methods in package erago/width to mobile device.

// because mobile can not be detect isEastAsian
// so set it explicitlly.
var cond = width.NewCondition(true)

func StringWidth(text string) int32 {
	return int32(cond.StringWidth(text))
}
