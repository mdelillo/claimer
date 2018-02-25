package commands

import (
	"fmt"
	"strings"

	clocker "github.com/mdelillo/claimer/locker"
	. "github.com/mdelillo/claimer/translate"
	"github.com/pkg/errors"
)

type ownerCommand struct {
	locker locker
	args   string
}

func (o *ownerCommand) Execute() (string, error) {
	args := strings.Fields(o.args)
	if len(args) < 1 {
		return T("owner.no_pool", nil), nil
	}
	pool := args[0]

	locks, err := o.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}
	if !poolExists(pool, locks) {
		return T("owner.pool_does_not_exist", TArgs{"pool": pool}), nil
	}
	if !poolClaimed(pool, locks) {
		return T("owner.pool_is_not_claimed", TArgs{"pool": pool}), nil
	}

	lock := getLock(pool, locks)
	response := T("owner.success", TArgs{"pool": pool, "owner": lock.Owner, "date": lock.Date})
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
