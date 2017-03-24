package commands

import "github.com/pkg/errors"

type createCommand struct {
	locker   locker
	command  string
	args     []string
	username string
}

func (c *createCommand) Execute() (string, error) {
	if len(c.args) < 1 {
		return "", errors.New("no pool specified")
	}
	pool := c.args[0]

	claimedPools, unclaimedPools, err := c.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}
	if contains(claimedPools, pool) || contains(unclaimedPools, pool) {
		return pool + " already exists", nil
	}

	if err := c.locker.CreatePool(pool, c.username); err != nil {
		return "", errors.Wrap(err, "failed to create pool")
	}

	return "Created " + pool, nil
}
