package cli

import (
	"fmt"
	"runtime"
)

// Version is the current version of ZenoEngine
const Version = "1.0.0"

// HandleVersion prints the current version of ZenoEngine
func HandleVersion() {
	fmt.Printf("ZenoEngine version %s %s/%s\n", Version, runtime.GOOS, runtime.GOARCH)
}
