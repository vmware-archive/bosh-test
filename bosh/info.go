package bosh

import (
	"fmt"

	"github.com/cloudfoundry/bosh-cli/director"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type DirectorInfo struct {
	UUID string
	CPI  string
}

func (c Client) Info() (DirectorInfo, error) {
	logger := boshlog.NewLogger(boshlog.LevelNone)

	factoryConfig, err := director.NewConfigFromURL(c.config.URL)
	if err != nil {
		return DirectorInfo{}, fmt.Errorf("Creating factory config from url %s: %s", c.config.URL, err)
	}

	factoryConfig.CACert = c.config.DirectorCACert

	d, err := director.NewFactory(logger).New(factoryConfig, director.NewNoopTaskReporter(), director.NewNoopFileReporter())
	if err != nil {
		return DirectorInfo{}, fmt.Errorf("Creating director with factory: %s", err)
	}

	info, err := d.Info()
	if err != nil {
		return DirectorInfo{}, fmt.Errorf("Getting /info from director: %s", err)
	}

	output := DirectorInfo{
		UUID: info.UUID,
		CPI:  info.CPI,
	}

	return output, nil
}
