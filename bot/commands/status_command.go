package commands

import (
	"strings"

	clocker "github.com/mdelillo/claimer/locker"
	. "github.com/mdelillo/claimer/translate"
	"github.com/pkg/errors"
)

type statusCommand struct {
	locker   locker
	command  string
	username string
}

func (s *statusCommand) Execute() (string, error) {
	locks, err := s.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}

	usersClaimedLocks := filterLocks(locks, func(lock clocker.Lock) bool {
		return lock.Claimed && lock.Owner == s.username
	})

	otherClaimedLocks := filterLocks(locks, func(lock clocker.Lock) bool {
		return lock.Claimed && lock.Owner != s.username
	})

	unclaimedLocks := filterLocks(locks, func(lock clocker.Lock) bool {
		return !lock.Claimed
	})

	tArgs := TArgs{
		"usersClaimed": lockNames(usersClaimedLocks),
		"otherClaimed": lockNames(otherClaimedLocks),
		"unclaimed":    lockNames(unclaimedLocks),
	}
	return T("status.success", tArgs), nil
}

func filterLocks(locks []clocker.Lock, filterFunc func(clocker.Lock) bool) []clocker.Lock {
	var filteredLocks []clocker.Lock
	for _, lock := range locks {
		if filterFunc(lock) {
			filteredLocks = append(filteredLocks, lock)
		}
	}
	return filteredLocks
}

func lockNames(locks []clocker.Lock) string {
	var names []string
	for _, lock := range locks {
		names = append(names, lock.Name)
	}
	return strings.Join(names, ", ")
}
