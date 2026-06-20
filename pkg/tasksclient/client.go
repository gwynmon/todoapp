package tasksclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"todoapp/internal/entity"
)

type Client struct {
	baseURL string
	secret  string
	http    *http.Client
}

func New(baseURL, secret string) *Client {
	return &Client{
		baseURL: baseURL,
		secret:  secret,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) GetUpcomingDeadlines(ctx context.Context, within time.Duration) ([]entity.Task, error) {
	url := fmt.Sprintf("%s/internal/tasks/upcoming-deadlines?within=%s", c.baseURL, within.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Internal-Secret", c.secret)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tasks-service returned status %d", resp.StatusCode)
	}

	var tasks []entity.Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}
