package commands

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

type statusCommand struct {
	locker  locker
	command string
}

func (s *statusCommand) Execute() (string, error) {
	claimedLocks, unclaimedLocks, err := s.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}
	return fmt.Sprintf(
			"*Claimed:* %s\n*Unclaimed:* %s",
			strings.Join(claimedLocks, ", "),
			strings.Join(unclaimedLocks, ", "),
		),
		nil
}
