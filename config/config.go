package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	SpotifyID           string        `yaml:"spotify_id"`
	SpotifySecret       string        `yaml:"spotify_secret"`
	PlaylistsToDownload []string      `yaml:"playlists_to_download"`
	Aria2RPCEndpoint    string        `yaml:"aria2_rpc_endpoint"`
	DownloadPath        string        `yaml:"download_path"`
	Interval            time.Duration `yaml:"interval"`
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}
