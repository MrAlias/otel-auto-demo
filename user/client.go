package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/MrAlias/otel-auto-demo/user/internal"
)

var (
	ErrUnavailable  = errors.New("user service unavailable")
	ErrInsufficient = errors.New("insufficient quota")
)

type Client struct {
	endpoint string
	client   *http.Client
}

func NewClient(c *http.Client, endpoint string) *Client {
	endpoint = strings.TrimRight(endpoint, "/")
	return &Client{endpoint: endpoint, client: c}
}

func (c *Client) get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func (c *Client) checkStatus(resp *http.Response) error {
	switch {
	case resp.StatusCode >= http.StatusInternalServerError:
		return ErrUnavailable
	case resp.StatusCode >= http.StatusBadRequest:
		return fmt.Errorf("bad request: %s", http.StatusText(resp.StatusCode))
	default:
		return nil
	}
}

func (c *Client) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/healthcheck", c.endpoint)
	resp, err := c.get(ctx, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return c.checkStatus(resp)
}

func (c *Client) UseQuota(ctx context.Context, name string) error {
	url := fmt.Sprintf("%s/user/%s/alloc", c.endpoint, name)
	resp, err := c.get(ctx, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.checkStatus(resp); err != nil {
		return err
	}

	var user internal.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return err
	}

	if user.Quota == 0 {
		return ErrInsufficient
	}
	return nil
}
