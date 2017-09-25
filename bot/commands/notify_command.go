package commands

import (
	"fmt"
	"strings"

	. "github.com/mdelillo/claimer/translate"
	"github.com/pkg/errors"
)

type notifyCommand struct {
	locker locker
}

func (n *notifyCommand) Execute() (string, error) {
	ownerStatus := make(map[string][]string)
	locks, err := n.locker.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get status of locks")
	}

	for _, l := range locks {
		if l.Claimed {
			owned, ok := ownerStatus[l.Owner]
			if ok {
				ownerStatus[l.Owner] = append(owned, l.Name)
			} else {
				ownerStatus[l.Owner] = []string{l.Name}
			}
		}
	}

	if len(ownerStatus) == 0 {
		return "No locks currently claimed.", nil
	}

	mentions := ""
	for owner, ls := range ownerStatus {
		mentions = mentions + fmt.Sprintf("<@%s>: %s\n", owner, strings.Join(ls, ", "))
	}
	mentions = strings.TrimSpace(mentions)

	return T("notify.success", TArgs{"mentions": mentions}), nil
}
