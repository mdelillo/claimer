package commands

import (
	"github.com/pkg/errors"
	"strings"
)

type createCommand struct {
	locker   locker
	command  string
	args     string
	username string
}

func (c *createCommand) Execute() (string, error) {
	args := strings.Fields(c.args)
	if len(c.args) < 1 {
		return "", errors.New("no pool specified")
	}
	pool := args[0]

	locks, err := c.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}
	if poolExists(pool, locks) {
		return pool + " already exists", nil
	}

	if err := c.locker.CreatePool(pool, c.username); err != nil {
		return "", errors.Wrap(err, "failed to create pool")
	}

	return "Created " + pool, nil
}
