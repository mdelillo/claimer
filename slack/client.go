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

func (c *client) Listen() (<-chan *Message, <-chan error, error) {
	messageChan := make(chan (*Message), 10)
	errorChan := make(chan (error))

	websocketUrl, botId, err := startRtmSession(c.url, c.apiToken)
	if err != nil {
		return nil, nil, err
	}

	if err := listenForMessages(websocketUrl, botId, messageChan, errorChan); err != nil {
		return nil, nil, err
	}

	return messageChan, errorChan, nil
}

func startRtmSession(url, apiToken string) (string, string, error) {
	resp, err := http.Get(url + "/api/rtm.start?token=" + apiToken)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("bad response code: %s", resp.Status)
	}

	var rtmStartResponse struct {
		Ok    bool
		Error string
		Url   string
		Self  struct {
			Id string
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if err := json.Unmarshal(body, &rtmStartResponse); err != nil {
		return "", "", err
	}

	if !rtmStartResponse.Ok {
		return "", "", fmt.Errorf("failed to start RTM session: %s", rtmStartResponse.Error)
	}

	return rtmStartResponse.Url, rtmStartResponse.Self.Id, nil
}

func listenForMessages(url, botId string, messageChan chan<- *Message, errorChan chan<- error) error {
	ws, err := websocket.Dial(url, "", "https://api.slack.com/")
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %s", err)
	}

	go func() {
		for {
			var message *Message
			if err := websocket.JSON.Receive(ws, &message); err != nil {
				close(messageChan)
				errorChan <- fmt.Errorf("failed to parse message: %s", err)
				close(errorChan)
				return
			}

			if messageMentionsUser(message, botId) {
				messageChan <- message
			}
		}
	}()

	return nil
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
