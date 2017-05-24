package requests

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

type startRtmRequest struct {
	url      string
	apiToken string
}

func (s *startRtmRequest) Execute() (string, string, error) {
	body, err := get(fmt.Sprintf("%s/api/rtm.start?token=%s", s.url, s.apiToken))
	if err != nil {
		return "", "", err
	}

	var startRtmResponse struct {
		Url  string
		Self struct {
			Id string
		}
	}
	if err := json.Unmarshal(body, &startRtmResponse); err != nil {
		return "", "", errors.Wrap(err, "failed to parse body")
	}

	return startRtmResponse.Url, startRtmResponse.Self.Id, nil
}
