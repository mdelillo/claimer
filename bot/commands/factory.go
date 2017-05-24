package commands

import clocker "github.com/mdelillo/claimer/locker"

type Factory interface {
	NewCommand(command string, args string, username string) Command
}

//go:generate counterfeiter . locker
type locker interface {
	ClaimLock(pool, username, message string) error
	CreatePool(pool, username string) error
	DestroyPool(pool, username string) error
	ReleaseLock(pool, username string) error
	Status() (locks []clocker.Lock, err error)
}

type commandFactory struct {
	locker locker
}

func NewFactory(locker locker) Factory {
	return &commandFactory{
		locker: locker,
	}
}

func (c *commandFactory) NewCommand(command string, args string, username string) Command {
	switch command {
	case "claim":
		return &claimCommand{
			locker:   c.locker,
			command:  command,
			args:     args,
			username: username,
		}
	case "create":
		return &createCommand{
			locker:   c.locker,
			command:  command,
			args:     args,
			username: username,
		}
	case "destroy":
		return &destroyCommand{
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
			locker:   c.locker,
			command:  command,
			username: username,
		}
	default:
		return &unknownCommand{}
	}
}
