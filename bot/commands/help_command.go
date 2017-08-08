package commands

type helpCommand struct{}

func (*helpCommand) Execute() (string, error) {
	return success_help, nil
}
