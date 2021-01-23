package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/browser"
	"github.com/zmb3/spotify"
)

// SpotifyCallbackPath indicates the path to redirect users to for Spotify OAuth2 flow
const SpotifyCallbackPath = "/spotify/callback"

// SpotifyCallbackPort indicates the port to redirect users to for Spotify OAuth2 flow
const SpotifyCallbackPort = 3000

// SpotifyManager owns Spotify related items
type SpotifyManager struct {
	Auth         spotify.Authenticator
	Client       spotify.Client
	AuthState    *uuid.UUID
	LoginServer  *http.Server
	LoginChannel chan *spotify.Client
}

// SetupSpotify sets up the Spotify integration
func (l *Lightshow) SetupSpotify() {

	// Check if we have a client ID and secret
	if l.Config.Spotify.ClientID == "" || l.Config.Spotify.ClientSecret == "" {
		l.Logger.Error("Cannot set up Spotify client, missing Client ID and Client Secret from configuration file")
		return
	}

	l.Spotify.Auth = spotify.NewAuthenticator(fmt.Sprintf("http://localhost:%d%s", SpotifyCallbackPort, SpotifyCallbackPath), spotify.ScopeUserReadPrivate, spotify.ScopeUserModifyPlaybackState, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState)
	l.Spotify.Auth.SetAuthInfo(l.Config.Spotify.ClientID, l.Config.Spotify.ClientSecret)

	// if l.Config.Spotify.UserToken != nil {
	// 	l.Spotify.Client = l.Spotify.Auth.NewClient(l.Config.Spotify.UserToken)

	// 	userData, err := l.Spotify.Client.CurrentUser()
	// 	if err != nil {
	// 		l.Logger.WithError(err).Error("Failed to get user data from Spotify")
	// 		return
	// 	}

	// 	l.Logger.WithField("spotify-user", userData.User.DisplayName).Info("Logged into Spotify")
	// 	return
	// }

	l.Spotify.LoginChannel = make(chan *spotify.Client, 1)

	loginState := uuid.New()
	l.Spotify.AuthState = &loginState
	loginURL := l.Spotify.Auth.AuthURL(loginState.String())

	loginHandler := http.NewServeMux()
	loginHandler.HandleFunc(SpotifyCallbackPath, l.spotifyLoginHandler)

	l.Spotify.LoginServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", SpotifyCallbackPort),
		Handler:      loginHandler,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	}

	go func() {
		l.Logger.Warn(l.Spotify.LoginServer.ListenAndServe())
	}()

	err := browser.OpenURL(loginURL)
	if err != nil {
		l.Logger.WithError(err).Warnf("Could not launch your browser, please paste this URL into your browser: %s", loginURL)
	}

	l.Logger.Info("Waiting for you to login to Spotify...")
	spotifyClient := <-l.Spotify.LoginChannel
	if spotifyClient != nil {
		l.Spotify.Client = *spotifyClient

		userData, err := l.Spotify.Client.CurrentUser()
		if err != nil {
			l.Logger.WithError(err).Error("Failed to get user data from Spotify")
			return
		}

		l.Logger.WithField("spotify-user", userData.User.DisplayName).Info("Logged into Spotify")

		l.Spotify.LoginServer.Shutdown(context.Background())
		l.Spotify.LoginServer = nil
		return
	}

	l.Logger.Fatal("Failed to login to Spotify")
}

func (l *Lightshow) spotifyLoginHandler(w http.ResponseWriter, r *http.Request) {
	token, err := l.Spotify.Auth.Token(l.Spotify.AuthState.String(), r)
	if err != nil {
		l.Logger.WithError(err).Error("Failed to login to spotify")
		l.Spotify.LoginChannel <- nil

		w.WriteHeader(403)
		w.Write([]byte("Login failed"))

		return
	}

	client := l.Spotify.Auth.NewClient(token)

	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(200)
	w.Write([]byte("<h1>Login succeeded. You may close this window.</h1> <script>window.setTimeout(function() { window.close(); }, 3000);</script>"))

	l.Spotify.AuthState = nil
	// l.Config.Spotify.UserToken = token
	// l.SaveConfig()

	l.Spotify.LoginChannel <- &client
}

func (l *Lightshow) testSpotifyThing() {
	var trackID spotify.ID = "0Oh5sFv6voDGSaa6KzCJlX"

	audioAnalysis, err := l.Spotify.Client.GetAudioAnalysis(trackID)
	if err != nil {
		l.Logger.WithError(err).Error("Failed to get track analysis")
		return
	}

	var uri spotify.URI = "spotify:album:10anzcAunKpB8AUxvR7siL"
	err = l.Spotify.Client.PlayOpt(&spotify.PlayOptions{
		PlaybackContext: &uri,
		PlaybackOffset: &spotify.PlaybackOffset{
			Position: 2,
		},
	})
	if err != nil {
		l.ContextCancelFunc()
		l.Logger.WithError(err).Fatal("Couldn't play song")
	}

	time.Sleep(time.Millisecond * time.Duration(audioAnalysis.Tatums[0].Start*1000))
	for _, tatum := range audioAnalysis.Beats {
		if l.Context.Err() != nil {
			return
		}
		l.SetLights([]int{5, 7, 8, 9, 10, 11, 12, 13, 14, 15}, 255, 255, 255, 1)
		time.Sleep(time.Millisecond * 120)
		l.SetLights([]int{5, 7, 8, 9, 10, 11, 12, 13, 14, 15}, 255, 0, 0, 0)
		time.Sleep((time.Millisecond * time.Duration(tatum.Duration*1000)) - (time.Millisecond * 120))
	}
}
