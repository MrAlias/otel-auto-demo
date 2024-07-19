package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var (
	ErrUnavailable  = errors.New("authorization service unavailable")
	ErrUnauthorized = errors.New("user not authorized")
)

type Client struct {
	endpoint string
	client   *http.Client
}

func NewClient(c *http.Client, endpoint string) *Client {
	endpoint = strings.TrimRight(endpoint, "/")
	return &Client{endpoint: endpoint, client: c}
}

func (c *Client) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.healthcheckURL(), http.NoBody)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrUnavailable
	}
	return nil
}

func (c *Client) healthcheckURL() string {
	return fmt.Sprintf("%s/healthcheck", c.endpoint)
}

func (c *Client) Check(ctx context.Context, user string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.userURL(user), http.NoBody)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusOK:
		return nil
	default:
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}
		return fmt.Errorf("authorization check error: %s", string(b))
	}
}

func (c *Client) userURL(name string) string {
	return fmt.Sprintf("%s/auth/%s", c.endpoint, name)
}
