package object_service

import (
	"fmt"
	"os"
)

// @TODO: Change the implementation of the closeFile function
func closeFile(f *os.File) {
	err := f.Close()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
