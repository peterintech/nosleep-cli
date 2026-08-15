package defaults

import (
	"runtime"
	"time"
)

const (
	Interval          = 10 * time.Second
	Sample            = 5 * time.Second
	Quiet             = 3 * time.Minute
	PowerTestDuration = 2 * time.Minute
)

func CPUThreshold() time.Duration {
	if runtime.GOOS == "linux" {
		return 10 * time.Millisecond
	}
	return 250 * time.Millisecond
}
