package discache

import (
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseUrl string
	client  *http.Client
}

func NewClient(baseUrl string) *Client {
	return &Client{
		baseUrl: baseUrl,
		client:  &http.Client{Timeout: time.Second * 5},
	}
}

func (c *Client) Get(key string) ([]byte, error) {
	url := c.baseUrl + "/get?key=" + key
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *Client) Add(key string, value []byte) error {
	url := c.baseUrl + "/add?key=" + key + "&value=" + string(value)
	resp, err := c.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
