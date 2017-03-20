package commands

import "github.com/pkg/errors"

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
		return "", errors.Wrap(err, "failed to get status of locks")
	}
	if !contains(claimedPools, pool) {
		return pool + " is not claimed", nil
	}

	if err := r.locker.ReleaseLock(pool, r.username); err != nil {
		return "", errors.Wrap(err, "failed to release lock")
	}

	return "Released " + pool, nil
}
