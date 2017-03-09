package main

import (
	"flag"
	"github.com/mdelillo/claimer/git"
	"github.com/mdelillo/claimer/slack"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	apiToken := flag.String("apiToken", "", "API Token for Slack")
	repoUrl := flag.String("repoUrl", "", "URL for git repository of locks")
	deployKey := flag.String("deployKey", "", "Deploy key for Github")
	flag.Parse()

	client := slack.NewClient("https://slack.com", *apiToken)
	messageChan, errorChan, err := client.Listen()
	if err != nil {
		panic(err)
	}

	for {
		select {
		case message := <-messageChan:
			command := strings.Fields(message.Text)[1]
			resource := strings.Fields(message.Text)[2]
			switch command {
			case "claim":
				claim(resource, *repoUrl, *deployKey)
				if err := client.PostMessage(message.Channel, "Claimed "+resource); err != nil {
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

func claim(resource, repoUrl, deployKey string) {
	repo, err := git.NewRepo(repoUrl, deployKey)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(repo.Dir)

	files, err := ioutil.ReadDir(filepath.Join(repo.Dir, resource, "unclaimed"))
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.Name() != ".gitkeep" {
			oldPath := filepath.Join(repo.Dir, resource, "unclaimed", file.Name())
			newPath := filepath.Join(repo.Dir, resource, "claimed", file.Name())
			if err := os.Rename(oldPath, newPath); err != nil {
				panic(err)
			}
			break
		}
	}

	repo.CommitAndPush("Claimer claiming resource " + resource)
}
