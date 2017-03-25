package commands

type helpCommand struct{}

func (*helpCommand) Execute() (string, error) {
	return "Available commands:\n" +
			"```\n" +
			"  claim <env>     Claim an unclaimed environment\n" +
			"  create <env>    Create a new environment\n" +
			"  destroy <env>   Destroy an environment\n" +
			"  owner <env>     Show the user who claimed the environment\n" +
			"  release <env>   Release a claimed environment\n" +
			"  status          Show claimed and unclaimed environments\n" +
			"  help            Display this message\n" +
			"```",
		nil
}
