package ai

import (
	"context"
	"encoding/base64"
	"fmt"
)

// AnalyzeItemVisionError returns the vision error if any, for diagnostics.
func AnalyzeItemVisionError(ctx context.Context, client *Client, imagePath string) error {
	if client == nil {
		return nil
	}
	_, err := analyzeWithVision(ctx, client, imagePath)
	return err
}

// AnalyzeItemVisionRaw returns raw model text for debugging.
func AnalyzeItemVisionRaw(ctx context.Context, client *Client, imagePath string) (string, error) {
	data, mime, err := prepareImageForVision(imagePath)
	if err != nil {
		return "", err
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mime, b64)
	content := []map[string]any{
		{"type": "text", "text": analyzePrompt},
		{"type": "image_url", "image_url": map[string]string{"url": dataURL}},
	}
	return client.VisionCompletion(ctx, []chatMessage{
		{Role: "user", Content: content},
	}, 2048)
}
