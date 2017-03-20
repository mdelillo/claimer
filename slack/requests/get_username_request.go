package requests

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

type getUsernameRequest struct {
	url      string
	apiToken string
	userId   string
}

func (g *getUsernameRequest) Execute() (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/users.info?token=%s&user=%s", g.url, g.apiToken, g.userId))
	if err != nil {
		return "", errors.Wrap(err, "failed to make request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", errors.Errorf("bad response code: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var apiResponse struct {
		Ok    bool
		Error string
		User  struct {
			Name string
		}
	}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return "", errors.Wrap(err, "failed to parse body")
	}

	if !apiResponse.Ok {
		return "", errors.Errorf("error in slack response: %s", apiResponse.Error)
	}

	return apiResponse.User.Name, nil
}
