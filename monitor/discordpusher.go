package monitor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"matu/gosol/types"
	"net/http"
	"time"
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

	extra_risk_fields := []map[string]interface{}{}
	for _, risk := range report.Risks {
		if risk.Level == "" || risk.Name == "" || risk.Score == 0 {
			continue
		}
		extra_risk_fields = append(extra_risk_fields, map[string]interface{}{"name": "Riesgo", "value": risk.Name, "inline": true})
		extra_risk_fields = append(extra_risk_fields, map[string]interface{}{"name": "Nivel", "value": risk.Level, "inline": true})
		extra_risk_fields = append(extra_risk_fields, map[string]interface{}{"name": "Score", "value": risk.Score, "inline": true})
	}

	for _, market := range report.Markets {
		if market.MarketType == "pump_fun" {
			extra_risk_fields = append(extra_risk_fields, map[string]interface{}{"name": "Market", "value": "https://pump.fun/coin/" + report.Mint, "inline": true})
		}
		if market.MarketType == "raydium" {
			extra_risk_fields = append(extra_risk_fields, map[string]interface{}{"name": "Market", "value": "https://raydium.io/swap/?inputMint=sol&outputMint=" + report.Mint, "inline": true})
		}
	}

	var color int
	switch {
	case report.Score <= 5000:
		color = 65280 // Green
	case report.Score <= 8000:
		color = 16776960 // Yellow
	default:
		color = 16711680 // Red
	}

	payload := map[string]interface{}{
		"username":   "Meme sniper",
		"avatar_url": tokenMetaData.Image,
		"content":    "Address: " + report.Mint,
		"embeds": []map[string]interface{}{
			{
				"title":       report.TokenMeta.Symbol,
				"description": report.TokenMeta.Name,
				"url":         "https://dexscreener.com/solana/" + report.Mint,
				"color":       color,
				"fields": []map[string]interface{}{
					{"name": "Rugged", "value": fmt.Sprintf("%t", report.Rugged), "inline": true},
					{"name": "SCORE", "value": report.Score, "inline": true},
					{"name": "Verification", "value": fmt.Sprintf("%s", report.Verification), "inline": true},
					{"name": "Detectado", "value": report.DetectedAt.In(time.Local).Format("2006-01-02 15:04"), "inline": true},
					{"name": "", "value": "", "inline": true},
					{"name": "", "value": "", "inline": true},
				},
			},
		},
	}

	embeds := payload["embeds"].([]map[string]interface{})
	embeds[0]["fields"] = append(embeds[0]["fields"].([]map[string]interface{}), extra_risk_fields...)
	payload["embeds"] = embeds

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
