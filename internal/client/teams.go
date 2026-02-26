package client

import (
	"encoding/json"
	"fmt"
	"io"
)

// FetchAllTeams returns all teams visible to the authenticated user.
func (c *Client) FetchAllTeams() ([]Team, error) {
	var teams []Team

	err := c.paginateGet("/teams", 200, func(body io.Reader) (int, error) {
		var page []Team
		if err := json.NewDecoder(body).Decode(&page); err != nil {
			return 0, err
		}
		teams = append(teams, page...)
		return len(page), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch teams: %w", err)
	}

	return teams, nil
}

// ResolveTeamByName looks up a team by its URL name.
func (c *Client) ResolveTeamByName(name string) (*Team, error) {
	var team Team
	_, err := c.doJSON("GET", "/teams/name/"+name, &team)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve team %q: %w", name, err)
	}
	return &team, nil
}
