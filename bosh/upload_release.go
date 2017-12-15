package bosh

import (
	"fmt"
	"io"
	"net/http"
)

type SizeReader interface {
	io.Reader
	Size() int64
}

func NewSizeReader(reader io.Reader, size int64) SizeReader {
	return sizeReader{
		Reader: reader,
		size:   size,
	}
}

type sizeReader struct {
	io.Reader
	size int64
}

func (r sizeReader) Size() int64 {
	return r.size
}

func (c Client) UploadRelease(contents SizeReader) (int, error) {
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/releases", c.config.URL), contents)
	if err != nil {
		return 0, err
	}

	request.Header.Set("Content-Type", "application/x-compressed")
	request.ContentLength = contents.Size()

	response, err := c.makeRequest(request)
	if err != nil {
		return 0, err
	}

	body, err := bodyReader(response.Body)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusFound {
		return 0, fmt.Errorf("unexpected response %s:\n%s", response.Status, body)
	}

	return c.checkTaskStatus(response.Header.Get("Location"))
}
