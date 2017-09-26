package commands

import (
	. "github.com/mdelillo/claimer/translate"
)

type helpCommand struct{}

func (*helpCommand) Execute() (string, error) {
	message := T("help.header", nil)
	message = message + T("help.body", nil)
	return message, nil
}
