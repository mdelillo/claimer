package commands

type unknownCommand struct{}

func (*unknownCommand) Execute() (string, error) {
	return "Unknown command. Try `@claimer help` to see usage.", nil
}
