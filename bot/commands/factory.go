package commands

//go:generate counterfeiter . Factory
type Factory interface {
	NewCommand(command string, args []string, username string) Command
}

//go:generate counterfeiter . locker
type locker interface {
	ClaimLock(pool, username string) error
	ReleaseLock(pool, username string) error
	Status() (claimedLocks, unclaimedLocks []string, err error)
	Owner(pool string) (username, date string, err error)
}

type commandFactory struct {
	locker locker
}

func NewFactory(locker locker) Factory {
	return &commandFactory{
		locker: locker,
	}
}

func (c *commandFactory) NewCommand(command string, args []string, username string) Command {
	switch command {
	case "claim":
		return &claimCommand{
			locker:   c.locker,
			command:  command,
			args:     args,
			username: username,
		}
	case "help":
		return &helpCommand{}
	case "owner":
		return &ownerCommand{
			locker:  c.locker,
			command: command,
			args:    args,
		}
	case "release":
		return &releaseCommand{
			locker:   c.locker,
			command:  command,
			args:     args,
			username: username,
		}
	case "status":
		return &statusCommand{
			locker:  c.locker,
			command: command,
		}
	default:
		return &unknownCommand{}
	}
}
