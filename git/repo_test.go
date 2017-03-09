package git_test

import (
	"github.com/mdelillo/claimer/git"

	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

var _ = Describe("Repo", func() {
	Describe("CommitAndPush", func() {
		var gitRemoteDir string

		BeforeEach(func() {
			var err error

			gitRemoteDir, err = ioutil.TempDir("", "claimer-git-unit-tests")
			Expect(err).NotTo(HaveOccurred())

			runGitCommand(gitRemoteDir, "init", ".")
			runGitCommand(gitRemoteDir, "config", "receive.denyCurrentBranch", "updateInstead")
			runGitCommand(gitRemoteDir, "commit", "--allow-empty", "-m", "Initial commit")
		})

		AfterEach(func() {
			os.RemoveAll(gitRemoteDir)
		})

		It("commits and pushes all changes to the repo", func() {
			commitMessage := "some-commit-message"
			newFileName := "some-new-file"

			repo := git.NewRepo("file://"+gitRemoteDir, "")
			defer os.RemoveAll(repo.Dir)
			Expect(ioutil.WriteFile(filepath.Join(repo.Dir, newFileName), nil, 0644)).To(Succeed())
			repo.CommitAndPush(commitMessage)

			commit := runGitCommand(repo.Dir, "log", "origin/master", "-1", "--name-only", "--format='%s'")
			Expect(commit).To(Equal(fmt.Sprintf("'%s'\n\n%s\n", commitMessage, newFileName)))
		})
	})
})

func runGitCommand(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), string(output))
	return string(output)
}
