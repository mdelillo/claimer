package bot

import (
	"github.com/Sirupsen/logrus"
	"github.com/mdelillo/claimer/bot/commands"
	"strings"
)

//go:generate counterfeiter . commandFactory
type commandFactory interface {
	NewCommand(command string, args []string, username string) commands.Command
}

//go:generate counterfeiter . slackClient
type slackClient interface {
	Listen(messageHandler func(text, channel, username string)) error
	PostMessage(channel, message string) error
}

type bot struct {
	commandFactory commandFactory
	slackClient    slackClient

	logger *logrus.Logger
}

func New(commandFactory commandFactory, slackClient slackClient, logger *logrus.Logger) *bot {
	return &bot{
		commandFactory: commandFactory,
		slackClient:    slackClient,
		logger:         logger,
	}
}

func (c *bot) Run() error {
	return c.slackClient.Listen(func(text, channel, username string) {
		splitText := strings.Fields(text)
		slackResponse, err := c.commandFactory.NewCommand(splitText[1], splitText[2:], username).Execute()
		if err != nil {
			c.logger.WithFields(logrus.Fields{
				"error":    err.Error(),
				"text":     text,
				"channel":  channel,
				"username": username,
			}).Error("failed to execute command")
		}
		if slackResponse != "" {
			if err := c.slackClient.PostMessage(channel, slackResponse); err != nil {
				c.logger.Errorf("failed to post to slack: %s", err)
			}
		}
	})
}
