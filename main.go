package main

import (
	"path/filepath"
	"fmt"

	"strings"
)

func main() {
	// var dir string
	// rest := `c:\foo\bar\baz\cpi.exe`
	fullPath := `/foo/bar/baz/cpi.exe`
	fullDir := filepath.Dir(fullPath)
	for _, dir := range strings.Split(fullDir, "/") {
		fmt.Printf("dir: %s\n", dir)
	}
}
