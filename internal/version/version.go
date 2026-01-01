package version

import (
	"fmt"
	"runtime"
)

// injected via -ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func Info() string {
	return fmt.Sprintf("multichat %s (commit: %s, built: %s, %s/%s)",
		Version, Commit, BuildTime, runtime.GOOS, runtime.GOARCH)
}
