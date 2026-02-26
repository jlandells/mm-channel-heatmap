package client

// Team represents a Mattermost team.
type Team struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// Channel represents a Mattermost channel.
type Channel struct {
	ID          string `json:"id"`
	TeamID      string `json:"team_id"`
	Type        string `json:"type"`     // "O" = public, "P" = private, "D" = DM, "G" = GM
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	DeleteAt    int64  `json:"delete_at"` // 0 if not deleted
	CreateAt    int64  `json:"create_at"`
}

// Post represents a Mattermost post.
type Post struct {
	ID       string `json:"id"`
	CreateAt int64  `json:"create_at"`
}

// PostList represents the Mattermost paginated post response.
type PostList struct {
	Order []string         `json:"order"`
	Posts map[string]*Post `json:"posts"`
}

// ChannelStats represents channel statistics from the stats endpoint.
type ChannelStats struct {
	ChannelID   string `json:"channel_id"`
	MemberCount int    `json:"member_count"`
}

// LoginRequest is the body for POST /users/login.
type LoginRequest struct {
	LoginID  string `json:"login_id"`
	Password string `json:"password"`
}
