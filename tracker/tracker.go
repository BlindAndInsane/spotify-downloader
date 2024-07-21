package tracker

import (
	"context"
	"log"
	"time"

	"spotify-downloader/downloader"
	"spotify-downloader/spotifyhelper"
)

type Tracker struct {
	client         *spotifyhelper.SpotifyClient
	dwn            *downloader.Downloader
	playlists      []string
	events         chan downloader.DownloadEvent
	processedSongs map[string]map[string]bool
}

func NewTracker(client *spotifyhelper.SpotifyClient, dwn *downloader.Downloader, playlists []string) *Tracker {
	processedSongs := make(map[string]map[string]bool)
	return &Tracker{
		client:         client,
		dwn:            dwn,
		playlists:      playlists,
		events:         dwn.DownloadChan,
		processedSongs: processedSongs,
	}
}

func (t *Tracker) Start(ctx context.Context, interval time.Duration) {
	for _, playlistID := range t.playlists {
		playlist, err := t.client.GetPlaylistTracks(ctx, playlistID)
		if err != nil {
			log.Printf("Failed to get playlist: %v", err)
			continue
		}

		playlistName := playlist.Name

		if _, exists := t.processedSongs[playlistName]; !exists {
			t.processedSongs[playlistName] = make(map[string]bool)
		}

		var trackIDs []string
		for _, item := range playlist.Tracks.Tracks {
			trackID := item.Track.ID.String()
			trackIDs = append(trackIDs, trackID)
			t.processedSongs[playlistName][trackID] = true
		}

		if err := t.dwn.DownloadPlaylist(ctx, playlistName, trackIDs); err != nil {
			log.Printf("Failed to download playlist: %v", err)
		}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for _, playlistID := range t.playlists {
					playlist, err := t.client.GetPlaylistTracks(ctx, playlistID)
					if err != nil {
						log.Printf("Failed to get playlist tracks: %v", err)
						continue
					}

					playlistName := playlist.Name

					if _, exists := t.processedSongs[playlistName]; !exists {
						t.processedSongs[playlistName] = make(map[string]bool)
					}

					for _, item := range playlist.Tracks.Tracks {
						trackID := item.Track.ID.String()
						if _, exists := t.processedSongs[playlistName][trackID]; !exists {
							log.Printf("detected new song %s in playlist %s\n", trackID, playlistID)
							t.events <- downloader.DownloadEvent{
								Type:         downloader.EventTypeSongAdded,
								TrackID:      trackID,
								PlaylistName: playlistName,
							}
							t.processedSongs[playlistName][trackID] = true
						}
					}
				}
				time.Sleep(interval)
			}
		}
	}()
}
