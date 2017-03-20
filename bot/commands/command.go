package commands

//go:generate counterfeiter . Command
type Command interface {
	Execute() (slackRepsonse string, err error)
}

func contains(slice []string, item string) bool {
	for _, value := range slice {
		if value == item {
			return true
		}
	}
	return false
}
