package commands

import (
	"github.com/pkg/errors"
	"strings"
)

type claimCommand struct {
	locker   locker
	command  string
	args     string
	username string
}

func (c *claimCommand) Execute() (string, error) {
	args := strings.SplitN(c.args, " ", 2)
	if len(c.args) < 1 {
		return "", errors.New("no pool specified")
	}
	pool := args[0]

	locks, err := c.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}
	if !poolExists(pool, locks) {
		return pool + " does not exist", nil
	}
	if poolClaimed(pool, locks) {
		return pool + " is already claimed", nil
	}

	var message string
	if len(args) > 1 {
		message = args[1]
	}
	if err := c.locker.ClaimLock(pool, c.username, message); err != nil {
		return "", errors.Wrap(err, "failed to claim lock")
	}

	return "Claimed " + pool, nil
}
