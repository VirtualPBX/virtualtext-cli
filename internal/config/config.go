package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type File struct {
	Host  string `yaml:"host"`
	Token string `yaml:"token"`
}

func Path() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "virtualtext", "config.yaml")
}

func Load() (*File, error) {
	cfg := &File{
		Host:  os.Getenv("VIRTUALTEXT_HOST"),
		Token: os.Getenv("VIRTUALTEXT_API_KEY"),
	}
	data, err := os.ReadFile(Path())
	if err == nil {
		var file File
		if err := yaml.Unmarshal(data, &file); err != nil {
			return nil, err
		}
		if cfg.Host == "" {
			cfg.Host = file.Host
		}
		if cfg.Token == "" {
			cfg.Token = file.Token
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	return cfg, nil
}

func Save(cfg *File) error {
	dir := filepath.Dir(Path())
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0o600)
}
