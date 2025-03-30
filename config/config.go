package config

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var config *Config

// Config represents a configuration parse
type Config struct {
	RoF2Path string `yaml:"rof2path"`
	LSPath   string `yaml:"lspath"`
	baseName string
}

func Get() *Config {
	if config == nil {
		panic("config is nil")
	}
	return config
}

// New creates a new configuration
func New(ctx context.Context, baseName string) (*Config, error) {
	var f *os.File
	cfg := Config{
		baseName: baseName,
	}
	path := baseName + ".yaml"

	isNewConfig := false
	fi, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("config info: %w", err)
		}
		f, err = os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("create %s.yaml: %w", baseName, err)
		}
		fi, err = os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("new config info: %w", err)
		}
		isNewConfig = true
	}
	if !isNewConfig {
		f, err = os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open config: %w", err)
		}
	}

	defer f.Close()
	if fi.IsDir() {
		return nil, fmt.Errorf("%s.yaml is a directory, should be a file", baseName)
	}

	if isNewConfig {
		enc := yaml.NewEncoder(f)
		cfg = Config{}
		err = enc.Encode(cfg)
		if err != nil {
			return nil, fmt.Errorf("encode default: %w", err)
		}
		return &cfg, nil
	}

	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return nil, fmt.Errorf("decode %s.yaml: %w", baseName, err)
	}

	config = &cfg
	return &cfg, nil
}

// Verify returns an error if configuration appears off
func (c *Config) Verify() error {

	return nil
}

// Save writes the config to disk
func (c *Config) Save() error {
	w, err := os.Create(fmt.Sprintf("%s.yaml", c.baseName))
	if err != nil {
		return fmt.Errorf("create %s.yaml: %w", c.baseName, err)
	}
	defer w.Close()

	enc := yaml.NewEncoder(w)
	err = enc.Encode(c)
	if err != nil {
		return fmt.Errorf("encode default: %w", err)
	}
	return nil
}
