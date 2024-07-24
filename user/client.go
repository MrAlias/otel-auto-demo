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

var ErrUnavailable = errors.New("user service unavailable")

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

func (c *Client) AllIDs(ctx context.Context) ([]int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.usersURL(), http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var users []internal.User
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		return nil, err
	}

	var out []int
	for _, user := range users {
		out = append(out, user.ID)
	}
	return out, nil
}

func (c *Client) usersURL() string {
	return fmt.Sprintf("%s/users", c.endpoint)
}

func (c *Client) UserID(ctx context.Context, name string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.userURL(name), http.NoBody)
	if err != nil {
		return 0, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, ErrUnavailable
	}

	var user internal.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (c *Client) userURL(username string) string {
	return fmt.Sprintf("%s/users/%s", c.endpoint, username)
}
