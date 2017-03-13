package main

import (
	"flag"
	. "github.com/mdelillo/claimer/fs"
	. "github.com/mdelillo/claimer/git"
	. "github.com/mdelillo/claimer/locker"
	. "github.com/mdelillo/claimer/slack"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	apiToken := flag.String("apiToken", "", "API Token for Slack")
	repoUrl := flag.String("repoUrl", "", "URL for git repository of locks")
	deployKey := flag.String("deployKey", "", "Deploy key for Github")
	flag.Parse()

	fs := NewFs()

	gitDir, err := ioutil.TempDir("", "claimer-git-repo")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(gitDir)

	repo := NewRepo(*repoUrl, *deployKey, gitDir)

	locker := NewLocker(fs, repo)

	client := NewClient("https://slack.com", *apiToken)
	err = client.Listen(func(text, channel string) {
		command := strings.Fields(text)[1]
		pool := strings.Fields(text)[2]
		switch command {
		case "claim":
			if err := locker.ClaimLock(pool); err != nil {
				panic(err)
			}
			if err := client.PostMessage(channel, "Claimed "+pool); err != nil {
				panic(err)
			}
		default:
			panic("Command " + command + " not implemented")
		}
	})
	if err != nil {
		panic(err)
	}
}
