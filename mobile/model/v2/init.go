package model

import (
	"os/exec"
	"strings"
	"time"
)

func init() {
	fixTimezone()
}

// ref: https://github.com/golang/go/issues/20455#issuecomment-342287698
// fix time zone for time package on mobile environment
func fixTimezone() {
	out, err := exec.Command("/system/bin/getprop", "persist.sys.timezone").Output()
	if err != nil {
		return
	}
	z, err := time.LoadLocation(strings.TrimSpace(string(out)))
	if err != nil {
		return
	}
	time.Local = z
}
