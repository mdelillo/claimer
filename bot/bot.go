package bot

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"strings"
)

//go:generate counterfeiter . locker
type locker interface {
	ClaimLock(pool, username string) error
	ReleaseLock(pool, username string) error
	Status() (claimedLocks, unclaimedLocks []string, err error)
	Owner(pool string) (username, date string, err error)
}

//go:generate counterfeiter . slackClient
type slackClient interface {
	Listen(messageHandler func(text, channel, username string)) error
	PostMessage(channel, message string) error
}

type bot struct {
	locker      locker
	slackClient slackClient

	logger *logrus.Logger
}

func New(locker locker, slackClient slackClient, logger *logrus.Logger) *bot {
	return &bot{
		locker:      locker,
		slackClient: slackClient,
		logger:      logger,
	}
}

func (c *bot) Run() error {
	return c.slackClient.Listen(func(text, channel, username string) {
		slackResponse, err := c.handleMessage(text, channel, username)
		if err != nil {
			c.logger.WithFields(logrus.Fields{
				"error":    err.Error(),
				"text":     text,
				"channel":  channel,
				"username": username,
			}).Error("Failed to handle message")
		}
		if slackResponse != "" {
			if err := c.slackClient.PostMessage(channel, slackResponse); err != nil {
				c.logger.Errorf("failed to post to slack: %s", err)
			}
		}
	})
}

func (c *bot) handleMessage(text, channel, username string) (string, error) {
	if len(strings.Fields(text)) < 2 {
		return "No command specified. Try `@claimer help` to see usage.", nil
	}
	command := strings.Fields(text)[1]
	switch command {
	case "claim":
		return c.claim(text, channel, username)
	case "help":
		return c.help(channel)
	case "owner":
		return c.owner(text, channel)
	case "release":
		return c.release(text, channel, username)
	case "status":
		return c.status(text, channel)
	default:
		return "Unknown command. Try `@claimer help` to see usage.", nil
	}
	return "", nil
}

func (c *bot) claim(text, channel, username string) (string, error) {
	if len(strings.Fields(text)) < 3 {
		return "", errors.New("no pool specified")
	}
	pool := strings.Fields(text)[2]

	_, unclaimedPools, err := c.locker.Status()
	if err != nil {
		return "", err
	}
	if !contains(unclaimedPools, pool) {
		return pool + " is not available", nil
	}

	if err := c.locker.ClaimLock(pool, username); err != nil {
		return "", err
	}

	return "Claimed " + pool, nil
}

func (c *bot) help(channel string) (string, error) {
	return "Available commands:\n" +
			"```\n" +
			"  claim <env>     Claim an unclaimed environment\n" +
			"  owner <env>     Show the user who claimed the environment\n" +
			"  release <env>   Release a claimed environment\n" +
			"  status          Show claimed and unclaimed environments\n" +
			"  help            Display this message\n" +
			"```",
		nil
}

func (c *bot) owner(text, channel string) (string, error) {
	if len(strings.Fields(text)) < 3 {
		return "", errors.New("no pool specified")
	}
	pool := strings.Fields(text)[2]

	claimedPools, _, err := c.locker.Status()
	if err != nil {
		return "", err
	}
	if !contains(claimedPools, pool) {
		return pool + " is not claimed", nil
	}

	username, date, err := c.locker.Owner(pool)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s was claimed by %s on %s", pool, username, date), nil
}

func (c *bot) release(text, channel, username string) (string, error) {
	if len(strings.Fields(text)) < 3 {
		return "", errors.New("no pool specified")
	}
	pool := strings.Fields(text)[2]

	claimedPools, _, err := c.locker.Status()
	if err != nil {
		return "", err
	}
	if !contains(claimedPools, pool) {
		return pool + " is not claimed", nil
	}

	if err := c.locker.ReleaseLock(pool, username); err != nil {
		return "", err
	}

	return "Released " + pool, nil
}

func (c *bot) status(text, channel string) (string, error) {
	claimedLocks, unclaimedLocks, err := c.locker.Status()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
			"*Claimed:* %s\n*Unclaimed:* %s",
			strings.Join(claimedLocks, ", "),
			strings.Join(unclaimedLocks, ", "),
		),
		nil
}

func contains(slice []string, item string) bool {
	for _, value := range slice {
		if value == item {
			return true
		}
	}
	return false
}
