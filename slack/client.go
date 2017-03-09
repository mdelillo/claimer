package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type client struct {
	apiToken string
	url      string
}

func NewClient(url, apiToken string) *client {
	return &client{
		apiToken: apiToken,
		url:      url,
	}
}

func (c *client) PostMessage(channel, message string) error {
	resp, err := http.PostForm(
		fmt.Sprintf("%s/api/chat.postMessage", c.url),
		url.Values{
			"token":   {c.apiToken},
			"channel": {channel},
			"text":    {message},
			"as_user": {"true"},
		},
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("error posting to slack: %s", resp.Status)
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
		return fmt.Errorf("error posting to slack: %s", apiResponse.Error)
	}
	return nil
}
