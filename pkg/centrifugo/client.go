package centrifugo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type Client struct {
	url        string
	apiKey     string
	httpClient *http.Client
	logger     *slog.Logger
}

func NewClient(cfg *Config, logger *slog.Logger) *Client {
	return &Client{
		url:        cfg.URL,
		apiKey:     cfg.APIKey,
		httpClient: &http.Client{},
		logger:     logger,
	}
}

type PublishRequest struct {
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

type PublishResponse struct {
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) Publish(ctx context.Context, channel string, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req := PublishRequest{
		Channel: channel,
		Data:    dataBytes,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.url+"/api/publish", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "apikey "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("centrifugo returned status %d", resp.StatusCode)
	}

	var publishResp PublishResponse
	if err := json.NewDecoder(resp.Body).Decode(&publishResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if publishResp.Error != nil {
		return fmt.Errorf("centrifugo error %d: %s", publishResp.Error.Code, publishResp.Error.Message)
	}

	return nil
}

func (c *Client) PublishToUsers(ctx context.Context, userIDs []string, channelPrefix string, data interface{}) error {
	for _, userID := range userIDs {
		channel := fmt.Sprintf("%s#%s", channelPrefix, userID)
		if err := c.Publish(ctx, channel, data); err != nil {
			c.logger.ErrorContext(ctx, "failed to publish to user channel",
				"user_id", userID,
				"channel", channel,
				"error", err,
			)
			return fmt.Errorf("failed to publish to channel %s: %w", channel, err)
		}
	}
	return nil
}
