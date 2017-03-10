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
	Describe("Clone", func() {
		It("clones public repos without authorization", func() {
			repo := git.NewRepo("https://github.com/octocat/Hello-World", "")
			Expect(repo.Clone()).To(Succeed())
			defer os.RemoveAll(repo.Dir())

			Expect(runGitCommand(repo.Dir(), "status")).To(ContainSubstring("working tree clean"))
		})

		It("clones private repos using an SSH key", func() {
			repoUrl := getEnv("CLAIMER_TEST_REPO_URL")
			deployKey := getEnv("CLAIMER_TEST_DEPLOY_KEY")

			repo := git.NewRepo(repoUrl, deployKey)
			Expect(repo.Clone()).To(Succeed())
			defer os.RemoveAll(repo.Dir())

			Expect(runGitCommand(repo.Dir(), "status")).To(ContainSubstring("working tree clean"))
		})

		Context("when the SSH key is invalid", func() {
			It("returns an error", func() {
				repoUrl := getEnv("CLAIMER_TEST_REPO_URL")

				repo := git.NewRepo(repoUrl, "some-invalid-deploy-key")
				Expect(repo.Clone()).To(MatchError(ContainSubstring("failed to parse public key: ")))
			})
		})

		Context("when the URL is invalid", func() {
			It("returns an error", func() {
				repo := git.NewRepo("some-invalid-url", "")
				Expect(repo.Clone()).To(MatchError(ContainSubstring("failed to clone repo: ")))
			})
		})
	})

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
			Expect(repo.Clone()).To(Succeed())
			defer os.RemoveAll(repo.Dir())
			Expect(ioutil.WriteFile(filepath.Join(repo.Dir(), newFileName), nil, 0644)).To(Succeed())

			Expect(repo.CommitAndPush(commitMessage)).To(Succeed())
			commit := runGitCommand(repo.Dir(), "log", "origin/master", "-1", "--name-only", "--format='%s'")
			Expect(commit).To(Equal(fmt.Sprintf("'%s'\n\n%s\n", commitMessage, newFileName)))
		})

		Context("when committing fails", func() {
			It("returns an error", func() {
				repo := git.NewRepo("file://"+gitRemoteDir, "")
				Expect(repo.Clone()).To(Succeed())
				defer os.RemoveAll(repo.Dir())

				err := repo.CommitAndPush("some-commit-message")
				Expect(err).To(MatchError(MatchRegexp("(?s:failed to commit: .*nothing to commit)")))
			})
		})

		Context("when pushing fails", func() {
			It("returns an error", func() {
				repo := git.NewRepo("file://"+gitRemoteDir, "")
				Expect(repo.Clone()).To(Succeed())
				defer os.RemoveAll(repo.Dir())
				Expect(ioutil.WriteFile(filepath.Join(repo.Dir(), "some-new-file"), nil, 0644)).To(Succeed())
				runGitCommand(repo.Dir(), "remote", "remove", "origin")

				err := repo.CommitAndPush("some-commit-message")
				Expect(err).To(MatchError(MatchRegexp("(?s:failed to push: .*'origin' does not appear to be a git repository)")))
			})
		})
	})
})

func getEnv(name string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		Fail(fmt.Sprintf("%s must be set", name))
	}
	return value
}

func runGitCommand(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), string(output))
	return string(output)
}
