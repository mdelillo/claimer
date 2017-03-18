package requests

//go:generate counterfeiter . Factory
type Factory interface {
	NewGetUsernameRequest(userId string) GetUsernameRequest
}

//go:generate counterfeiter . GetUsernameRequest
type GetUsernameRequest interface {
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

func (r *requestFactory) NewGetUsernameRequest(userId string) GetUsernameRequest {
	return &getUsernameRequest{
		url: r.url,
		apiToken: r.apiToken,
		userId: userId,
	}
}
