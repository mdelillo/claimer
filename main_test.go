package main_test

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

const CHANNEL_ID = "C4DRR335H"

var _ = Describe("Claimer", func() {
	var claimer string
	var gitDir string

	BeforeSuite(func() {
		var err error

		gitDir, err = ioutil.TempDir("", "claimer")
		Expect(err).NotTo(HaveOccurred())

		resetClaimerTestPool(gitDir)

		claimer, err = gexec.Build(filepath.Join("github.com", "mdelillo", "claimer"))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
		resetClaimerTestPool(gitDir)
		os.RemoveAll(gitDir)
	})

	It("claims and releases locks", func() {
		apiToken := getEnv("CLAIMER_TEST_API_TOKEN")
		repoUrl := getEnv("CLAIMER_TEST_REPO_URL")
		deployKey := getEnv("CLAIMER_TEST_DEPLOY_KEY")

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

		claimerCommand := exec.Command(
			claimer,
			"--apiToken", apiToken,
			"--repoUrl", repoUrl,
			"--deployKey", deployKey,
		)
		session, err := gexec.Start(claimerCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		// @claimer status
		// assert about initial status

		postSlackMessage("@claimer claim pool-1", apiToken)
		Expect(latestSlackMessage(apiToken)).To(Equal("Claimed pool-1"))
		// assert about repo
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
	})
})

func getEnv(name string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		Fail(fmt.Sprintf("%s must be set", name))
	}
	return value
}

func postSlackMessage(text, apiToken string) {
	resp, err := http.PostForm(
		"https://slack.com/api/chat.postMessage",
		url.Values{
			"token":    {apiToken},
			"channel":  {CHANNEL_ID},
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

func latestSlackMessage(apiToken string) string {
	resp, err := http.PostForm(
		"https://slack.com/api/channels.history",
		url.Values{
			"token":   {apiToken},
			"channel": {CHANNEL_ID},
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

func resetClaimerTestPool(_ string) {
	// TODO
}
