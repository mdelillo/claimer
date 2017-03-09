package fs_test

import (
	"github.com/mdelillo/claimer/fs"

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

	Describe("Mv", func() {
		Context("when src exists and dst does not exist", func() {
			It("moves src to dst", func() {
				src := filepath.Join(tempDir, "some-src-path")
				dst := filepath.Join(tempDir, "some-dst-path")
				contents := []byte("some-contents")

				writeFile(src, contents)

				Expect(fs.NewFs().Mv(src, dst)).To(Succeed())
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

				Expect(fs.NewFs().Mv(src, dst)).To(Succeed())
				Expect(src).NotTo(BeAnExistingFile())
				Expect(ioutil.ReadFile(dst)).To(Equal(contents))

			})
		})

		Context("when src does not exist", func() {
			It("returns an error", func() {
				src := filepath.Join(tempDir, "some-src-path")
				dst := filepath.Join(tempDir, "some-dst-path")

				Expect(fs.NewFs().Mv(src, dst)).To(MatchError(ContainSubstring("failed to move file:")))
			})
		})
	})

	Describe("Ls", func() {
		It("lists files in a directory", func() {
			firstFile := "some-file"
			secondFile := "some-other-file"

			writeFile(filepath.Join(tempDir, firstFile), nil)
			writeFile(filepath.Join(tempDir, secondFile), nil)
			mkdir(filepath.Join(tempDir, "some-directory"))

			Expect(fs.NewFs().Ls(tempDir)).To(Equal([]string{firstFile, secondFile}))
		})

		Context("when listing the directory fails", func() {
			It("returns an error", func() {
				_, err := fs.NewFs().Ls("some-bad-dir")
				Expect(err).To(MatchError(ContainSubstring("failed to list directory: ")))
			})
		})
	})

	Describe("Rm", func() {
		It("removes path and any children", func() {
			writeFile(filepath.Join(tempDir, "some-file"), nil)
			mkdir(filepath.Join(tempDir, "some-directory"))
			writeFile(filepath.Join(tempDir, "some-directory", "some-file"), nil)

			Expect(fs.NewFs().Rm(tempDir)).To(Succeed())
			Expect(tempDir).NotTo(BeADirectory())
		})

	})
})

func writeFile(path string, contents []byte) {
	ExpectWithOffset(1, ioutil.WriteFile(path, contents, 0644)).To(Succeed())
}

func mkdir(dir string) {
	ExpectWithOffset(1, os.Mkdir(dir, 0755)).To(Succeed())
}
