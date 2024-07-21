package main

import (
	"context"
	"log"
	"spotify-downloader/config"
	"spotify-downloader/downloader"
	"spotify-downloader/spotifyhelper"
	"spotify-downloader/tracker"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	spotifyClient, err := spotifyhelper.NewSpotifyClient(cfg.SpotifyID, cfg.SpotifySecret)
	if err != nil {
		log.Fatalf("failed to create Spotify client: %v", err)
	}

	downldr := downloader.NewDownloader(cfg.Aria2RPCEndpoint, cfg.DownloadPath)
	go downldr.Run()

	trackr := tracker.NewTracker(spotifyClient, downldr, cfg.PlaylistsToDownload)
	go trackr.Start(context.Background(), cfg.Interval)

	select {}
}
