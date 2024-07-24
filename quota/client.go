package quota

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/MrAlias/otel-auto-demo/quota/internal"
)

var (
	ErrUnavailable  = errors.New("quota service unavailable")
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

func (c *Client) UseQuota(ctx context.Context, id int) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.burnURL(id), http.NoBody)
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

	var quota internal.Quota
	err = json.NewDecoder(resp.Body).Decode(&quota)
	if err != nil {
		return err
	}

	if quota.Remaining == 0 {
		return ErrInsufficient
	}
	return nil
}

func (c *Client) burnURL(id int) string {
	return fmt.Sprintf("%s/user/%d/burn", c.endpoint, id)
}
