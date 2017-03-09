package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"flag"
	"golang.org/x/net/websocket"
	"net/url"
)

func main() {
	apiToken := flag.String("apiToken", "", "API Token for Slack")
	flag.String("repoUrl", "", "URL for git repository of locks")
	flag.String("deployKey", "", "Deploy key for Github")
	flag.Parse()

	resp, err := http.Get("https://slack.com/api/rtm.start?token=" + *apiToken)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic("Bad response from slack: " + resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	var rtmStart struct {
		Ok    bool
		Error string
		Url   string
		Self  struct {
			Id string
		}
	}

	if err := json.Unmarshal(body, &rtmStart); err != nil {
		panic(err)
	}
	if !rtmStart.Ok {
		panic("RTM start failed: " + rtmStart.Error)
	}

	botId := rtmStart.Self.Id

	ws, err := websocket.Dial(rtmStart.Url, "", "https://api.slack.com/")
	if err != nil {
		panic(err)
	}

	for {
		var message struct {
			Type    string
			Text    string
			Channel string
		}
		if err := websocket.JSON.Receive(ws, &message); err != nil {
			panic(err)
		}
		if message.Type == "message" && strings.HasPrefix(message.Text, "<@"+botId+">") {
			splitMessage := strings.Fields(message.Text)
			switch splitMessage[1] {
			case "claim":
				_, err := http.PostForm(
					"https://slack.com/api/chat.postMessage",
					url.Values{
						"token":   {*apiToken},
						"channel": {message.Channel},
						"text":    {"Claimed pool-1"},
						"as_user": {"true"},
					},
				)
				if err != nil {
					panic(err)
				}
				// TODO check response
			default:
				panic("Command " + splitMessage[1] + " not implemented")
			}
		}
	}
}
