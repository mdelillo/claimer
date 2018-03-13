package integration_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

const (
	slackRateLimitDelay  = 200 * time.Millisecond
	slackRequestAttempts = 5
)

var _ = Describe("Claimer", func() {
	var (
		apiToken       string
		channelId      string
		repoUrl        string
		deployKey      string
		botId          string
		userApiToken   string
		username       string
		userId         string
		otherChannelId string
		runCommand     func(string) string
		startClaimer   func(string)
		gitDir         string
	)

	BeforeSuite(func() {
		claimer, err := gexec.Build(filepath.Join("github.com", "mdelillo", "claimer"))
		Expect(err).NotTo(HaveOccurred())

		apiToken = getEnv("CLAIMER_TEST_API_TOKEN")
		channelId = getEnv("CLAIMER_TEST_CHANNEL_ID")
		repoUrl = getEnv("CLAIMER_TEST_REPO_URL")
		deployKey = getEnv("CLAIMER_TEST_DEPLOY_KEY")
		botId = getEnv("CLAIMER_TEST_BOT_ID")
		userApiToken = getEnv("CLAIMER_TEST_USER_API_TOKEN")
		username = getEnv("CLAIMER_TEST_USERNAME")
		userId = getEnv("CLAIMER_TEST_USER_ID")
		otherChannelId = getEnv("CLAIMER_TEST_OTHER_CHANNEL_ID")

		runCommand = func(command string) string {
			message := fmt.Sprintf("<@%s> %s", botId, command)
			postSlackMessage(message, channelId, userApiToken)
			EventuallyWithOffset(1, func() string { return latestSlackMessage(channelId, apiToken) }, "20s").
				ShouldNot(Equal(message), fmt.Sprintf(`Did not get response from command "%s"`, command))
			return latestSlackMessage(channelId, apiToken)
		}
		startClaimer = func(translationFile string) {
			args := []string{
				"-apiToken", apiToken,
				"-channelId", channelId,
				"-repoUrl", repoUrl,
				"-deployKey", deployKey,
			}
			if translationFile != "" {
				args = append(args, "-translationFile", translationFile)
			}
			cmd := exec.Command(claimer, args...)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			Eventually(session, "20s").Should(gbytes.Say("Listening for messages"))
		}
	})

	BeforeEach(func() {
		var err error
		gitDir, err = ioutil.TempDir("", "claimer-integration-tests")
		Expect(err).NotTo(HaveOccurred())

		key, err := ssh.NewPublicKeys("git", []byte(deployKey), "")
		Expect(err).NotTo(HaveOccurred())
		_, err = git.PlainClone(gitDir, false, &git.CloneOptions{
			URL:  repoUrl,
			Auth: key,
		})
		Expect(err).NotTo(HaveOccurred())

		resetClaimerTestPool(gitDir, deployKey)
	})

	AfterEach(func() {
		gexec.KillAndWait()
		resetClaimerTestPool(gitDir, deployKey)
		Expect(os.RemoveAll(gitDir)).To(Succeed())
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	It("claims, releases, and shows status of locks", func() {
		startClaimer("")

		Expect(runCommand("help")).To(ContainSubstring("Available commands:"))

		Expect(runCommand("status")).To(Equal("*Claimed by you:* \n*Claimed by others:* pool-3\n*Unclaimed:* pool-1"))

		Expect(runCommand("claim pool-1")).To(Equal("Claimed pool-1"))
		updateGitRepo(gitDir, deployKey)
		Expect(filepath.Join(gitDir, "pool-1", "claimed", "lock-a")).To(BeAnExistingFile())
		Expect(filepath.Join(gitDir, "pool-1", "unclaimed", "lock-a")).NotTo(BeAnExistingFile())

		status := runCommand("status")
		Expect(status).To(ContainSubstring("*Claimed by you:* pool-1\n"))
		Expect(status).NotTo(MatchRegexp(`\*Unclaimed:\*.*pool-1`))

		Expect(runCommand("claim pool-1")).To(Equal("pool-1 is already claimed"))

		Expect(runCommand("release pool-1")).To(Equal("Released pool-1"))
		updateGitRepo(gitDir, deployKey)
		Expect(filepath.Join(gitDir, "pool-1", "unclaimed", "lock-a")).To(BeAnExistingFile())
		Expect(filepath.Join(gitDir, "pool-1", "claimed", "lock-a")).NotTo(BeAnExistingFile())

		status = runCommand("status")
		Expect(status).To(ContainSubstring("*Claimed by you:* \n"))
		Expect(status).To(MatchRegexp(`\*Unclaimed:\*.*pool-1`))

		Expect(runCommand("release pool-1")).To(Equal("pool-1 is not claimed"))

		Expect(runCommand("claim non-existent-pool")).To(Equal("non-existent-pool does not exist"))

		Expect(runCommand("claim")).To(Equal("must specify pool to claim"))

		Expect(runCommand("release non-existent-pool")).To(Equal("non-existent-pool does not exist"))

		Expect(runCommand("release")).To(Equal("must specify pool to release"))

		Expect(runCommand("unknown-command")).To(Equal("Unknown command. Try `@claimer help` to see usage."))
	})

	FIt("claims, releases, and shows status of locks in pools with multiple locks", func() {
		startClaimer("")

		Expect(runCommand("status")).To(MatchRegexp(`\*Unclaimed:\*.*pool-2/lock-a, pool-2/lock-b, pool-2/lock-c`))

		Expect(runCommand("claim pool-2/lock-a")).To(Equal("Claimed pool-2/lock-a"))
		Expect(runCommand("claim pool-2/lock-b")).To(Equal("Claimed pool-2/lock-b"))
		Expect(runCommand("claim pool-2/lock-c")).To(Equal("Claimed pool-2/lock-c"))

		status := runCommand("status")
		Expect(status).To(ContainSubstring("*Claimed by you:* pool-2/lock-a, pool-2/lock-b, pool-2/lock-c\n"))
		Expect(status).NotTo(MatchRegexp(`\*Unclaimed:\*.*pool-2`))

		Expect(runCommand("release pool-2")).To(Equal("Must specify which lock in pool-2 to release"))
		Expect(runCommand("release pool-2/lock-a")).To(Equal("Released pool-2/lock-a"))
		Expect(runCommand("release pool-2/lock-b")).To(Equal("Released pool-2/lock-b"))
		Expect(runCommand("release pool-2/lock-c")).To(Equal("Released pool-2/lock-c"))

		Expect(runCommand("status")).To(MatchRegexp(`\*Unclaimed:\*.*pool-2/lock-a, pool-2/lock-b, pool-2/lock-c`))

		claimedRandomLock := false
		var firstLock string
		for i := 0; i < 20; i++ {
			claimResponse := runCommand("claim pool-2")
			Expect(claimResponse).To(HavePrefix("Claimed pool-2/"))
			lock := strings.TrimPrefix(claimResponse, "Claimed pool-2/")
			runCommand(fmt.Sprintf("release pool-2/%s", lock))
			if firstLock == "" {
				firstLock = lock
			} else {
				if lock != firstLock {
					claimedRandomLock = true
					break
				}
			}
		}
		Expect(claimedRandomLock).To(BeTrue())
	})

	It("shows the owner of a lock", func() {
		startClaimer("")

		Expect(runCommand("owner pool-1")).To(Equal("pool-1 is not claimed"))

		claimTime := time.Now()
		Expect(runCommand("claim pool-1")).To(Equal("Claimed pool-1"))

		owner := runCommand("owner pool-1")
		ownerPrefix := fmt.Sprintf("pool-1 was claimed by %s on ", username)
		Expect(owner).To(HavePrefix(ownerPrefix))

		date := strings.TrimPrefix(owner, ownerPrefix)
		parsedDate, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", date)
		Expect(err).NotTo(HaveOccurred())
		Expect(parsedDate).To(BeTemporally("~", claimTime, 10*time.Second))

		Expect(runCommand("release pool-1")).To(Equal("Released pool-1"))

		Expect(runCommand("claim pool-1 some message")).To(Equal("Claimed pool-1"))

		Expect(runCommand("owner pool-1")).To(HaveSuffix(" (some message)"))

		Expect(runCommand("owner")).To(Equal("must specify pool"))
	})

	It("notifies users who have claimed locks", func() {
		startClaimer("")

		Expect(runCommand("claim pool-1")).To(Equal("Claimed pool-1"))

		notification := runCommand("notify")
		Expect(notification).To(ContainSubstring("Currently claimed locks, please release if not in use:\n"))
		Expect(notification).To(ContainSubstring(fmt.Sprintf("<@%s>: pool-1", userId)))
	})

	It("creates and destroys locks", func() {
		startClaimer("")

		Expect(runCommand("create new-pool")).To(Equal("Created new-pool"))

		updateGitRepo(gitDir, deployKey)
		Expect(filepath.Join(gitDir, "new-pool", "unclaimed", "new-pool")).To(BeAnExistingFile())

		Expect(runCommand("status")).To(MatchRegexp(`\*Unclaimed:\*.*new-pool`))

		Expect(runCommand("create new-pool")).To(Equal("new-pool already exists"))

		Expect(runCommand("destroy new-pool")).To(Equal("Destroyed new-pool"))

		updateGitRepo(gitDir, deployKey)
		Expect(filepath.Join(gitDir, "new-pool")).NotTo(BeADirectory())

		Expect(runCommand("destroy new-pool")).To(Equal("new-pool does not exist"))

		Expect(runCommand("status")).NotTo(MatchRegexp(`\*Unclaimed:\*.*new-pool`))

		Expect(runCommand("create")).To(Equal("must specify name of pool to create"))

		Expect(runCommand("destroy")).To(Equal("must specify pool to destroy"))
	})

	It("does not respond in other channels", func() {
		startClaimer("")

		postSlackMessage(fmt.Sprintf("<@%s> help", botId), otherChannelId, userApiToken)

		Consistently(func() string { return latestSlackMessage(otherChannelId, apiToken) }, "10s").
			Should(Equal(fmt.Sprintf("<@%s> help", botId)))
	})

	Context("when a translation file is provided", func() {
		var translationFilePath string

		BeforeEach(func() {
			translationFile, err := ioutil.TempFile("", "claimer-integration-tests")
			Expect(err).NotTo(HaveOccurred())
			translationFilePath = translationFile.Name()
		})

		AfterEach(func() {
			Expect(os.Remove(translationFilePath)).To(Succeed())
		})

		It("responds with a message from the given translation file", func() {
			translations := `help: {header: "foo\n"}`
			Expect(ioutil.WriteFile(translationFilePath, []byte(translations), 0644)).To(Succeed())

			startClaimer(translationFilePath)

			Expect(runCommand("help")).To(HavePrefix("foo"))

			Expect(runCommand("status")).To(ContainSubstring("Claimed by you:"))
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
	_, err := slackPostForm("https://slack.com/api/chat.postMessage", url.Values{
		"token":   {apiToken},
		"channel": {channelId},
		"text":    {text},
		"as_user": {"true"},
	})
	Expect(err).NotTo(HaveOccurred())
}

func latestSlackMessage(channelId, apiToken string) string {
	body, err := slackPostForm("https://slack.com/api/channels.history", url.Values{
		"token":   {apiToken},
		"channel": {channelId},
		"count":   {"1"},
	})
	Expect(err).NotTo(HaveOccurred())
	var slackResponse struct {
		Messages []struct {
			Text string
		}
	}
	Expect(json.Unmarshal(body, &slackResponse)).To(Succeed())

	return slackResponse.Messages[0].Text
}

func slackPostForm(url string, values url.Values) ([]byte, error) {
	delay := slackRateLimitDelay
	for i := 0; i < slackRequestAttempts; i++ {
		time.Sleep(delay)

		body, err := postForm(url, values)
		if err != nil {
			return nil, err
		}

		var slackResponse struct {
			Ok    bool
			Error string
		}
		if err := json.Unmarshal(body, &slackResponse); err != nil {
			return nil, err
		}

		if slackResponse.Ok {
			return body, nil
		} else if slackResponse.Error != "ratelimited" {
			return nil, fmt.Errorf("Slack request failed: %s", slackResponse.Error)
		}

		delay *= 2
	}
	return nil, fmt.Errorf("Slack request failed %d times", slackRequestAttempts)
}

func postForm(url string, values url.Values) ([]byte, error) {
	response, err := http.PostForm(url, values)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return ioutil.ReadAll(response.Body)
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
