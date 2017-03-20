package commands

import (
	"errors"
)

type releaseCommand struct {
	locker   locker
	command  string
	args     []string
	username string
}

func (r *releaseCommand) Execute() (string, error) {
	if len(r.args) < 1 {
		return "", errors.New("no pool specified")
	}
	pool := r.args[0]

	claimedPools, _, err := r.locker.Status()
	if err != nil {
		return "", err
	}
	if !contains(claimedPools, pool) {
		return pool + " is not claimed", nil
	}

	if err := r.locker.ReleaseLock(pool, r.username); err != nil {
		return "", err
	}

	return "Released " + pool, nil
}
