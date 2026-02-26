package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient_TokenAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/users/me" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer validtoken" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"id": "user1"})
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "validtoken", "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.token != "validtoken" {
		t.Fatal("token not set")
	}
}

func TestNewClient_InvalidToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL, "badtoken", "", false)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestNewClient_LoginAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v4/users/login" && r.Method == "POST" {
			w.Header().Set("Token", "logintoken")
			json.NewEncoder(w).Encode(map[string]string{"id": "user1"})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	t.Setenv("MM_PASSWORD", "secret")

	c, err := NewClient(srv.URL, "", "admin", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.token != "logintoken" {
		t.Fatalf("expected logintoken, got %q", c.token)
	}
}

func TestDoJSON_ForbiddenError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := &Client{
		baseURL:    srv.URL,
		httpClient: http.DefaultClient,
		token:      "tok",
	}

	var dst map[string]interface{}
	_, err := c.doJSON("GET", "/test", &dst)
	if err == nil {
		t.Fatal("expected error for 403")
	}
}
