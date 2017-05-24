package requests

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

type getUsernameRequest struct {
	url      string
	apiToken string
	userId   string
}

func (g *getUsernameRequest) Execute() (string, error) {
	body, err := get(fmt.Sprintf("%s/api/users.info?token=%s&user=%s", g.url, g.apiToken, g.userId))
	if err != nil {
		return "", err
	}

	var getUsernameResponse struct {
		User struct {
			Name string
		}
	}
	if err := json.Unmarshal(body, &getUsernameResponse); err != nil {
		return "", errors.Wrap(err, "failed to parse body")
	}

	return getUsernameResponse.User.Name, nil
}
