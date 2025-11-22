package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"rgs/observability"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type WebhookClient struct {
	client *http.Client
	secret string
}

func NewWebhookClient(secret string) *WebhookClient {
	return &WebhookClient{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		secret: secret,
	}
}

func (wc *WebhookClient) signPayload(timestamp string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(wc.secret))
	mac.Write([]byte(timestamp))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func (wc *WebhookClient) Send(ctx context.Context, webhookURL string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := wc.signPayload(timestamp, body)

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-RGS-Timestamp", timestamp)
	req.Header.Set("X-RGS-Signature", signature)

	resp, err := wc.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			observability.Logger.Error("failed to close webhook sent response body", zap.Error(err))
		}
	}(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("webhook delivery returned status %d", resp.StatusCode)
}
