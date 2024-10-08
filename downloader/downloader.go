package downloader

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/siku2/arigo"
)

type Downloader struct {
	rpcEndpoint  string
	downloadPath string
	DownloadChan chan DownloadEvent
	rateLimiter  chan struct{}
}

type DownloadEvent struct {
	Type         EventType
	TrackID      string
	PlaylistName string
}

type EventType int

const (
	EventTypeSongAdded EventType = iota
)

func NewDownloader(rpcEndpoint, downloadPath string) *Downloader {
	d := &Downloader{
		rpcEndpoint:  rpcEndpoint,
		downloadPath: downloadPath,
		DownloadChan: make(chan DownloadEvent),
		rateLimiter:  make(chan struct{}, 10), // buffer size 10 for rate limiting
	}

	// Fill the rateLimiter initially
	for i := 0; i < 10; i++ {
		d.rateLimiter <- struct{}{}
	}

	// Start a goroutine to refill the rateLimiter every minute
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
		tokenadder:
			for i := 0; i < 10; i++ {
				select {
				case d.rateLimiter <- struct{}{}:
				default:
					break tokenadder
				}
			}
		}
	}()

	return d
}

func (d *Downloader) SubmitLink(ctx context.Context, link, downloadPath string) (string, error) {
	// Acquire a token from the rate limiter
	<-d.rateLimiter

	// Dial Arigo
	c, err := arigo.Dial(d.rpcEndpoint, "")
	if err != nil {
		return "", fmt.Errorf("failed to connect to Aria2: %w", err)
	}

	// Submit download
	status, err := c.Download(arigo.URIs(link), &arigo.Options{
		Dir: downloadPath,
	})
	if err != nil {
		return "", fmt.Errorf("failed to submit download: %w", err)
	}

	return status.GID, nil
}

func (d *Downloader) DownloadPlaylist(ctx context.Context, playlistName string, trackIDs []string) error {
	downloadPath := filepath.Join(d.downloadPath, playlistName)
	if err := os.MkdirAll(downloadPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	for _, trackID := range trackIDs {
		link := fmt.Sprintf("https://hund.lucida.to/api/fetch/stream?url=https://open.spotify.com/track/%s&downscale=original&meta=true&private=false&country=auto", trackID)
		log.Printf("Submitting download for track %s from playlist %s", trackID, playlistName)
		go func(trackID string) {
			_, err := d.SubmitLink(ctx, link, downloadPath)
			if err != nil {
				log.Printf("Failed to submit link for track %s: %v", trackID, err)
			} else {
				log.Printf("Download submitted for track %s", trackID)
			}
		}(trackID)
	}
	return nil
}

func (d *Downloader) Run() {
	for event := range d.DownloadChan {
		if event.Type == EventTypeSongAdded {
			log.Printf("Song added: %s", event.TrackID)
			ctx := context.Background()
			playlistPath := filepath.Join(d.downloadPath, event.PlaylistName)
			link := fmt.Sprintf("https://hund.lucida.to/api/fetch/stream?url=https://open.spotify.com/track/%s&downscale=original&meta=true&private=false&country=auto", event.TrackID)
			_, err := d.SubmitLink(ctx, link, playlistPath)
			if err != nil {
				log.Printf("Failed to download song: %v", err)
			}
		} else {
			log.Printf("Unknown event type: %d", event.Type)
		}
	}
}
