package main

import (
	"strings"

	"flag"
	"github.com/mdelillo/claimer/slack"
)

func main() {
	apiToken := flag.String("apiToken", "", "API Token for Slack")
	flag.String("repoUrl", "", "URL for git repository of locks")
	flag.String("deployKey", "", "Deploy key for Github")
	flag.Parse()

	client := slack.NewClient("https://slack.com", *apiToken)
	messages, err := client.Listen()
	if err != nil {
		panic(err)
	}

	for message := range messages {
		splitMessage := strings.Fields(message.Text)
		switch splitMessage[1] {
		case "claim":
			if err := client.PostMessage(message.Channel, "Claimed pool-1"); err != nil {
				panic(err)
			}
		// TODO check response
		default:
			panic("Command " + splitMessage[1] + " not implemented")
		}
	}
}
