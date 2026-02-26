package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func makePostList(ids []string, createAts []int64) PostList {
	pl := PostList{
		Order: ids,
		Posts: make(map[string]*Post),
	}
	for i, id := range ids {
		pl.Posts[id] = &Post{ID: id, CreateAt: createAts[i]}
	}
	return pl
}

func TestCountPostsInWindow_AllInWindow(t *testing.T) {
	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	pl := makePostList(
		[]string{"p3", "p2", "p1"},
		[]int64{
			time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC).UnixMilli(),
			time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC).UnixMilli(),
			time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC).UnixMilli(),
		},
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(pl)
	}))
	defer srv.Close()

	c := &Client{
		baseURL:    srv.URL + "/api/v4",
		httpClient: http.DefaultClient,
		token:      "tok",
	}

	count, err := c.CountPostsInWindow("ch1", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3, got %d", count)
	}
}

func TestCountPostsInWindow_EarlyStop(t *testing.T) {
	since := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	pl := makePostList(
		[]string{"p3", "p2", "p1"},
		[]int64{
			time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC).UnixMilli(),
			time.Date(2025, 1, 12, 0, 0, 0, 0, time.UTC).UnixMilli(),
			time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC).UnixMilli(), // before window
		},
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(pl)
	}))
	defer srv.Close()

	c := &Client{
		baseURL:    srv.URL + "/api/v4",
		httpClient: http.DefaultClient,
		token:      "tok",
	}

	count, err := c.CountPostsInWindow("ch1", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

func TestCountPostsInWindow_ExactBoundary(t *testing.T) {
	since := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	sinceMillis := since.UnixMilli()

	pl := makePostList(
		[]string{"p2", "p1"},
		[]int64{
			sinceMillis,     // exactly at the boundary — should be counted
			sinceMillis - 1, // 1ms before — should not be counted
		},
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(pl)
	}))
	defer srv.Close()

	c := &Client{
		baseURL:    srv.URL + "/api/v4",
		httpClient: http.DefaultClient,
		token:      "tok",
	}

	count, err := c.CountPostsInWindow("ch1", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1, got %d", count)
	}
}

func TestCountPostsInWindow_EmptyChannel(t *testing.T) {
	pl := PostList{
		Order: []string{},
		Posts: map[string]*Post{},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(pl)
	}))
	defer srv.Close()

	c := &Client{
		baseURL:    srv.URL + "/api/v4",
		httpClient: http.DefaultClient,
		token:      "tok",
	}

	count, err := c.CountPostsInWindow("ch1", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}
}

func TestCountPostsInWindow_MultiPage(t *testing.T) {
	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	callCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First page: 200 posts, all in window
			ids := make([]string, 200)
			posts := make(map[string]*Post)
			for i := 0; i < 200; i++ {
				id := "p" + string(rune('a'+i%26)) + string(rune('0'+i/26))
				ids[i] = id
				posts[id] = &Post{
					ID:       id,
					CreateAt: time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC).UnixMilli(),
				}
			}
			json.NewEncoder(w).Encode(PostList{Order: ids, Posts: posts})
		} else {
			// Second page: 1 post before the window (triggers early stop)
			json.NewEncoder(w).Encode(makePostList(
				[]string{"old"},
				[]int64{time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC).UnixMilli()},
			))
		}
	}))
	defer srv.Close()

	c := &Client{
		baseURL:    srv.URL + "/api/v4",
		httpClient: http.DefaultClient,
		token:      "tok",
	}

	count, err := c.CountPostsInWindow("ch1", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 200 {
		t.Fatalf("expected 200, got %d", count)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 API calls, got %d", callCount)
	}
}
