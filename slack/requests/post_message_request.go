package requests

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
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
		return errors.Wrap(err, "failed to make request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.Errorf("bad response code: %s", resp.Status)
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
		return errors.Wrap(err, "failed to parse body")
	}

	if !apiResponse.Ok {
		return errors.Errorf("error in slack response: %s", apiResponse.Error)
	}
	return nil

}
