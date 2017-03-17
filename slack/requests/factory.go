package requests

//go:generate counterfeiter . Factory
type Factory interface {
	NewUsernameRequest(userId string) UsernameRequest
}

//go:generate counterfeiter . UsernameRequest
type UsernameRequest interface {
	Execute() (string, error)
}

type requestFactory struct {
	url      string
	apiToken string
}

func NewFactory(url, apiToken string) Factory {
	return &requestFactory{
		url: url,
		apiToken: apiToken,
	}
}

func (r *requestFactory) NewUsernameRequest(userId string) UsernameRequest {
	return &usernameRequest{
		url: r.url,
		apiToken: r.apiToken,
		userId: userId,
	}
}
