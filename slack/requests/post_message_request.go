package requests

import (
	"fmt"
	"net/url"
	"encoding/json"
	"net/http"
	"io/ioutil"
)

type postMessageRequest struct {
	url      string
	apiToken string
	channel  string
	message  string
}

func (p *postMessageRequest) Execute() error {
	resp, err := http.PostForm(
		fmt.Sprintf("%s/api/chat.postMessage", p.url),
		url.Values{
			"token":   {p.apiToken},
			"channel": {p.channel},
			"text":    {p.message},
			"as_user": {"true"},
		},
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("error posting message: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResponse struct {
		Ok    bool
		Error string
	}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return err
	}

	if !apiResponse.Ok {
		return fmt.Errorf("error in slack response: %s", apiResponse.Error)
	}
	return nil

}
