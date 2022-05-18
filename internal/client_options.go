package internal

import "net/http"

type ClientOptionFunc func(*Client) error

func WithBaseURL(urlStr string) ClientOptionFunc {
	return func(c *Client) error {
		return c.setBaseUrl(urlStr)
	}
}

func WithHTTPClient(client *http.Client) ClientOptionFunc {
	return func(c *Client) error {
		c.client = client
		return nil
	}
}
