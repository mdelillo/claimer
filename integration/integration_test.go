package integration_test

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"srcd.works/go-git.v4"
	gitssh "srcd.works/go-git.v4/plumbing/transport/ssh"
	"strings"
	"time"
)

var _ = Describe("Claimer", func() {
	var (
		claimer   string
		gitDir    string
		apiToken  string
		channelId string
		repoUrl   string
		deployKey string
	)

	BeforeSuite(func() {
		var err error

		gitDir, err = ioutil.TempDir("", "claimer-integration-tests")
		Expect(err).NotTo(HaveOccurred())

		claimer, err = gexec.Build(filepath.Join("github.com", "mdelillo", "claimer"))
		Expect(err).NotTo(HaveOccurred())

		apiToken = getEnv("CLAIMER_TEST_API_TOKEN")
		channelId = getEnv("CLAIMER_TEST_CHANNEL_ID")
		repoUrl = getEnv("CLAIMER_TEST_REPO_URL")
		deployKey = getEnv("CLAIMER_TEST_DEPLOY_KEY")
	})

	AfterSuite(func() {
		gexec.KillAndWait()
		gexec.CleanupBuildArtifacts()
		os.RemoveAll(gitDir)
	})

	It("claims and releases locks", func() {
		botId := getEnv("CLAIMER_TEST_BOT_ID")
		userApiToken := getEnv("CLAIMER_TEST_USER_API_TOKEN")
		username := getEnv("CLAIMER_TEST_USERNAME")
		otherChannelId := getEnv("CLAIMER_TEST_OTHER_CHANNEL_ID")

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
		defer resetClaimerTestPool(gitDir, deployKey)

		resetClaimerTestPool(gitDir, deployKey)

		claimerCommand := exec.Command(
			claimer,
			"-apiToken", apiToken,
			"-channelId", channelId,
			"-repoUrl", repoUrl,
			"-deployKey", deployKey,
		)
		session, err := gexec.Start(claimerCommand, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gbytes.Say("Claimer starting"))

		By("Displaying the help message")
		postSlackMessage(fmt.Sprintf("<@%s> help", botId), channelId, userApiToken)
		Eventually(func() string { return latestSlackMessage(channelId, apiToken) }, "10s").
			Should(ContainSubstring("Available commands:"))

		By("Checking the status")
		postSlackMessage(fmt.Sprintf("<@%s> status", botId), channelId, userApiToken)
		Eventually(func() string { return latestSlackMessage(channelId, apiToken) }, "10s").
			Should(Equal("*Claimed:* \n*Unclaimed:* pool-1, pool-2"))

		By("Claiming pool-1")
		postSlackMessage(fmt.Sprintf("<@%s> claim pool-1", botId), channelId, userApiToken)
		Eventually(func() string { return latestSlackMessage(channelId, apiToken) }, "10s").
			Should(Equal("Claimed pool-1"))
		updateGitRepo(gitDir, deployKey)
		Expect(filepath.Join(gitDir, "pool-1", "claimed", "lock-a")).To(BeAnExistingFile())
		Expect(filepath.Join(gitDir, "pool-1", "unclaimed", "lock-a")).NotTo(BeAnExistingFile())

		By("Checking the status")
		postSlackMessage(fmt.Sprintf("<@%s> status", botId), channelId, userApiToken)
		Eventually(func() string { return latestSlackMessage(channelId, apiToken) }, "10s").
			Should(Equal("*Claimed:* pool-1\n*Unclaimed:* pool-2"))

		By("Checking the owner of pool-1")
		postSlackMessage(fmt.Sprintf("<@%s> owner pool-1", botId), channelId, userApiToken)
		ownerMessage := fmt.Sprintf("pool-1 was claimed by %s on ", username)
		Eventually(func() string { return latestSlackMessage(channelId, apiToken) }, "10s").Should(ContainSubstring(ownerMessage))
		date := strings.TrimPrefix(latestSlackMessage(channelId, apiToken), ownerMessage)
		parsedDate, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", date)
		Expect(err).NotTo(HaveOccurred())
		Expect(parsedDate).To(BeTemporally("~", time.Now(), 10*time.Second))

		By("Trying to claim pool-1 again")
		postSlackMessage(fmt.Sprintf("<@%s> claim pool-1", botId), channelId, userApiToken)
		Eventually(func() string { return latestSlackMessage(channelId, apiToken) }, "10s").
			Should(Equal("pool-1 is not available"))

		By("Releasing pool-1")
		postSlackMessage(fmt.Sprintf("<@%s> release pool-1", botId), channelId, userApiToken)
		Eventually(func() string { return latestSlackMessage(channelId, apiToken) }, "10s").
			Should(Equal("Released pool-1"))
		updateGitRepo(gitDir, deployKey)
		Expect(filepath.Join(gitDir, "pool-1", "unclaimed", "lock-a")).To(BeAnExistingFile())
		Expect(filepath.Join(gitDir, "pool-1", "claimed", "lock-a")).NotTo(BeAnExistingFile())

		By("Checking the status")
		postSlackMessage(fmt.Sprintf("<@%s> status", botId), channelId, userApiToken)
		Eventually(func() string { return latestSlackMessage(channelId, apiToken) }, "10s").
			Should(Equal("*Claimed:* \n*Unclaimed:* pool-1, pool-2"))

		By("Checking the status of pool-1")
		postSlackMessage(fmt.Sprintf("<@%s> owner pool-1", botId), channelId, userApiToken)
		Eventually(func() string { return latestSlackMessage(channelId, apiToken) }, "10s").
			Should(Equal("pool-1 is not claimed"))

		By("Trying to release pool-1 again")
		postSlackMessage(fmt.Sprintf("<@%s> release pool-1", botId), channelId, userApiToken)
		Eventually(func() string { return latestSlackMessage(channelId, apiToken) }, "10s").
			Should(Equal("pool-1 is not claimed"))

		By("Trying to claim non-existent pool")
		postSlackMessage(fmt.Sprintf("<@%s> claim non-existent-pool", botId), channelId, userApiToken)
		Eventually(func() string { return latestSlackMessage(channelId, apiToken) }, "10s").
			Should(Equal("non-existent-pool is not available"))

		By("Trying to release non-existent-pool")
		postSlackMessage(fmt.Sprintf("<@%s> release non-existent-pool", botId), channelId, userApiToken)
		Eventually(func() string { return latestSlackMessage(channelId, apiToken) }, "10s").
			Should(Equal("non-existent-pool is not claimed"))

		By("Trying to run an unknown command")
		postSlackMessage(fmt.Sprintf("<@%s> unknown-command", botId), channelId, userApiToken)
		Eventually(func() string { return latestSlackMessage(channelId, apiToken) }, "10s").
			Should(Equal("Unknown command. Try `@claimer help` to see usage."))

		By("Mentioning claimer in a different channel")
		postSlackMessage(fmt.Sprintf("<@%s> help", botId), otherChannelId, userApiToken)
		Consistently(func() string { return latestSlackMessage(otherChannelId, apiToken) }, "10s").
			Should(Equal(fmt.Sprintf("<@%s> help", botId)))
	})

	Context("when $PORT is set", func() {
		var savedPort string

		BeforeEach(func() {
			savedPort = os.Getenv("PORT")
		})

		AfterEach(func() {
			os.Setenv("PORT", savedPort)
		})

		It("responds to get requests on $PORT", func() {
			port := freePort()
			os.Setenv("PORT", port)

			claimerCommand := exec.Command(
				claimer,
				"-apiToken", apiToken,
				"-repoUrl", repoUrl,
				"-deployKey", deployKey,
			)
			session, err := gexec.Start(claimerCommand, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gbytes.Say("Starting healthcheck listener on port " + port))

			Eventually(func() error {
				_, err := net.Dial("tcp", "127.0.0.1:"+port)
				return err
			}).Should(BeNil())

			resp, err := http.Get("http://127.0.0.1:" + port)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(body)).To(Equal("I'm alive"))
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

func postSlackMessage(text, channelId, apiToken string) {
	resp, err := http.PostForm(
		"https://slack.com/api/chat.postMessage",
		url.Values{
			"token":   {apiToken},
			"channel": {channelId},
			"text":    {text},
			"as_user": {"true"},
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

func updateGitRepo(gitDir, deployKey string) {
	runGitCommand(gitDir, deployKey, "fetch")
	runGitCommand(gitDir, deployKey, "reset", "--hard", "origin/master")
}

func resetClaimerTestPool(gitDir, deployKey string) {
	runGitCommand(gitDir, deployKey, "checkout", "master")
	runGitCommand(gitDir, deployKey, "reset", "--hard", "initial-state")
	runGitCommand(gitDir, deployKey, "push", "--force", "origin", "master")
}

func runGitCommand(dir, deployKey string, args ...string) {
	deployKeyDir, err := ioutil.TempDir("", "claimer-integration-test-deploy-key")
	Expect(err).NotTo(HaveOccurred())
	defer os.RemoveAll(deployKeyDir)

	deployKeyPath := filepath.Join(deployKeyDir, "key.pem")
	Expect(ioutil.WriteFile(deployKeyPath, []byte(deployKey), 0600)).To(Succeed())

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), fmt.Sprintf(`GIT_SSH_COMMAND=/usr/bin/ssh -i %s`, deployKeyPath))
	output, err := cmd.CombinedOutput()
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), fmt.Sprintf("Error running git command: %s", string(output)))
}

func freePort() string {
	conn, err := net.Listen("tcp", "127.0.0.1:0")
	Expect(err).NotTo(HaveOccurred())
	defer conn.Close()

	address := strings.Split(conn.Addr().String(), ":")
	return address[1]
}
