package fs

import (
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type filesystem struct{}

func NewFs() *filesystem {
	return &filesystem{}
}

func (*filesystem) Ls(dir string) ([]string, error) {
	var files []string

	children, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list directory")
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
		return nil, errors.Wrap(err, "failed to list directory")
	}

	for _, child := range children {
		if child.IsDir() && !isHidden(child) {
			dirs = append(dirs, child.Name())
		}
	}

	return dirs, nil
}

func (*filesystem) Mv(src, dst string) error {
	if err := os.Rename(src, dst); err != nil {
		return errors.Wrap(err, "failed to move file")
	}
	return nil
}

func (*filesystem) Rm(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return errors.Wrap(err, "failed to remove path")
	}
	return nil
}

func (*filesystem) Touch(file string) error {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, "failed to create directory")
	}
	return ioutil.WriteFile(file, nil, 0644)
}

func isHidden(fileInfo os.FileInfo) bool {
	return strings.HasPrefix(fileInfo.Name(), ".")
}
