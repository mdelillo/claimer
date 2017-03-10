package main

import (
	"flag"
	claimerfs "github.com/mdelillo/claimer/fs"
	"github.com/mdelillo/claimer/git"
	"github.com/mdelillo/claimer/slack"
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
			pool := strings.Fields(message.Text)[2]
			switch command {
			case "claim":
				claim(pool, *repoUrl, *deployKey)
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

func claim(pool, repoUrl, deployKey string) {
	fs := claimerfs.NewFs()

	gitDir, err := fs.TempDir("claimer-git-repo")
	if err != nil {
		panic(err)
	}
	defer fs.Rm(gitDir)

	repo := git.NewRepo(repoUrl, deployKey, gitDir)
	if err := repo.Clone(); err != nil {
		panic(err)
	}

	files, err := fs.Ls(filepath.Join(gitDir, pool, "unclaimed"))
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file != ".gitkeep" {
			oldPath := filepath.Join(gitDir, pool, "unclaimed", file)
			newPath := filepath.Join(gitDir, pool, "claimed", file)
			if err := fs.Mv(oldPath, newPath); err != nil {
				panic(err)
			}
			break
		}
	}

	repo.CommitAndPush("Claimer claiming " + pool)
}
