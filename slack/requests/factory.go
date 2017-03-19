package requests

//go:generate counterfeiter . Factory
type Factory interface {
	NewGetUsernameRequest(userId string) GetUsernameRequest
	NewPostMessageRequest(channel, message string) PostMessageRequest
	NewStartRtmRequest() StartRtmRequest
}

//go:generate counterfeiter . GetUsernameRequest
type GetUsernameRequest interface {
	Execute() (username string, err error)
}

//go:generate counterfeiter . PostMessageRequest
type PostMessageRequest interface {
	Execute() error
}

//go:generate counterfeiter . StartRtmRequest
type StartRtmRequest interface {
	Execute() (websocketUrl, botId string, err error)
}

type requestFactory struct {
	url      string
	apiToken string
}

func NewFactory(url, apiToken string) Factory {
	return &requestFactory{
		url:      url,
		apiToken: apiToken,
	}
}

func (r *requestFactory) NewGetUsernameRequest(userId string) GetUsernameRequest {
	return &getUsernameRequest{
		url:      r.url,
		apiToken: r.apiToken,
		userId:   userId,
	}
}

func (r *requestFactory) NewPostMessageRequest(channel, message string) PostMessageRequest {
	return &postMessageRequest{
		url:      r.url,
		apiToken: r.apiToken,
		channel:  channel,
		message:  message,
	}
}

func (r *requestFactory) NewStartRtmRequest() StartRtmRequest {
	return &startRtmRequest{
		url:      r.url,
		apiToken: r.apiToken,
	}
}
