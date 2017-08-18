package commands

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type destroyCommand struct {
	locker   locker
	command  string
	args     string
	username string
}

func (c *destroyCommand) Execute() (string, error) {
	args := strings.Fields(c.args)
	if len(args) < 1 {
		return "", errors.New(pool_not_specified_destroy)
	}
	pool := args[0]

	locks, err := c.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}
	if !poolExists(pool, locks) {
		return fmt.Sprintf(pool_does_not_exist_destroy, pool), nil
	}

	if err := c.locker.DestroyPool(pool, c.username); err != nil {
		return "", errors.Wrap(err, "failed to destroy pool")
	}

	return fmt.Sprintf(success_destroy, pool), nil
}
