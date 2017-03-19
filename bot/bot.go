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
		if err := c.handleMessage(text, channel, username); err != nil {
			c.logger.WithFields(logrus.Fields{
				"error":    err.Error(),
				"text":     text,
				"channel":  channel,
				"username": username,
			}).Error("Failed to handle message")
		}
	})
}

func (c *bot) handleMessage(text, channel, username string) error {
	if len(strings.Fields(text)) < 2 {
		return errors.New("no command specified")
	}
	command := strings.Fields(text)[1]
	switch command {
	case "claim":
		if err := c.claim(text, channel, username); err != nil {
			return err
		}
	case "help":
		if err := c.help(channel); err != nil {
			return err
		}
	case "owner":
		if err := c.owner(text, channel); err != nil {
			return err
		}
	case "release":
		if err := c.release(text, channel, username); err != nil {
			return err
		}
	case "status":
		if err := c.status(text, channel); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown command '%s'", command)
	}
	return nil
}

func (c *bot) claim(text, channel, username string) error {
	if len(strings.Fields(text)) < 3 {
		return errors.New("no pool specified")
	}
	pool := strings.Fields(text)[2]

	_, unclaimedPools, err := c.locker.Status()
	if err != nil {
		return err
	}
	if !contains(unclaimedPools, pool) {
		if err := c.slackClient.PostMessage(channel, pool+" is not available"); err != nil {
			return err
		}
		return nil
	}

	if err := c.locker.ClaimLock(pool, username); err != nil {
		return err
	}

	if err := c.slackClient.PostMessage(channel, "Claimed "+pool); err != nil {
		return err
	}
	return nil
}

func (c *bot) help(channel string) error {
	helpText := "Available commands:\n" +
		"```\n" +
		"  claim <env>     Claim an unclaimed environment\n" +
		"  owner <env>     Show the user who claimed the environment\n" +
		"  release <env>   Release a claimed environment\n" +
		"  status          Show claimed and unclaimed environments\n" +
		"  help            Display this message\n" +
		"```"
	if err := c.slackClient.PostMessage(channel, helpText); err != nil {
		return err
	}
	return nil
}

func (c *bot) owner(text, channel string) error {
	if len(strings.Fields(text)) < 3 {
		return errors.New("no pool specified")
	}
	pool := strings.Fields(text)[2]

	claimedPools, _, err := c.locker.Status()
	if err != nil {
		return err
	}
	if !contains(claimedPools, pool) {
		if err := c.slackClient.PostMessage(channel, pool+" is not claimed"); err != nil {
			return err
		}
		return nil
	}

	username, date, err := c.locker.Owner(pool)
	if err != nil {
		return err
	}

	if err := c.slackClient.PostMessage(channel, fmt.Sprintf("%s was claimed by %s on %s", pool, username, date)); err != nil {
		return err
	}
	return nil
}

func (c *bot) release(text, channel, username string) error {
	if len(strings.Fields(text)) < 3 {
		return errors.New("no pool specified")
	}
	pool := strings.Fields(text)[2]

	claimedPools, _, err := c.locker.Status()
	if err != nil {
		return err
	}
	if !contains(claimedPools, pool) {
		if err := c.slackClient.PostMessage(channel, pool+" is not claimed"); err != nil {
			return err
		}
		return nil
	}

	if err := c.locker.ReleaseLock(pool, username); err != nil {
		return err
	}

	if err := c.slackClient.PostMessage(channel, "Released "+pool); err != nil {
		return err
	}
	return nil
}

func (c *bot) status(text, channel string) error {
	claimedLocks, unclaimedLocks, err := c.locker.Status()
	if err != nil {
		return err
	}
	statusMessage := fmt.Sprintf(
		"*Claimed:* %s\n*Unclaimed:* %s",
		strings.Join(claimedLocks, ", "),
		strings.Join(unclaimedLocks, ", "),
	)
	if err := c.slackClient.PostMessage(channel, statusMessage); err != nil {
		return err
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, value := range slice {
		if value == item {
			return true
		}
	}
	return false
}
