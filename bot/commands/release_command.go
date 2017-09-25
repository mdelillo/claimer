package commands

import (
	"strings"

	. "github.com/mdelillo/claimer/translate"
	"github.com/pkg/errors"
)

type releaseCommand struct {
	locker   locker
	args     string
	username string
}

func (r *releaseCommand) Execute() (string, error) {
	args := strings.Fields(r.args)
	if len(r.args) < 1 {
		return "", errors.New("no pool specified")
	}
	pool := args[0]

	locks, err := r.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}
	if !poolExists(pool, locks) {
		return T("release.pool_does_not_exist", TArgs{"pool": pool}), nil
	}
	if !poolClaimed(pool, locks) {
		return T("release.pool_is_not_claimed", TArgs{"pool": pool}), nil
	}

	if err := r.locker.ReleaseLock(pool, r.username); err != nil {
		return "", errors.Wrap(err, "failed to release lock")
	}

	return T("release.success", TArgs{"pool": pool}), nil
}
