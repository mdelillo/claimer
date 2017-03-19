package main

import (
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	. "github.com/mdelillo/claimer/bot"
	. "github.com/mdelillo/claimer/fs"
	. "github.com/mdelillo/claimer/git"
	. "github.com/mdelillo/claimer/locker"
	. "github.com/mdelillo/claimer/slack"
	. "github.com/mdelillo/claimer/slack/requests"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	apiToken := flag.String("apiToken", "", "API Token for Slack")
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

	if port := os.Getenv("PORT"); port != "" {
		logger.Info("Starting healthcheck listener on port " + port)
		startHealthcheckListener(port)
	}

	fs := NewFs()
	repo := NewRepo(*repoUrl, *deployKey, gitDir)
	locker := NewLocker(fs, repo)
	slackRequestFactory := NewFactory("https://slack.com", *apiToken)
	slackClient := NewClient(slackRequestFactory)
	bot := New(locker, slackClient, logger)
	logger.Info("Claimer starting")
	if err := bot.Run(); err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	logger.Info("Claimer finished")
}

func startHealthcheckListener(port string) {
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "I'm alive")
		})
		http.ListenAndServe("127.0.0.1:"+port, nil)
	}()
}
