package cmd

import (
	"fmt"
	"os"
)

// Exists checks if the path exists
func Exists(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("%q does not exist", path)
	}
	return nil
}
