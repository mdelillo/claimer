package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type usernameRequest struct {
	url      string
	apiToken string
	userId   string
}

func (u *usernameRequest) Execute() (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/users.info?token=%s&user=%s", u.url, u.apiToken, u.userId))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("error getting user info: " + resp.Status)
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
		return "", err
	}

	if !apiResponse.Ok {
		return "", fmt.Errorf("error in slack response: " + apiResponse.Error)
	}

	return apiResponse.User.Name, nil
}
