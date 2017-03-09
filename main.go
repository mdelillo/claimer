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
