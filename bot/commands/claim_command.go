package commands

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
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
		return "", errors.New(pool_not_specified_claim)
	}
	pool := args[0]

	locks, err := c.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}
	if !poolExists(pool, locks) {
		return fmt.Sprintf(pool_does_not_exist_claim, pool), nil
	}
	if poolClaimed(pool, locks) {
		return fmt.Sprintf(pool_already_claimed_claim, pool), nil
	}

	var message string
	if len(args) > 1 {
		message = args[1]
	}
	if err := c.locker.ClaimLock(pool, c.username, message); err != nil {
		return "", errors.Wrap(err, "failed to claim lock")
	}

	return fmt.Sprintf(success_claim, pool), nil
}
