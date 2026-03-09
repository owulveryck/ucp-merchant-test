package webhook

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// ResolveWebhookURL extracts webhook_url from the agent profile referenced in the UCP-Agent header.
func ResolveWebhookURL(ucpAgentHeader string) string {
	profileURL := ""
	for _, part := range strings.Split(ucpAgentHeader, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "profile=") {
			profileURL = strings.Trim(strings.TrimPrefix(part, "profile="), "\"")
		}
	}
	if profileURL == "" {
		return ""
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(profileURL)
	if err != nil {
		log.Printf("webhook: failed to fetch agent profile %s: %v", profileURL, err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var profile struct {
		UCP struct {
			Capabilities []struct {
				Name   string `json:"name"`
				Config struct {
					WebhookURL string `json:"webhook_url"`
				} `json:"config"`
			} `json:"capabilities"`
		} `json:"ucp"`
	}
	if err := json.Unmarshal(body, &profile); err != nil {
		log.Printf("webhook: failed to parse agent profile: %v", err)
		return ""
	}

	for _, cap := range profile.UCP.Capabilities {
		if cap.Config.WebhookURL != "" {
			return cap.Config.WebhookURL
		}
	}
	return ""
}

// SendWebhookEvent sends a webhook event to the given URL.
func SendWebhookEvent(webhookURL string, event map[string]interface{}) {
	if webhookURL == "" {
		return
	}
	go func() {
		body, err := json.Marshal(event)
		if err != nil {
			log.Printf("webhook: marshal error: %v", err)
			return
		}
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(webhookURL, "application/json", bytes.NewReader(body))
		if err != nil {
			log.Printf("webhook: POST to %s failed: %v", webhookURL, err)
			return
		}
		resp.Body.Close()
		log.Printf("webhook: sent %s to %s (status %d)", event["event_type"], webhookURL, resp.StatusCode)
	}()
}
