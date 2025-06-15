package infrastructure

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type DiscordNotifier struct {
	WebhookURL string
}

func (n *DiscordNotifier) Send(message string) error {
	payload := map[string]string{"content": message}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(n.WebhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
