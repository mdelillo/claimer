package commands

import "github.com/pkg/errors"

type claimCommand struct {
	locker   locker
	command  string
	args     []string
	username string
}

func (c *claimCommand) Execute() (string, error) {
	if len(c.args) < 1 {
		return "", errors.New("no pool specified")
	}
	pool := c.args[0]

	_, unclaimedPools, err := c.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}
	if !contains(unclaimedPools, pool) {
		return pool + " is not available", nil
	}

	if err := c.locker.ClaimLock(pool, c.username); err != nil {
		return "", errors.Wrap(err, "failed to claim lock")
	}

	return "Claimed " + pool, nil
}
