package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var errNoSubscriptions = errors.New("no subscriptions")
var errUpstream = errors.New("upstream unavailable")

type botClient struct {
	url              string
	key              string
	whitelistBaseURL string
	basicBaseURL     string
	http             *http.Client
}

func newBotClient(url, key string, timeout time.Duration, whitelistBaseURL, basicBaseURL string) *botClient {
	return &botClient{url: url, key: key, http: &http.Client{Timeout: timeout}, whitelistBaseURL: strings.TrimRight(whitelistBaseURL, "/"), basicBaseURL: strings.TrimRight(basicBaseURL, "/")}
}

type botKey struct {
	ShortUUID      string `json:"shortUuid"`
	ShortUUIDSnake string `json:"short_uuid"`
	Type           string `json:"type"`
	Kind           string `json:"kind"`
}

func (k botKey) uuid() string {
	if k.ShortUUID != "" {
		return k.ShortUUID
	}
	return k.ShortUUIDSnake
}

func (k botKey) kind() string {
	if k.Type != "" {
		return k.Type
	}
	return k.Kind
}

type botSubscriptionsResponse struct {
	Keys          []botKey `json:"keys"`
	Subscriptions []botKey `json:"subscriptions"`
}

func (r botSubscriptionsResponse) all() []botKey {
	if len(r.Keys) > 0 {
		return r.Keys
	}
	return r.Subscriptions
}

type subscription struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	URL   string `json:"subscription_url"`
}

func mapSubscriptions(keys []botKey, whitelistBaseURL, basicBaseURL string) []subscription {
	subs := make([]subscription, 0, len(keys))
	for _, k := range keys {
		uuid := k.uuid()
		if uuid == "" {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(k.kind())) {
		case "_wh", "wh", "whitelist":
			subs = append(subs, subscription{ID: "_wh", Label: "Обход", URL: whitelistBaseURL + "/" + uuid})
		case "_basic", "basic", "":
			subs = append(subs, subscription{ID: "_basic", Label: "Базовый", URL: basicBaseURL + "/" + uuid})
		default:
			continue
		}
	}
	return subs
}

func maskUUID(u string) string {
	if len(u) <= 8 {
		return "****"
	}
	return u[:4] + "***" + u[len(u)-2:]
}

func (c *botClient) fetchSubscriptions(ctx context.Context, telegramID int64) ([]subscription, error) {
	body, err := json.Marshal(map[string]any{"telegram_id": telegramID})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errUpstream, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errUpstream, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", c.key)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errUpstream, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errNoSubscriptions
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", errUpstream, resp.StatusCode)
	}
	var decoded botSubscriptionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("%w: %v", errUpstream, err)
	}
	keys := decoded.all()
	if len(keys) == 0 {
		return nil, errNoSubscriptions
	}
	subs := mapSubscriptions(keys, c.whitelistBaseURL, c.basicBaseURL)
	if len(subs) == 0 {
		return nil, errNoSubscriptions
	}
	return subs, nil
}
