package requests

import (
	"fmt"
	"net/url"
)

type postMessageRequest struct {
	url      string
	apiToken string
	channel  string
	message  string
}

func (p *postMessageRequest) Execute() error {
	form := url.Values{}
	form.Set("token", p.apiToken)
	form.Add("channel", p.channel)
	form.Add("text", p.message)
	form.Add("as_user", "true")

	_, err := postForm(fmt.Sprintf("%s/api/chat.postMessage", p.url), form)
	return err
}
