package config

import (
	"errors"
	"io/ioutil"
	"os"

	"fmt"

	"gopkg.in/yaml.v2"
)

// Config holds data needed to run vigilante
type Config struct {
	SlackToken   string `yaml:"slack_token"`
	GithubToken  string `yaml:"github_token"`
	Maximum      int    `yaml:"maximum"`
	Channel      string `yaml:"channel"`
	Organization string `yaml:"organization"`
}

// Load reads configuration from the passed file name
func Load(src string) ([]byte, error) {
	f, err := os.Open(src)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load configuration file. Check that %s exists and that it can be read. Exiting...", src)
	}
	return ioutil.ReadAll(f)
}

// Parse unmarshals the data into a YAML and validates it
func Parse(data []byte) (*Config, error) {
	var err error
	c := &Config{}
	if err = yaml.Unmarshal(data, c); err != nil {
		return c, err
	}
	err = c.validate()
	return c, err
}

func (c *Config) validate() error {
	if c.SlackToken == "" {
		return errors.New("Invalid Slack token.")
	}

	if c.GithubToken == "" {
		return errors.New("Invalid Github token.")
	}

	if c.Maximum < 1 {
		return errors.New("The maximum pull request limit must be at least 1.")
	}

	if c.Channel == "" {
		return errors.New("No channel defined for sending communications.")
	}

	if c.Organization == "" {
		return errors.New("No organization to watch.")
	}

	return nil
}
