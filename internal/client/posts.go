package client

import (
	"fmt"
	"time"
)

// CountPostsInWindow counts posts in a channel created at or after the given time.
// Posts are fetched newest-first; pagination stops early when posts fall outside the window.
func (c *Client) CountPostsInWindow(channelID string, since time.Time) (int, error) {
	sinceMillis := since.UnixMilli()
	count := 0

	for page := 0; ; page++ {
		path := fmt.Sprintf("/channels/%s/posts?page=%d&per_page=200", channelID, page)

		var postList PostList
		_, err := c.doJSON("GET", path, &postList)
		if err != nil {
			return 0, fmt.Errorf("failed to fetch posts for channel %s page %d: %w", channelID, page, err)
		}

		if len(postList.Order) == 0 {
			break // no more posts
		}

		pastWindow := false
		for _, postID := range postList.Order {
			post, ok := postList.Posts[postID]
			if !ok {
				continue
			}
			if post.CreateAt >= sinceMillis {
				count++
			} else {
				// Posts are newest-first; once we see one before the window, stop
				pastWindow = true
				break
			}
		}

		if pastWindow || len(postList.Order) < 200 {
			break
		}
	}

	return count, nil
}

