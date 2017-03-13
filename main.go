package main

import (
	"flag"
	. "github.com/mdelillo/claimer/fs"
	. "github.com/mdelillo/claimer/git"
	. "github.com/mdelillo/claimer/locker"
	. "github.com/mdelillo/claimer/slack"
	"strings"
)

func main() {
	apiToken := flag.String("apiToken", "", "API Token for Slack")
	repoUrl := flag.String("repoUrl", "", "URL for git repository of locks")
	deployKey := flag.String("deployKey", "", "Deploy key for Github")
	flag.Parse()

	fs := NewFs()

	gitDir, err := fs.TempDir("claimer-git-repo")
	if err != nil {
		panic(err)
	}
	defer fs.Rm(gitDir)

	repo := NewRepo(*repoUrl, *deployKey, gitDir)

	locker := NewLocker(fs, repo)

	client := NewClient("https://slack.com", *apiToken)
	messageChan, errorChan, err := client.Listen()
	if err != nil {
		panic(err)
	}

	for {
		select {
		case message := <-messageChan:
			command := strings.Fields(message.Text)[1]
			pool := strings.Fields(message.Text)[2]
			switch command {
			case "claim":
				if err := locker.ClaimLock(pool); err != nil {
					panic(err)
				}
				if err := client.PostMessage(message.Channel, "Claimed "+pool); err != nil {
					panic(err)
				}
			default:
				panic("Command " + command + " not implemented")
			}
		case err := <-errorChan:
			panic(err)
		}
	}
}
