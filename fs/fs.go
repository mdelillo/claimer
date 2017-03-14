package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type filesystem struct{}

func NewFs() *filesystem {
	return &filesystem{}
}

func (*filesystem) Mv(src, dst string) error {
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("failed to move file: %s", err)
	}
	return nil
}

func (*filesystem) Ls(dir string) ([]string, error) {
	var files []string

	children, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %s", err)
	}

	for _, child := range children {
		if !child.IsDir() && !isHidden(child) {
			files = append(files, child.Name())
		}
	}

	return files, nil
}

func (*filesystem) LsDirs(dir string) ([]string, error) {
	var dirs []string

	children, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %s", err)
	}

	for _, child := range children {
		if child.IsDir() && !isHidden(child) {
			dirs = append(dirs, child.Name())
		}
	}

	return dirs, nil
}

func isHidden(fileInfo os.FileInfo) bool {
	return strings.HasPrefix(fileInfo.Name(), ".")
}
