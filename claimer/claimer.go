package claimer

import (
	"errors"
	"fmt"
	"strings"
)

//go:generate counterfeiter . locker
type locker interface {
	ClaimLock(pool string) error
	ReleaseLock(pool string) error
}

//go:generate counterfeiter . slackClient
type slackClient interface {
	Listen(messageHandler func(text, channel string)) error
	PostMessage(channel, message string) error
}

type claimer struct {
	locker      locker
	slackClient slackClient
}

func New(locker locker, slackClient slackClient) *claimer {
	return &claimer{
		locker:      locker,
		slackClient: slackClient,
	}
}

func (c *claimer) Run() error {
	var messageHandlingErr error
	err := c.slackClient.Listen(func(text, channel string) {
		messageHandlingErr = c.handleMessage(text, channel)
	})
	if err != nil {
		return err
	}
	return messageHandlingErr
}

func (c *claimer) handleMessage(text, channel string) error {
	if len(strings.Fields(text)) < 2 {
		return errors.New("no command specified")
	}
	command := strings.Fields(text)[1]
	switch command {
	case "claim":
		if err := c.claim(text, channel); err != nil {
			return err
		}
	case "release":
		if err := c.release(text, channel); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown command '%s'", command)
	}
	return nil
}

func (c *claimer) claim(text, channel string) error {
	if len(strings.Fields(text)) < 3 {
		return errors.New("no pool specified")
	}
	pool := strings.Fields(text)[2]
	if err := c.locker.ClaimLock(pool); err != nil {
		return err
	}
	if err := c.slackClient.PostMessage(channel, "Claimed "+pool); err != nil {
		return err
	}
	return nil
}

func (c *claimer) release(text, channel string) error {
	if len(strings.Fields(text)) < 3 {
		return errors.New("no pool specified")
	}
	pool := strings.Fields(text)[2]
	if err := c.locker.ReleaseLock(pool); err != nil {
		return err
	}
	if err := c.slackClient.PostMessage(channel, "Released "+pool); err != nil {
		return err
	}
	return nil
}
