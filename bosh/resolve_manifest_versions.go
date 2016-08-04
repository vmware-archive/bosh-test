package bosh

import yaml "gopkg.in/yaml.v2"

type manifest struct {
	DirectorUUID  interface{} `yaml:"director_uuid"`
	Name          interface{} `yaml:"name"`
	Compilation   interface{} `yaml:"compilation"`
	Update        interface{} `yaml:"update"`
	Networks      interface{} `yaml:"networks"`
	ResourcePools []struct {
		Name            interface{} `yaml:"name"`
		Network         interface{} `yaml:"network"`
		Size            interface{} `yaml:"size,omitempty"`
		CloudProperties interface{} `yaml:"cloud_properties,omitempty"`
		Env             interface{} `yaml:"env,omitempty"`
		Stemcell        struct {
			Name    string `yaml:"name"`
			Version string `yaml:"version"`
		} `yaml:"stemcell"`
	} `yaml:"resource_pools"`
	Jobs       interface{} `yaml:"jobs"`
	Properties interface{} `yaml:"properties"`
	Releases   []struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	} `yaml:"releases"`
}

func (c Client) ResolveManifestVersions(manifestYAML []byte) ([]byte, error) {
	m := manifest{}
	err := yaml.Unmarshal(manifestYAML, &m)
	if err != nil {
		return nil, err
	}

	for i, r := range m.Releases {
		if r.Version == "latest" {
			release, err := c.Release(r.Name)
			if err != nil {
				return nil, err
			}
			r.Version = release.Latest()
			m.Releases[i] = r
		}
	}

	for i, pool := range m.ResourcePools {
		if pool.Stemcell.Version == "latest" {
			stemcell, err := c.Stemcell(pool.Stemcell.Name)
			if err != nil {
				return nil, err
			}
			pool.Stemcell.Version, err = stemcell.Latest()
			if err != nil {
				return nil, err
			}
			m.ResourcePools[i] = pool
		}
	}

	return yaml.Marshal(m)
}
