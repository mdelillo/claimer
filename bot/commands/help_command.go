package commands

import (
	. "github.com/mdelillo/claimer/translate"
)

type helpCommand struct{}

func (*helpCommand) Execute() (string, error) {
	return T("help", nil), nil
}
