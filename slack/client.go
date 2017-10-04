package slack

import (
	"encoding/json"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/mdelillo/claimer/slack/requests"
	"github.com/pkg/errors"
	"golang.org/x/net/websocket"
)

type client struct {
	requestFactory requests.Factory
	channelId      string
	logger         *logrus.Logger
}

type rtmEvent struct {
	Type string
}

type message struct {
	Text    string
	Channel string
	User    string
}

func NewClient(requestFactory requests.Factory, channelId string, logger *logrus.Logger) *client {
	return &client{
		requestFactory: requestFactory,
		channelId:      channelId,
		logger:         logger,
	}
}

func (c *client) Listen(messageHandler func(text, channel, username string)) error {
	websocketUrl, botId, err := c.requestFactory.NewStartRtmRequest().Execute()
	if err != nil {
		return errors.Wrap(err, "failed to start RTM")
	}

	if err := c.handleEvents(websocketUrl, botId, messageHandler); err != nil {
		return err
	}

	return nil
}

func (c *client) handleEvents(websocketUrl, botId string, messageHandler func(string, string, string)) error {
	ws, err := websocket.Dial(websocketUrl, "", "https://api.slack.com/")
	if err != nil {
		return errors.Wrap(err, "failed to connect to websocket")
	}

	c.logger.Info("Listening for messages")
	for {
		var data []byte
		if err := websocket.Message.Receive(ws, &data); err != nil {
			return errors.Wrap(err, "failed to receive event")
		}

		if err := c.handleEvent(data, botId, messageHandler); err != nil {
			return err
		}
	}
}

func (c *client) handleEvent(data []byte, botId string, messageHandler func(string, string, string)) error {
	var event *rtmEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return errors.Wrap(err, "failed to parse event")
	}

	if isMessage(event) {
		var message *message
		if err := json.Unmarshal(data, &message); err != nil {
			return errors.Wrap(err, "failed to parse message")
		}

		if inChannel(message, c.channelId) && mentionsBot(message, botId) {
			request := c.requestFactory.NewGetUsernameRequest(message.User)
			username, err := request.Execute()
			if err != nil {
				return errors.Wrap(err, "failed to get username")
			}
			messageHandler(message.Text, message.Channel, username)
		}
	}

	return nil
}

func isMessage(e *rtmEvent) bool {
	return e.Type == "message"
}

func inChannel(message *message, channelId string) bool {
	return message.Channel == channelId
}

func mentionsBot(message *message, botId string) bool {
	return strings.Contains(message.Text, "<@"+botId)
}

func (c *client) PostMessage(channel, message string) error {
	if err := c.requestFactory.NewPostMessageRequest(channel, message).Execute(); err != nil {
		return errors.Wrap(err, "failed to post message")
	}
	return nil
}
