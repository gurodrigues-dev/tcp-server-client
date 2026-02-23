package client

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/gurodrigues-dev/tcp-server-client/server"
)

const (
	protocol = "tcp"
)

type TCPClient struct {
	conn            net.Conn
	decoder         *json.Decoder
	encoder         *json.Encoder
	address         string
	username        string
	serverNonce     string
	jobID           int
	mu              sync.Mutex
	rateLimiter     *RateLimiter
	messageChan     chan interface{}
	responseChan    chan *server.Response
	pendingRequests map[int]chan *server.Response
	pendingMu       sync.Mutex
	nextRequestID   int
	listenerRunning bool
	listenerDone    chan struct{}
}

func NewTCPClient(address string) *TCPClient {
	return &TCPClient{
		address:         address,
		rateLimiter:     NewRateLimiter(),
		messageChan:     make(chan interface{}, 100),
		responseChan:    make(chan *server.Response, 100),
		pendingRequests: make(map[int]chan *server.Response),
		nextRequestID:   1,
		listenerDone:    make(chan struct{}),
	}
}

func (c *TCPClient) Connect() error {
	conn, err := net.Dial(protocol, c.address)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	c.conn = conn
	c.decoder = json.NewDecoder(conn)
	c.encoder = json.NewEncoder(conn)

	c.startMessageListener()

	return nil
}

func (c *TCPClient) Disconnect() {
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
		c.decoder = nil
		c.encoder = nil
	}

	close(c.messageChan)
	close(c.responseChan)
	<-c.listenerDone
}

func (c *TCPClient) sendRequest(req *server.Request) error {
	if c.encoder == nil {
		return fmt.Errorf("client is not connected")
	}

	return c.encoder.Encode(req)
}

func (c *TCPClient) SendRequestSync(req *server.Request) (*server.Response, error) {
	if c.encoder == nil {
		return nil, fmt.Errorf("client is not connected")
	}

	if req.ID == 0 {
		c.mu.Lock()
		req.ID = c.nextRequestID
		c.nextRequestID++
		c.mu.Unlock()
	}

	responseChan := make(chan *server.Response, 1)

	c.pendingMu.Lock()
	c.pendingRequests[req.ID] = responseChan
	c.pendingMu.Unlock()

	defer func() {
		c.pendingMu.Lock()
		delete(c.pendingRequests, req.ID)
		c.pendingMu.Unlock()
		close(responseChan)
	}()

	if err := c.sendRequest(req); err != nil {
		return nil, err
	}

	select {
	case response := <-responseChan:
		return response, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("request timeout after 30 seconds")
	}
}

func (c *TCPClient) startMessageListener() {
	if c.decoder == nil {
		return
	}

	c.listenerRunning = true

	go func() {
		defer close(c.listenerDone)

		for {
			var msg json.RawMessage
			err := c.decoder.Decode(&msg)
			if err != nil {
				return
			}

			var response server.Response
			if err := json.Unmarshal(msg, &response); err == nil && response.ID != 0 {
				c.pendingMu.Lock()
				if ch, ok := c.pendingRequests[response.ID]; ok {
					ch <- &response
				}
				c.pendingMu.Unlock()
				continue
			}

			var request server.Request
			if err := json.Unmarshal(msg, &request); err == nil && request.Method != "" {
				select {
				case c.messageChan <- request.Params:
				default:
				}
				continue
			}
		}
	}()
}
