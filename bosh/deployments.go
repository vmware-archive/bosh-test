package bosh

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Deployment struct {
	Name string
}

func (c Client) Deployments() ([]Deployment, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/deployments", c.config.URL), nil)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(c.config.Username, c.config.Password)

	response, err := transport.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		body, err := bodyReader(response.Body)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		return nil, fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), body)
	}

	var deployments []Deployment
	err = json.NewDecoder(response.Body).Decode(&deployments)
	if err != nil {
		return nil, err
	}

	return deployments, nil
}
