package monitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gosol/types"
	"net/http"
)

func PushToDiscord(report types.Report, statusUpdates chan<- StatusMessage) {
	// Fetch the image URL from report.TokenMeta.uri
	resp, err := http.Get(report.TokenMeta.URI)
	if err != nil {
		statusUpdates <- StatusMessage{Level: ERR, Message: fmt.Sprintf("Error fetching image URL: %v", err)}
		// Handle error
		return
	}
	defer resp.Body.Close()

	var tokenMetaData struct {
		Image string `json:"image"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenMetaData); err != nil {
		statusUpdates <- StatusMessage{Level: ERR, Message: fmt.Sprintf("Error decoding image URL: %v", err)}
		return
	}

	payload := map[string]interface{}{
		"username":   "Report Bot",
		"avatar_url": tokenMetaData.Image,
		"content":    report.TokenMeta.Name,
		"embeds": []map[string]interface{}{
			{
				"title":       "Report Details",
				"description": "Details of the report for " + report.TokenMeta.Name,
				"url":         "https://api.rugcheck.xyz/v1/tokens/" + report.Mint,
				"color":       15258703,
				"fields": []map[string]interface{}{
					{"name": "Total Market Liquidity", "value": fmt.Sprintf("%f", report.TotalMarketLiquidity), "inline": true},
					{"name": "Total LP Providers", "value": fmt.Sprintf("%d", report.TotalLPProviders), "inline": true},
					{"name": "Rugged", "value": fmt.Sprintf("%t", report.Rugged), "inline": true},
					{"name": "Verification", "value": report.Verification, "inline": true},
					// Add more fields as necessary
				},
			},
		},
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		// Handle error
		return
	}

	// Send the POST request to Discord webhook
	_, err = http.Post("https://discord.com/api/webhooks/1311165446115823666/hFBba1KUKbM8dtr99i30dZSl8ma1pf815hVWuiCbw7PTdgG38EMliYpQKuMXn0jiUXb8", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		statusUpdates <- StatusMessage{Level: ERR, Message: fmt.Sprintf("Error sending webhook: %v", err)}
		return
	}

	/*
		// webhook example to push to discord

		curl -H "Content-Type: application/json" \
		-d '{"content": "Hello, this is a message from my webhook!"}' \
		-X POST https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN


		// example of json payload structure

		{
			"username": "Webhook Name",
			"avatar_url": "https://example.com/avatar.png",
			"content": "This is a plain text message.",
			"embeds": [
				{
				"title": "Embed Title",
				"description": "Description of the embed.",
				"url": "https://example.com",
				"color": 15258703,
				"fields": [
					{
					"name": "Field Name",
					"value": "Field Value",
					"inline": true
					}
				]
				}
			]
		}
	*/

}
