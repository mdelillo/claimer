package requests

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type startRtmRequest struct {
	url      string
	apiToken string
}

func (s *startRtmRequest) Execute() (string, string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/rtm.start?token=%s", s.url, s.apiToken))
	if err != nil {
		return "", "", errors.Wrap(err, "failed to make request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", "", errors.Errorf("bad response code: %s", resp.Status)
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
		return "", "", errors.Wrap(err, "failed to parse body")
	}

	if !rtmStartResponse.Ok {
		return "", "", errors.Errorf("error in slack response: %s", rtmStartResponse.Error)
	}

	return rtmStartResponse.Url, rtmStartResponse.Self.Id, nil
}
