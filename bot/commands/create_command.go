package commands

import (
	"strings"

	. "github.com/mdelillo/claimer/translate"
	"github.com/pkg/errors"
)

type createCommand struct {
	locker   locker
	args     string
	username string
}

func (c *createCommand) Execute() (string, error) {
	args := strings.Fields(c.args)
	if len(c.args) < 1 {
		return T("create.no_pool", nil), nil
	}
	pool := args[0]

	locks, err := c.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}
	if poolExists(pool, locks) {
		return T("create.pool_already_exists", TArgs{"pool": pool}), nil
	}

	if err := c.locker.CreatePool(pool, c.username); err != nil {
		return "", errors.Wrap(err, "failed to create pool")
	}

	return T("create.success", TArgs{"pool": pool}), nil
}
