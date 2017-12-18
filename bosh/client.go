package bosh

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	client     = http.DefaultClient
	transport  = http.DefaultTransport
	bodyReader = ioutil.ReadAll
)

type Config struct {
	URL                 string
	Host                string
	DirectorCACert      string
	Username            string
	Password            string
	TaskPollingInterval time.Duration
	AllowInsecureSSL    bool
	Transport           http.RoundTripper
	UAA                 bool
}

type Client struct {
	config Config
}

type Task struct {
	Id     int
	State  string
	Result string
}

func NewClient(config Config) Client {
	if config.TaskPollingInterval == time.Duration(0) {
		config.TaskPollingInterval = 5 * time.Second
	}

	config.Transport = http.DefaultTransport
	if config.AllowInsecureSSL {
		config.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client = &http.Client{
		Transport: config.Transport,
	}

	transport = config.Transport
	return Client{
		config: config,
	}
}

func (c Client) GetConfig() Config {
	return c.config
}

func (c Client) rewriteURL(uri string) (string, error) {
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", err
	}

	parsedURL.Scheme = ""
	parsedURL.Host = ""

	return c.config.URL + parsedURL.String(), nil
}

func (c Client) checkTask(location string) (Task, error) {
	location, err := c.rewriteURL(location)
	if err != nil {
		return Task{}, err
	}

	var task Task
	request, err := http.NewRequest("GET", location, nil)
	if err != nil {
		return task, err
	}
	response, err := c.makeRequest(request)

	if err != nil {
		return task, err
	}

	err = json.NewDecoder(response.Body).Decode(&task)
	if err != nil {
		return task, err
	}

	return task, nil
}

func (c Client) checkTaskStatus(location string) (int, error) {
	for {
		task, err := c.checkTask(location)
		if err != nil {
			return 0, err
		}

		switch task.State {
		case "done":
			return task.Id, nil
		case "error":
			taskOutputs, err := c.GetTaskOutput(task.Id)
			if err != nil {
				return task.Id, fmt.Errorf("failed to get full bosh task event log, bosh task failed with an error status %q", task.Result)
			}
			return task.Id, taskOutputs[len(taskOutputs)-1].Error
		case "errored":
			taskOutputs, err := c.GetTaskOutput(task.Id)
			if err != nil {
				return task.Id, fmt.Errorf("failed to get full bosh task event log, bosh task failed with an errored status %q", task.Result)
			}
			return task.Id, taskOutputs[len(taskOutputs)-1].Error
		case "cancelled":
			return task.Id, errors.New("bosh task was cancelled")
		default:
			time.Sleep(c.config.TaskPollingInterval)
		}
	}
}

func (c Client) makeRequest(request *http.Request) (*http.Response, error) {
	if c.config.UAA {
		urlParts, err := url.Parse(c.config.URL)
		if err != nil {
			return &http.Response{}, err
		}

		boshHost, _, err := net.SplitHostPort(urlParts.Host)
		if err != nil {
			return &http.Response{}, err
		}

		ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
			Transport: c.config.Transport,
		})
		conf := &clientcredentials.Config{
			ClientID:     c.config.Username,
			ClientSecret: c.config.Password,
			TokenURL:     fmt.Sprintf("https://%s:8443/oauth/token", boshHost),
		}

		httpClient := conf.Client(ctx)
		return httpClient.Do(request)
	} else {
		request.SetBasicAuth(c.config.Username, c.config.Password)
		return transport.RoundTrip(request)
	}
}
