package commands

import (
	"fmt"
	"strings"
)

type statusCommand struct {
	locker  locker
	command string
}

func (s *statusCommand) Execute() (string, error) {
	claimedLocks, unclaimedLocks, err := s.locker.Status()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
			"*Claimed:* %s\n*Unclaimed:* %s",
			strings.Join(claimedLocks, ", "),
			strings.Join(unclaimedLocks, ", "),
		),
		nil
}
