package fs

import (
	"fmt"
	"io/ioutil"
	"os"
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
	var filenames []string

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %s", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			filenames = append(filenames, file.Name())
		}
	}

	return filenames, nil
}

func (*filesystem) Rm(dir string) error {
	return os.RemoveAll(dir)
}

func (*filesystem) TempDir(prefix string) (string, error) {
	return ioutil.TempDir("", prefix)
}
