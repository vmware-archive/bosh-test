package bosh

import (
	"bytes"
	"fmt"
	"net/http"
)

func (c Client) UpdateCloudConfig(cloudConfig []byte) error {
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/cloud_configs", c.config.URL), bytes.NewBuffer(cloudConfig))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "text/yaml")

	response, err := c.makeRequest(request)
	if err != nil {
		return err
	}

	body, err := bodyReader(response.Body)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), body)
	}

	return nil
}
