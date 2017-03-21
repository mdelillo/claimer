package main

import (
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/mdelillo/claimer/bot"
	"github.com/mdelillo/claimer/bot/commands"
	"github.com/mdelillo/claimer/fs"
	"github.com/mdelillo/claimer/git"
	"github.com/mdelillo/claimer/locker"
	"github.com/mdelillo/claimer/slack"
	"github.com/mdelillo/claimer/slack/requests"
	"io/ioutil"
	"os"
)

func main() {
	apiToken := flag.String("apiToken", "", "API Token for Slack")
	channelId := flag.String("channelId", "", "ID of slack channel to listen in")
	repoUrl := flag.String("repoUrl", "", "URL for git repository of locks")
	deployKey := flag.String("deployKey", "", "Deploy key for Github")
	flag.Parse()

	logger := logrus.New()
	logger.Out = os.Stdout
	logger.Formatter = &logrus.TextFormatter{FullTimestamp: true}

	gitDir, err := ioutil.TempDir("", "claimer-git-repo")
	if err != nil {
		fmt.Printf("Error creating temp directory: %s\n", err)
	}
	defer os.RemoveAll(gitDir)

	claimer := bot.New(
		commands.NewFactory(
			locker.NewLocker(
				fs.NewFs(),
				git.NewRepo(*repoUrl, *deployKey, gitDir),
			),
		),
		slack.NewClient(
			requests.NewFactory("https://slack.com", *apiToken),
			*channelId,
		),
		logger,
	)

	logger.Info("Claimer starting")
	if err := claimer.Run(); err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	logger.Info("Claimer finished")
}
