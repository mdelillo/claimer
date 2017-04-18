package commands

import (
	"fmt"
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

	claimedPools, _, err := o.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}
	if !contains(claimedPools, pool) {
		return pool + " is not claimed", nil
	}

	owner, date, message, err := o.locker.Owner(pool)
	if err != nil {
		return "", errors.Wrap(err, "failed to get lock owner")
	}

	response := fmt.Sprintf("%s was claimed by %s on %s", pool, owner, date)
	if message != "" {
		response = fmt.Sprintf("%s (%s)", response, message)
	}
	return response, nil
}
