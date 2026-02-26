package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchChannelsForTeam_FiltersDMsAndArchived(t *testing.T) {
	channels := []Channel{
		{ID: "1", Type: "O", Name: "public", DeleteAt: 0},
		{ID: "2", Type: "D", Name: "dm-channel", DeleteAt: 0},       // DM — exclude
		{ID: "3", Type: "G", Name: "group-msg", DeleteAt: 0},        // GM — exclude
		{ID: "4", Type: "O", Name: "archived", DeleteAt: 1234567890}, // Archived — exclude
		{ID: "5", Type: "O", Name: "another-public", DeleteAt: 0},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v4/teams/team1/channels":
			json.NewEncoder(w).Encode(channels)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := &Client{
		baseURL:    srv.URL + "/api/v4",
		httpClient: http.DefaultClient,
		token:      "tok",
	}

	result, err := c.FetchChannelsForTeam("team1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(result))
	}

	for _, ch := range result {
		if ch.Type == "D" || ch.Type == "G" {
			t.Errorf("unexpected channel type %q for %s", ch.Type, ch.Name)
		}
		if ch.DeleteAt != 0 {
			t.Errorf("unexpected archived channel %s", ch.Name)
		}
	}
}

func TestFetchChannelsForTeam_PrivateChannels403_WarnNotFatal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/teams/team1/channels":
			json.NewEncoder(w).Encode([]Channel{
				{ID: "1", Type: "O", Name: "public"},
			})
		case "/api/v4/teams/team1/channels/private":
			w.WriteHeader(http.StatusForbidden)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := &Client{
		baseURL:    srv.URL + "/api/v4",
		httpClient: http.DefaultClient,
		token:      "tok",
	}

	result, err := c.FetchChannelsForTeam("team1", true)
	if err != nil {
		t.Fatalf("should not fail on private 403: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(result))
	}
}

func TestShouldIncludeChannel(t *testing.T) {
	tests := []struct {
		name   string
		ch     Channel
		expect bool
	}{
		{"public", Channel{Type: "O", DeleteAt: 0}, true},
		{"private", Channel{Type: "P", DeleteAt: 0}, true},
		{"dm", Channel{Type: "D", DeleteAt: 0}, false},
		{"gm", Channel{Type: "G", DeleteAt: 0}, false},
		{"archived_public", Channel{Type: "O", DeleteAt: 123}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldIncludeChannel(tt.ch)
			if got != tt.expect {
				t.Errorf("shouldIncludeChannel(%+v) = %v, want %v", tt.ch, got, tt.expect)
			}
		})
	}
}
