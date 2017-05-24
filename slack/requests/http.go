package requests

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func get(url string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return makeRequest(request)
}

func postForm(url string, form url.Values) ([]byte, error) {
	request, err := http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return makeRequest(request)
}

func makeRequest(request *http.Request) ([]byte, error) {
	httpResponse, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != 200 {
		return nil, errors.Errorf("bad response code: %s", httpResponse.Status)
	}

	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	var slackResponse struct {
		Ok    bool
		Error string
	}
	if err := json.Unmarshal(body, &slackResponse); err != nil {
		return nil, errors.Wrap(err, "failed to parse body")
	}

	if !slackResponse.Ok {
		return nil, errors.Errorf("error in slack response: %s", slackResponse.Error)
	}

	return body, nil
}
