package main

import (
	"flag"
	"fmt"
	. "github.com/mdelillo/claimer/claimer"
	. "github.com/mdelillo/claimer/fs"
	. "github.com/mdelillo/claimer/git"
	. "github.com/mdelillo/claimer/locker"
	. "github.com/mdelillo/claimer/slack"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	apiToken := flag.String("apiToken", "", "API Token for Slack")
	repoUrl := flag.String("repoUrl", "", "URL for git repository of locks")
	deployKey := flag.String("deployKey", "", "Deploy key for Github")
	flag.Parse()

	gitDir, err := ioutil.TempDir("", "claimer-git-repo")
	if err != nil {
		fmt.Printf("Error creating temp directory: %s\n", err)
	}
	defer os.RemoveAll(gitDir)

	if port := os.Getenv("PORT"); port != "" {
		startHealthcheckListener(port)
	}

	fs := NewFs()
	repo := NewRepo(*repoUrl, *deployKey, gitDir)
	locker := NewLocker(fs, repo)
	slackClient := NewClient("https://slack.com", *apiToken)
	claimer := New(locker, slackClient)
	if err := claimer.Run(); err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Println("All done: " + err.Error())
}

func startHealthcheckListener(port string) {
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "I'm alive")
		})
		http.ListenAndServe("127.0.0.1:"+port, nil)
	}()
}
