package bot

import (
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/mdelillo/claimer/bot/commands"
)

//go:generate counterfeiter . commandFactory
type commandFactory interface {
	NewCommand(command string, args string, username string) commands.Command
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
		noPrefix := "<@" + strings.SplitN(text, "<@", 2)[1]
		splitText := strings.SplitN(noPrefix, " ", 3)
		var cmd string
		if len(splitText) > 1 {
			cmd = splitText[1]
		}
		var args string
		if len(splitText) > 2 {
			args = splitText[2]
		}

		c.logger.WithFields(logrus.Fields{
			"command":  cmd,
			"args":     args,
			"username": username,
		}).Debug("Running command")
		slackResponse, err := c.commandFactory.NewCommand(cmd, args, username).Execute()
		if err != nil {
			c.logger.WithFields(logrus.Fields{
				"error":    err.Error(),
				"text":     text,
				"channel":  channel,
				"username": username,
			}).Error("failed to execute command")
		}

		c.logger.WithFields(logrus.Fields{
			"response": slackResponse,
		}).Debug("Received response to command")
		if slackResponse != "" {
			if err := c.slackClient.PostMessage(channel, slackResponse); err != nil {
				c.logger.Errorf("failed to post to slack: %s", err)
			}
		}
	})
}
