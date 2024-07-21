package spotifyhelper

import (
	"context"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

type SpotifyClient struct {
	client *spotify.Client
}

func NewSpotifyClient(clientID, clientSecret string) (*SpotifyClient, error) {
	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		return nil, err
	}
	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)

	return &SpotifyClient{client: client}, nil
}

func (s *SpotifyClient) GetPlaylistTracks(ctx context.Context, playlistID string) (*spotify.FullPlaylist, error) {
	return s.client.GetPlaylist(ctx, spotify.ID(playlistID))
}

func (s *SpotifyClient) GetLikedSongs(ctx context.Context) ([]spotify.SimpleTrack, error) {
	var tracks []spotify.SimpleTrack
	likedTracks, err := s.client.CurrentUsersTracks(ctx)
	if err != nil {
		return nil, err
	}

	for _, item := range likedTracks.Tracks {
		tracks = append(tracks, item.SimpleTrack)
	}
	return tracks, nil
}
