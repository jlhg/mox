package mox

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Mox MoxConfig `toml:"mox"`
}

func (c *Config) Init() (err error) {
	toCreatePaths := []string{
		c.Mox.DownloadPath,
	}

	for _, v := range toCreatePaths {
		err = os.MkdirAll(v, os.ModePerm)
		if err != nil {
			err = fmt.Errorf("failed to create directory (%s): %w", v, err)
			goto EndConfigInit
		}
	}

EndConfigInit:
	return
}

type MoxConfig struct {
	Email        string `toml:"email"`
	Password     string `toml:"password"`
	DownloadPath string `toml:"download_path"`
}

func NewConfig(path string) (config *Config, err error) {
	var d []byte
	config = &Config{}

	d, err = os.ReadFile(path)
	if err != nil {
		return
	}

	err = toml.Unmarshal(d, config)

	return
}
