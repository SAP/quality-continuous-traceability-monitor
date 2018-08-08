package utils

import (
	"time"

	"github.com/golang/glog"
)

// TimeTrack tracks performance runtime and logs it out
// Call: defer utils.TimeTrack(time.Now(), "Some meaningful comment")
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	glog.Info(name, " took ", elapsed)
}
