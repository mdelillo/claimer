package requests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type startRtmRequest struct {
	url      string
	apiToken string
}

func (s *startRtmRequest) Execute() (string, string, error) {
	resp, err := http.Get(s.url + "/api/rtm.start?token=" + s.apiToken)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("error starting RTM session: %s", resp.Status)
	}

	var rtmStartResponse struct {
		Ok    bool
		Error string
		Url   string
		Self  struct {
			Id string
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if err := json.Unmarshal(body, &rtmStartResponse); err != nil {
		return "", "", err
	}

	if !rtmStartResponse.Ok {
		return "", "", fmt.Errorf("error in slack response: %s", rtmStartResponse.Error)
	}

	return rtmStartResponse.Url, rtmStartResponse.Self.Id, nil
}
