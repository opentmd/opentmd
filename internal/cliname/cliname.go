package cliname

import (
	"os"
	"path/filepath"
)

const (
	Long = "opentmd"
)

// Current returns the CLI name based on how the binary was invoked.
func Current() string {
	_ = filepath.Base(os.Args[0])
	return Long
}
