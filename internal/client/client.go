package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/term"
)

// Client wraps HTTP calls to the Mattermost API v4.
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
	verbose    bool
}

// NewClient creates a client authenticated by token or username/password login.
func NewClient(baseURL, token, username string, verbose bool) (*Client, error) {
	c := &Client{
		baseURL: baseURL + "/api/v4",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		verbose: verbose,
	}

	if token != "" {
		c.token = token
		// Verify the token works
		if err := c.ping(); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
		return c, nil
	}

	// Username/password login
	password := os.Getenv("MM_PASSWORD")
	if password == "" {
		fmt.Fprintf(os.Stderr, "Password for %s: ", username)
		pwBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr) // newline after password input
		if err != nil {
			return nil, fmt.Errorf("failed to read password: %w", err)
		}
		password = string(pwBytes)
	}

	if err := c.login(username, password); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) login(username, password string) error {
	body, err := json.Marshal(LoginRequest{
		LoginID:  username,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/users/login", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d — check username and password", resp.StatusCode)
	}

	c.token = resp.Header.Get("Token")
	if c.token == "" {
		return fmt.Errorf("login succeeded but no token was returned")
	}
	return nil
}

func (c *Client) ping() error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/users/me", nil)
	if err != nil {
		return err
	}
	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach server: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid token (401 Unauthorized)")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from /users/me", resp.StatusCode)
	}
	return nil
}

func (c *Client) setAuth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
}

// doJSON executes a request and decodes the JSON response into dst.
func (c *Client) doJSON(method, path string, dst interface{}) (*http.Response, error) {
	req, err := http.NewRequest(method, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", path, err)
	}
	c.setAuth(req)

	if c.verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s %s\n", method, path)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", path, err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		return resp, fmt.Errorf("unauthorized (401) for %s — check token permissions", path)
	}
	if resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()
		return resp, fmt.Errorf("forbidden (403) for %s — insufficient permissions", path)
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return resp, fmt.Errorf("not found (404) for %s", path)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return resp, fmt.Errorf("unexpected status %d for %s: %s", resp.StatusCode, path, string(body))
	}

	if dst != nil {
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
			return resp, fmt.Errorf("failed to decode response from %s: %w", path, err)
		}
	}

	return resp, nil
}

// paginateGet fetches all pages from a paginated endpoint.
// decode is called for each page's response body and should return the number of items decoded.
func (c *Client) paginateGet(basePath string, perPage int, decode func(body io.Reader) (int, error)) error {
	for page := 0; ; page++ {
		path := fmt.Sprintf("%s?page=%d&per_page=%d", basePath, page, perPage)
		req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
		if err != nil {
			return fmt.Errorf("failed to create request for %s: %w", path, err)
		}
		c.setAuth(req)

		if c.verbose {
			fmt.Fprintf(os.Stderr, "[DEBUG] GET %s\n", path)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("request to %s failed: %w", path, err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return fmt.Errorf("unexpected status %d for %s: %s", resp.StatusCode, path, string(body))
		}

		n, err := decode(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to decode page %d of %s: %w", page, basePath, err)
		}

		if n < perPage {
			break // last page
		}
	}
	return nil
}

// Logf prints a verbose log message to stderr.
func (c *Client) Logf(format string, args ...interface{}) {
	if c.verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}
