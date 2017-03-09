package main

import (
	"strings"

	"flag"
	"github.com/mdelillo/claimer/slack"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"srcd.works/go-git.v4"
	gitssh "srcd.works/go-git.v4/plumbing/transport/ssh"
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
	gitDir, err := ioutil.TempDir("", "claimer")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(gitDir)

	signer, err := ssh.ParsePrivateKey([]byte(deployKey))
	if err != nil {
		panic(err)
	}

	_, err = git.PlainClone(gitDir, false, &git.CloneOptions{
		URL: repoUrl,
		Auth: &gitssh.PublicKeys{
			User:   "git",
			Signer: signer,
		},
	})
	if err != nil {
		panic(err)
	}

	files, err := ioutil.ReadDir(filepath.Join(gitDir, resource, "unclaimed"))
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.Name() != ".gitkeep" {
			oldPath := filepath.Join(gitDir, resource, "unclaimed", file.Name())
			newPath := filepath.Join(gitDir, resource, "claimed", file.Name())
			if err := os.Rename(oldPath, newPath); err != nil {
				panic(err)
			}
			break
		}
	}

	runGitCommand(gitDir, "add", "-A", ".")
	runGitCommand(gitDir, "commit", "-m", "Claimer claiming resource "+resource)
	runGitCommand(gitDir, "push", "origin", "master")
}

func runGitCommand(gitDir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = gitDir
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
