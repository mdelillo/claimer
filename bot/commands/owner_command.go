package commands

import (
	"fmt"
	"strings"

	clocker "github.com/mdelillo/claimer/locker"
	"github.com/pkg/errors"
)

type ownerCommand struct {
	locker  locker
	command string
	args    string
}

func (o *ownerCommand) Execute() (string, error) {
	args := strings.Fields(o.args)
	if len(args) < 1 {
		return "", errors.New("no pool specified")
	}
	pool := args[0]

	locks, err := o.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}
	if !poolExists(pool, locks) {
		return fmt.Sprintf(pool_does_not_exist_owner, pool), nil
	}
	if !poolClaimed(pool, locks) {
		return fmt.Sprintf(pool_is_not_claimed_owner, pool), nil
	}

	lock := getLock(pool, locks)
	response := fmt.Sprintf(success_owner, pool, lock.Owner, lock.Date)
	if lock.Message != "" {
		response = fmt.Sprintf("%s (%s)", response, lock.Message)
	}
	return response, nil
}

func getLock(pool string, locks []clocker.Lock) *clocker.Lock {
	for _, lock := range locks {
		if lock.Name == pool {
			return &lock
		}
	}
	return nil
}
