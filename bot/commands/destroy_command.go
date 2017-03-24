package commands

import "github.com/pkg/errors"

type destroyCommand struct {
	locker   locker
	command  string
	args     []string
	username string
}

func (c *destroyCommand) Execute() (string, error) {
	if len(c.args) < 1 {
		return "", errors.New("no pool specified")
	}
	pool := c.args[0]

	claimedPools, unclaimedPools, err := c.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}
	if !contains(claimedPools, pool) && !contains(unclaimedPools, pool) {
		return pool + " does not exist", nil
	}

	if err := c.locker.DestroyPool(pool, c.username); err != nil {
		return "", errors.Wrap(err, "failed to destroy pool")
	}

	return "Destroyed " + pool, nil
}
