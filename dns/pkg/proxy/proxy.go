package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	endpoint string
	http     *http.Client
}

func NewDoHClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		http: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     90 * time.Second,
				ForceAttemptHTTP2:   true,
			},
		},
	}
}

// SendQuery forwards a raw DNS query to the DoH endpoint (RFC 8484) and returns
// the raw DNS response. It retries once so a stale keep-alive connection can't
// fail an otherwise-healthy query.
func (c *Client) SendQuery(query []byte) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		resp, err := c.exchange(query)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("doh query failed: %w", lastErr)
}

func (c *Client) exchange(query []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, c.endpoint, bytes.NewReader(query))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 65535))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return body, nil
}
