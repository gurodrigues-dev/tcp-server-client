package client

import "github.com/gurodrigues-dev/tcp-server-client/server"

func (c *TCPClient) Authorize(username string) (*server.Response, error) {
	authParams := &server.AuthParams{
		Username: username,
	}

	req := &server.Request{
		Method: "authorize",
		Params: authParams,
	}

	response, err := c.SendRequestSync(req)
	if err != nil {
		return nil, err
	}

	if response.Error == "" {
		c.username = username
	}

	return response, nil
}
