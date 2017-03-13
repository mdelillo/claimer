package integration_test

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"srcd.works/go-git.v4"
	gitssh "srcd.works/go-git.v4/plumbing/transport/ssh"
)

var _ = Describe("Claimer", func() {
	var claimer string
	var gitDir string

	BeforeSuite(func() {
		var err error

		gitDir, err = ioutil.TempDir("", "claimer-integration-tests")
		Expect(err).NotTo(HaveOccurred())

		claimer, err = gexec.Build(filepath.Join("github.com", "mdelillo", "claimer"))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
		os.RemoveAll(gitDir)
	})

	It("claims and releases locks", func() {
		apiToken := getEnv("CLAIMER_TEST_API_TOKEN")
		repoUrl := getEnv("CLAIMER_TEST_REPO_URL")
		deployKey := getEnv("CLAIMER_TEST_DEPLOY_KEY")
		channelId := getEnv("CLAIMER_TEST_CHANNEL_ID")
		botId := getEnv("CLAIMER_TEST_BOT_ID")

		signer, err := ssh.ParsePrivateKey([]byte(deployKey))
		Expect(err).NotTo(HaveOccurred())

		_, err = git.PlainClone(gitDir, false, &git.CloneOptions{
			URL: repoUrl,
			Auth: &gitssh.PublicKeys{
				User:   "git",
				Signer: signer,
			},
		})
		Expect(err).NotTo(HaveOccurred())

		resetClaimerTestPool(gitDir)

		claimerCommand := exec.Command(
			claimer,
			"-apiToken", apiToken,
			"-repoUrl", repoUrl,
			"-deployKey", deployKey,
		)
		session, err := gexec.Start(claimerCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		// @claimer status
		// assert about initial status

		By("Claiming a pool")
		postSlackMessage(fmt.Sprintf("<@%s> claim pool-1", botId), channelId, apiToken)
		Eventually(func() string { return latestSlackMessage(channelId, apiToken) }, "10s").
			Should(Equal("Claimed pool-1"))
		updateGitRepo(gitDir)
		Expect(filepath.Join(gitDir, "pool-1", "claimed", "lock-a")).To(BeARegularFile())
		Expect(filepath.Join(gitDir, "pool-1", "unclaimed", "lock-a")).NotTo(BeAnExistingFile())

		// @claimer status
		// assert about status

		// @claimer claim pool-1
		// assert about error

		// @claimer release pool-1
		// assert about response
		// assert about repo
		// @claimer status
		// assert about status

		// @claimer release pool-1
		// assert about response

		// @claimer claim non-existent-pool
		// assert about error

		// @claimer release non-existent-pool
		// assert about error

		// @claimer claim pool-2
		// assert about error
		// assert about repo
		// @claimer status
		// assert about status

		session.Terminate().Wait()

		resetClaimerTestPool(gitDir)
	})

})

func getEnv(name string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		Fail(fmt.Sprintf("%s must be set", name))
	}
	return value
}

func postSlackMessage(text, channelId, apiToken string) {
	resp, err := http.PostForm(
		"https://slack.com/api/chat.postMessage",
		url.Values{
			"token":    {apiToken},
			"channel":  {channelId},
			"text":     {text},
			"as_user":  {"false"},
			"username": {"claimer-integration-test"},
		},
	)
	Expect(err).NotTo(HaveOccurred())
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())

	var slackResponse struct {
		Ok    bool
		Error string
	}
	Expect(json.Unmarshal(body, &slackResponse)).To(Succeed())

	Expect(slackResponse.Ok).To(BeTrue(), fmt.Sprintf("Posting to slack failed: %s", slackResponse.Error))
}

func latestSlackMessage(channelId, apiToken string) string {
	resp, err := http.PostForm(
		"https://slack.com/api/channels.history",
		url.Values{
			"token":   {apiToken},
			"channel": {channelId},
			"count":   {"1"},
		},
	)
	Expect(err).NotTo(HaveOccurred())
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())

	var slackResponse struct {
		Ok       bool
		Messages []struct {
			Text string
		}
		Error string
	}
	Expect(json.Unmarshal(body, &slackResponse)).To(Succeed())

	Expect(slackResponse.Ok).To(BeTrue(), fmt.Sprintf("Getting message from slack failed: %s", slackResponse.Error))

	return slackResponse.Messages[0].Text
}

func updateGitRepo(gitDir string) {
	runGitCommand(gitDir, "fetch")
	runGitCommand(gitDir, "reset", "--hard", "origin/master")
}

func resetClaimerTestPool(gitDir string) {
	runGitCommand(gitDir, "checkout", "master")
	runGitCommand(gitDir, "reset", "--hard", "initial-state")
	runGitCommand(gitDir, "push", "--force", "origin", "master")

}

func runGitCommand(dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	ExpectWithOffset(1, cmd.Run()).To(Succeed())
}
