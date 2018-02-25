package commands

import (
	"strings"

	. "github.com/mdelillo/claimer/translate"
	"github.com/pkg/errors"
)

type destroyCommand struct {
	locker   locker
	args     string
	username string
}

func (c *destroyCommand) Execute() (string, error) {
	args := strings.Fields(c.args)
	if len(args) < 1 {
		return T("destroy.no_pool", nil), nil
	}
	pool := args[0]

	locks, err := c.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}
	if !poolExists(pool, locks) {
		return T("destroy.pool_does_not_exist", TArgs{"pool": pool}), nil
	}

	if err := c.locker.DestroyPool(pool, c.username); err != nil {
		return "", errors.Wrap(err, "failed to destroy pool")
	}

	return T("destroy.success", TArgs{"pool": pool}), nil
}
