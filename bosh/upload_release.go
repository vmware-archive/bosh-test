package bosh

import (
	"fmt"
	"io"
	"net/http"
)

func (c Client) UploadRelease(ioReader io.Reader) (int, error) {
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/releases", c.config.URL), ioReader)
	if err != nil {
		return 0, err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)

	response, err := transport.RoundTrip(request)
	if err != nil {
		return 0, err
	}

	if response.StatusCode != http.StatusFound {
		return 0, fmt.Errorf("unexpected response %s", response.Status)
	}

	return c.checkTaskStatus(response.Header.Get("Location"))
}
