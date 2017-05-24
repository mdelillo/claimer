package commands

import (
	"fmt"
	clocker "github.com/mdelillo/claimer/locker"
	"github.com/pkg/errors"
	"strings"
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
		return pool + " does not exist", nil
	}
	if !poolClaimed(pool, locks) {
		return pool + " is not claimed", nil
	}

	lock := getLock(pool, locks)
	response := fmt.Sprintf("%s was claimed by %s on %s", pool, lock.Owner, lock.Date)
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
