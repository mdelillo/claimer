package git_test

import (
	. "github.com/mdelillo/claimer/git"

	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"srcd.works/go-git.v4"
)

var _ = Describe("Repo", func() {
	var gitDir string

	BeforeEach(func() {
		var err error
		gitDir, err = ioutil.TempDir("", "claimer-test-git-dir")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(gitDir)
	})

	Describe("CloneOrPull", func() {
		Context("when the directory already contains a repo", func() {
			Context("when the repo is public", func() {
				It("updates the repo", func() {
					repo := NewRepo("https://github.com/octocat/Hello-World", "", gitDir)
					Expect(repo.CloneOrPull()).To(Succeed())

					master := runGitCommand(gitDir, "rev-parse", "HEAD")
					runGitCommand(gitDir, "reset", "--hard", "HEAD~")

					Expect(repo.CloneOrPull()).To(Succeed())
					Expect(runGitCommand(gitDir, "rev-parse", "HEAD")).To(Equal(master))
				})
			})

			Context("when the repo is private", func() {
				It("updates the repo", func() {
					repoUrl := getEnv("CLAIMER_TEST_REPO_URL")
					deployKey := getEnv("CLAIMER_TEST_DEPLOY_KEY")

					repo := NewRepo(repoUrl, deployKey, gitDir)
					Expect(repo.CloneOrPull()).To(Succeed())

					master := runGitCommand(gitDir, "rev-parse", "HEAD")
					runGitCommand(gitDir, "reset", "--hard", "HEAD~")

					Expect(repo.CloneOrPull()).To(Succeed())
					Expect(runGitCommand(gitDir, "rev-parse", "HEAD")).To(Equal(master))
				})
			})
		})

		Context("when the directory does not contain a repo", func() {
			Context("when the repo is public", func() {
				It("clones the repo", func() {
					repo := NewRepo("https://github.com/octocat/Hello-World", "", gitDir)
					Expect(repo.CloneOrPull()).To(Succeed())
					Expect(runGitCommand(gitDir, "status")).To(ContainSubstring("working tree clean"))
				})
			})

			Context("when the repo is private", func() {
				It("clones the repo", func() {
					repoUrl := getEnv("CLAIMER_TEST_REPO_URL")
					deployKey := getEnv("CLAIMER_TEST_DEPLOY_KEY")

					repo := NewRepo(repoUrl, deployKey, gitDir)
					Expect(repo.CloneOrPull()).To(Succeed())
					Expect(runGitCommand(gitDir, "status")).To(ContainSubstring("working tree clean"))
				})
			})
		})

		Context("when the SSH key is invalid", func() {
			It("returns an error", func() {
				repoUrl := getEnv("CLAIMER_TEST_REPO_URL")

				repo := NewRepo(repoUrl, "some-invalid-deploy-key", gitDir)
				Expect(repo.CloneOrPull()).To(MatchError(ContainSubstring("failed to parse public key: ")))
			})
		})

		Context("when pulling the repo fails", func() {
			It("returns an error", func() {
				repo := NewRepo("https://github.com/octocat/Hello-World", "", gitDir)
				Expect(repo.CloneOrPull()).To(Succeed())

				runGitCommand(gitDir, "remote", "remove", "origin")

				Expect(repo.CloneOrPull()).To(MatchError(ContainSubstring("failed to pull repo: ")))
			})
		})

		Context("when cloning the repo fails", func() {
			It("returns an error", func() {
				repo := NewRepo("some-invalid-url", "", gitDir)
				Expect(repo.CloneOrPull()).To(MatchError(ContainSubstring("failed to clone repo: ")))
			})
		})
	})

	Describe("CommitAndPush", func() {
		var gitRemoteDir string
		var gitRemoteUrl string

		BeforeEach(func() {
			var err error

			gitRemoteDir, err = ioutil.TempDir("", "claimer-test-git-remote")
			Expect(err).NotTo(HaveOccurred())
			gitRemoteUrl = "file://" + gitRemoteDir

			runGitCommand(gitRemoteDir, "init", ".")
			runGitCommand(gitRemoteDir, "config", "receive.denyCurrentBranch", "updateInstead")
			runGitCommand(gitRemoteDir, "commit", "--allow-empty", "-m", "Initial commit")

			_, err = git.PlainClone(gitDir, false, &git.CloneOptions{URL: "file://" + gitRemoteDir})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(gitRemoteDir)
		})

		It("commits and pushes all changes to the repo", func() {
			commitMessage := "some-commit-message"
			newFileName := "some-new-file"

			touchFile(filepath.Join(gitDir, newFileName))

			repo := NewRepo(gitRemoteUrl, "", gitDir)
			Expect(repo.CommitAndPush(commitMessage)).To(Succeed())
			commit := runGitCommand(gitDir, "log", "origin/master", "-1", "--name-only", "--format='%s'")
			Expect(commit).To(Equal(fmt.Sprintf("'%s'\n\n%s\n", commitMessage, newFileName)))
		})

		Context("when committing fails", func() {
			It("returns an error", func() {
				repo := NewRepo(gitRemoteUrl, "", gitDir)
				err := repo.CommitAndPush("some-commit-message")
				Expect(err).To(MatchError(MatchRegexp("(?s:failed to commit: .*nothing to commit)")))
			})
		})

		Context("when pushing fails", func() {
			It("returns an error", func() {
				runGitCommand(gitDir, "remote", "remove", "origin")
				touchFile(filepath.Join(gitDir, "some-new-file"))

				repo := NewRepo(gitRemoteUrl, "", gitDir)
				err := repo.CommitAndPush("some-commit-message")
				Expect(err).To(MatchError(MatchRegexp("(?s:failed to push: .*'origin' does not appear to be a git repository)")))
			})
		})
	})

	Describe("Dir", func() {
		It("returns the git directory", func() {
			repo := NewRepo("", "", "some-dir")
			Expect(repo.Dir()).To(Equal("some-dir"))
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

func touchFile(path string) {
	Expect(ioutil.WriteFile(path, nil, 0644)).To(Succeed())
}
