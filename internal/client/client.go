package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	Host  string
	Token string
	HTTP  *http.Client
}

func New(host, token string) *Client {
	return &Client{
		Host:  strings.TrimRight(host, "/"),
		Token: token,
		HTTP:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Get(path string, query url.Values) (json.RawMessage, error) {
	return c.do(http.MethodGet, path, query, nil)
}

func (c *Client) Post(path string, body any) (json.RawMessage, error) {
	return c.do(http.MethodPost, path, nil, body)
}

func (c *Client) Patch(path string, body any) (json.RawMessage, error) {
	return c.do(http.MethodPatch, path, nil, body)
}

func (c *Client) do(method, path string, query url.Values, body any) (json.RawMessage, error) {
	if c.Host == "" || c.Token == "" {
		return nil, fmt.Errorf("not logged in: run vt auth login --host https://YOUR_HOST")
	}
	u, err := url.Parse(c.Host + path)
	if err != nil {
		return nil, err
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}
	var rdr io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rdr = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, u.String(), rdr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	return json.RawMessage(payload), nil
}
