package ai

import (
	"context"
	"fmt"
)

// TestConnection verifies the AI provider responds to a simple prompt.
func TestConnection(ctx context.Context, client *Client) (string, error) {
	if client == nil {
		return "", fmt.Errorf("no AI client configured")
	}
	return client.TextCompletion(ctx, []chatMessage{
		{Role: "user", Content: "Reply with exactly: CONNECTION_OK"},
	}, 100)
}
