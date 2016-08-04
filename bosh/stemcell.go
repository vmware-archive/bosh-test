package bosh

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
)

type Stemcell struct {
	Name     string
	Versions []string
}

func NewStemcell() Stemcell {
	return Stemcell{}
}

func (c Client) Stemcell(name string) (Stemcell, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/stemcells", c.config.URL), nil)
	if err != nil {
		return Stemcell{}, err
	}

	request.SetBasicAuth(c.config.Username, c.config.Password)
	response, err := client.Do(request)
	if err != nil {
		return Stemcell{}, err
	}

	if response.StatusCode == http.StatusNotFound {
		return Stemcell{}, fmt.Errorf("stemcell %s could not be found", name)
	}

	if response.StatusCode != http.StatusOK {
		body, err := bodyReader(response.Body)
		if err != nil {
			return Stemcell{}, err
		}
		defer response.Body.Close()

		return Stemcell{}, fmt.Errorf("unexpected response %d %s:\n%s", response.StatusCode, http.StatusText(response.StatusCode), body)
	}

	stemcells := []struct {
		Name    string
		Version string
	}{}

	err = json.NewDecoder(response.Body).Decode(&stemcells)
	if err != nil {
		return Stemcell{}, err
	}

	stemcell := NewStemcell()
	stemcell.Name = name

	for _, s := range stemcells {
		if s.Name == name {
			stemcell.Versions = append(stemcell.Versions, s.Version)
		}
	}

	return stemcell, nil
}

func (s Stemcell) Latest() (string, error) {
	tmp := []int{}

	if len(s.Versions) == 0 {
		return "", errors.New("no stemcell versions found, cannot get latest")
	}

	for _, version := range s.Versions {
		num, err := strconv.Atoi(version)
		if err != nil {
			return s.Versions[len(s.Versions)-1], nil
		}
		tmp = append(tmp, num)
	}
	sort.Ints(tmp)

	s.Versions = []string{}

	for _, version := range tmp {
		s.Versions = append(s.Versions, strconv.Itoa(version))
	}

	return s.Versions[len(s.Versions)-1], nil
}
