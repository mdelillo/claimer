package commands

import (
	"errors"
	"fmt"
)

type ownerCommand struct {
	locker  locker
	command string
	args    []string
}

func (o *ownerCommand) Execute() (string, error) {
	if len(o.args) < 1 {
		return "", errors.New("no pool specified")
	}
	pool := o.args[0]

	claimedPools, _, err := o.locker.Status()
	if err != nil {
		return "", err
	}
	if !contains(claimedPools, pool) {
		return pool + " is not claimed", nil
	}

	owner, date, err := o.locker.Owner(pool)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s was claimed by %s on %s", pool, owner, date), nil
}
