package main

import (
	"git-light/cmd"
	"path/filepath"
)

func main() {
	cmd.Execute()
}

func FilterDirsGlob(dir, suffix string) ([]string, error) {
	return filepath.Glob(filepath.Join(dir, suffix))
}
