package commands

import (
	clocker "github.com/mdelillo/claimer/locker"
)

//go:generate counterfeiter . Command
type Command interface {
	Execute() (slackRepsonse string, err error)
}

func poolExists(pool string, locks []clocker.Lock) bool {
	for _, lock := range locks {
		if lock.Name == pool {
			return true
		}
	}
	return false
}

func poolClaimed(pool string, locks []clocker.Lock) bool {
	for _, lock := range locks {
		if lock.Name == pool && lock.Claimed {
			return true
		}
	}
	return false
}
