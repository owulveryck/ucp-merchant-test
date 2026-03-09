package main

import "github.com/owulveryck/ucp-merchant-test/internal/webhook"

func resolveWebhookURL(ucpAgentHeader string) string {
	return webhook.ResolveWebhookURL(ucpAgentHeader)
}

func sendWebhookEvent(webhookURL string, event map[string]interface{}) {
	webhook.SendWebhookEvent(webhookURL, event)
}
