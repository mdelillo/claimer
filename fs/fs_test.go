package fs_test

import (
	. "github.com/mdelillo/claimer/fs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"path/filepath"
)

var _ = Describe("Fs", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "claimer-fs-unit-tests")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Describe("Ls", func() {
		It("lists non-hidden files in a directory", func() {
			firstFile := "some-file"
			secondFile := "some-other-file"

			writeFile(filepath.Join(tempDir, firstFile), nil)
			writeFile(filepath.Join(tempDir, secondFile), nil)
			writeFile(filepath.Join(tempDir, ".some-hidden-file"), nil)
			mkdir(filepath.Join(tempDir, "some-directory"))

			Expect(NewFs().Ls(tempDir)).To(Equal([]string{firstFile, secondFile}))
		})

		Context("when listing the directory fails", func() {
			It("returns an error", func() {
				_, err := NewFs().Ls("some-bad-dir")
				Expect(err).To(MatchError(ContainSubstring("failed to list directory: ")))
			})
		})
	})

	Describe("LsDirs", func() {
		It("lists non-hidden directories", func() {
			firstDir := "some-directory"
			secondDir := "some-other-directory"

			mkdir(filepath.Join(tempDir, firstDir))
			mkdir(filepath.Join(tempDir, secondDir))
			mkdir(filepath.Join(tempDir, ".some-hidden-directory"))
			writeFile(filepath.Join(tempDir, "some-file"), nil)

			Expect(NewFs().LsDirs(tempDir)).To(Equal([]string{firstDir, secondDir}))
		})

		Context("when listing the directory fails", func() {
			It("returns an error", func() {
				_, err := NewFs().LsDirs("some-bad-dir")
				Expect(err).To(MatchError(ContainSubstring("failed to list directory: ")))
			})
		})
	})

	Describe("Mv", func() {
		Context("when src exists and dst does not exist", func() {
			It("moves src to dst", func() {
				src := filepath.Join(tempDir, "some-src-path")
				dst := filepath.Join(tempDir, "some-dst-path")
				contents := []byte("some-contents")

				writeFile(src, contents)

				Expect(NewFs().Mv(src, dst)).To(Succeed())
				Expect(src).NotTo(BeAnExistingFile())
				Expect(ioutil.ReadFile(dst)).To(Equal(contents))

			})
		})

		Context("when src and dst exist", func() {
			It("replaces dst with src", func() {
				src := filepath.Join(tempDir, "some-src-path")
				dst := filepath.Join(tempDir, "some-dst-path")
				contents := []byte("some-contents")

				writeFile(src, contents)
				writeFile(dst, []byte("some-old-contents"))

				Expect(NewFs().Mv(src, dst)).To(Succeed())
				Expect(src).NotTo(BeAnExistingFile())
				Expect(ioutil.ReadFile(dst)).To(Equal(contents))

			})
		})

		Context("when src does not exist", func() {
			It("returns an error", func() {
				src := filepath.Join(tempDir, "some-src-path")
				dst := filepath.Join(tempDir, "some-dst-path")

				Expect(NewFs().Mv(src, dst)).To(MatchError(ContainSubstring("failed to move file:")))
			})
		})
	})

	Describe("Rm", func() {
		It("recursively removes the path", func() {
			mkdir(filepath.Join(tempDir, "some-dir"))
			writeFile(filepath.Join(tempDir, "some-dir", "some-file"), nil)

			Expect(NewFs().Rm(tempDir)).To(Succeed())
			Expect(tempDir).NotTo(BeADirectory())
		})
	})

	Describe("Touch", func() {
		It("touches a file and creates any required directories", func() {
			file := filepath.Join(tempDir, "some", "nested", "file")
			Expect(NewFs().Touch(file)).To(Succeed())
			Expect(file).To(BeAnExistingFile())
		})

		Context("when creating the directory fails", func() {
			It("returns an error", func() {
				notADirectory := filepath.Join(tempDir, "not-a-directory")
				writeFile(notADirectory, nil)

				path := filepath.Join(notADirectory, "some-file")
				Expect(NewFs().Touch(path)).To(MatchError(ContainSubstring("failed to create directory:")))
			})
		})
	})
})

func writeFile(path string, contents []byte) {
	ExpectWithOffset(1, ioutil.WriteFile(path, contents, 0644)).To(Succeed())
}

func mkdir(dir string) {
	ExpectWithOffset(1, os.Mkdir(dir, 0755)).To(Succeed())
}
