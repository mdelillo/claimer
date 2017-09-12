package commands

import (
	. "github.com/mdelillo/claimer/translate"
)

type unknownCommand struct{}

func (*unknownCommand) Execute() (string, error) {
	return T("unknown_command", nil), nil
}
