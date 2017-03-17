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
	apiToken       string
	url            string
	requestFactory RequestFactory
}

type rtmEvent struct {
	Type string
}

type message struct {
	Text     string
	Channel  string
	User string
}

func NewClient(url, apiToken string, requestFactory RequestFactory) *client {
	return &client{
		apiToken:       apiToken,
		url:            url,
		requestFactory: requestFactory,
	}
}

func (c *client) Listen(messageHandler func(text, channel, username string)) error {
	websocketUrl, botId, err := c.startRtmSession()
	if err != nil {
		return err
	}

	if err := c.handleEvents(websocketUrl, botId, messageHandler); err != nil {
		return err
	}

	return nil
}

func (c *client) startRtmSession() (string, string, error) {
	resp, err := http.Get(c.url + "/api/rtm.start?token=" + c.apiToken)
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

func (c *client) handleEvents(websocketUrl, botId string, messageHandler func(string, string, string)) error {
	ws, err := websocket.Dial(websocketUrl, "", "https://api.slack.com/")
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %s", err)
	}

	for {
		var data []byte
		if err := websocket.Message.Receive(ws, &data); err != nil {
			return fmt.Errorf("failed to receive event: %s", err)
		}

		if err := c.handleEvent(data, botId, messageHandler); err != nil {
			return err
		}
	}

	return nil
}

func (c *client) handleEvent(data []byte, botId string, messageHandler func(string, string, string)) error {
	var event *rtmEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to parse event: %s", err)
	}

	if isMessage(event) {
		var message *message
		if err := json.Unmarshal(data, &message); err != nil {
			return fmt.Errorf("failed to parse message: %s", err)
		}

		if mentionsBot(message, botId) {
			request := c.requestFactory.NewUsernameRequest(message.User)
			username, err := request.Execute()
			if err != nil {
				return fmt.Errorf("failed to get username for %s: %s", message.User, err)
			}
			messageHandler(message.Text, message.Channel, username)
		}
	}

	return nil
}

func isMessage(e *rtmEvent) bool {
	return e.Type == "message"
}

func mentionsBot(message *message, botId string) bool {
	return strings.HasPrefix(message.Text, "<@"+botId+">")
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
