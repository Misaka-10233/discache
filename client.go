package discache

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Client struct {
	ctx       context.Context
	client    *http.Client
	serverMap *ServerMap
}

func NewClient(nodeReplica int, dataReplica int, url ...string) *Client {
	return &Client{
		ctx:       context.Background(),
		client:    &http.Client{},
		serverMap: NewServerMap(nodeReplica, dataReplica, url...),
	}
}

var (
	ERR_NO_AVAILABLES_SERVER = errors.New("no available server")
	ERR_GET_FAILED           = errors.New("get failed")
	ERR_GET_TIMEOUT          = errors.New("get timeout")
)

func (c *Client) Get(key string) ([]byte, error) {
	urls := c.serverMap.GetServerUrls(key)
	count := int32(len(urls))
	if count == 0 {
		return nil, ERR_NO_AVAILABLES_SERVER
	}

	ctx, cancel := context.WithTimeout(c.ctx, time.Second*3)
	defer cancel()
	result := make(chan []byte)
	once := sync.Once{}
	for _, url := range urls {
		go func(u string) {
			resp, err := c.doGet(ctx, u+"/get", key)
			if err == nil {
				once.Do(func() {
					result <- resp
					cancel()
				})
			} else {
				atomic.AddInt32(&count, -1)
				if atomic.LoadInt32(&count) == 0 {
					once.Do(func() {
						close(result)
					})
				}
			}
		}(url)
	}

	select {
	case res, ok := <-result:
		if !ok {
			return nil, ERR_GET_FAILED
		}
		return res, nil
	case <-ctx.Done():
		return nil, ERR_GET_TIMEOUT
	}
}

func (c *Client) doGet(ctx context.Context, url string, key string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("key", key)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, ERR_GET_FAILED
	}
	return io.ReadAll(resp.Body)
}
