package mcp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// fetchAndEncodeImage fetches an image URL and returns its base64-encoded data and MIME type.
func fetchAndEncodeImage(url string) (string, string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	return base64.StdEncoding.EncodeToString(body), mimeType, nil
}

// extractImageURLs walks a result structure and returns all image_url values found.
func extractImageURLs(result interface{}) []string {
	data, err := json.Marshal(result)
	if err != nil {
		return nil
	}

	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}

	seen := map[string]bool{}
	var urls []string
	var walk func(v interface{})
	walk = func(v interface{}) {
		switch val := v.(type) {
		case map[string]interface{}:
			if u, ok := val["image_url"].(string); ok && u != "" && !seen[u] {
				seen[u] = true
				urls = append(urls, u)
			}
			for _, child := range val {
				walk(child)
			}
		case []interface{}:
			for _, child := range val {
				walk(child)
			}
		}
	}
	walk(raw)
	return urls
}
