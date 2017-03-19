package slack

import (
	"encoding/json"
	"fmt"
	"github.com/mdelillo/claimer/slack/requests"
	"golang.org/x/net/websocket"
	"strings"
)

type client struct {
	requestFactory requests.Factory
}

type rtmEvent struct {
	Type string
}

type message struct {
	Text    string
	Channel string
	User    string
}

func NewClient(requestFactory requests.Factory) *client {
	return &client{requestFactory: requestFactory}
}

func (c *client) Listen(messageHandler func(text, channel, username string)) error {
	websocketUrl, botId, err := c.requestFactory.NewStartRtmRequest().Execute()
	if err != nil {
		return err
	}

	if err := c.handleEvents(websocketUrl, botId, messageHandler); err != nil {
		return err
	}

	return nil
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
			request := c.requestFactory.NewGetUsernameRequest(message.User)
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
	return c.requestFactory.NewPostMessageRequest(channel, message).Execute()
}
