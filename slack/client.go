package slack

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type client struct {
	apiToken string
	url      string
}

func NewClient(url, apiToken string) *client {
	return &client{
		apiToken: apiToken,
		url:      url,
	}
}

type Message struct {
	Type    string
	Text    string
	Channel string
}

func (c *client) Listen() (<-chan *Message, error) {
	resp, err := http.Get(c.url + "/api/rtm.start?token=" + c.apiToken)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic("error from slack:: " + resp.Status)
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
		panic("error from slack: " + rtmStart.Error)
	}

	botId := rtmStart.Self.Id

	ws, err := websocket.Dial(rtmStart.Url, "", "https://api.slack.com/")
	if err != nil {
		panic(err)
	}

	messageChan := make(chan (*Message), 10)

	go func() {
		for {
			var message *Message
			if err := websocket.JSON.Receive(ws, &message); err != nil {
				// TODO: close message channel and put err on an error channel
				return
			}

			if messageMentionsUser(message, botId) {
				messageChan <- message
			}
		}
	}()

	return messageChan, nil
}

func messageMentionsUser(message *Message, botId string) bool {
	return message.Type == "message" && strings.HasPrefix(message.Text, "<@"+botId+">")
}

func (c *client) PostMessage(channel, message string) error {
	resp, err := http.PostForm(
		fmt.Sprintf("%s/api/chat.postMessage", c.url),
		url.Values{
			"token":   {c.apiToken},
			"channel": {channel},
			"text":    {message},
			"as_user": {"true"},
		},
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("error posting to slack: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResponse struct {
		Ok    bool
		Error string
	}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return err
	}

	if !apiResponse.Ok {
		return fmt.Errorf("error posting to slack: %s", apiResponse.Error)
	}
	return nil
}
