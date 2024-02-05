package ctrlsock

import (
	"net"
)

type Client struct {
	conn net.Conn
}

func NewClient(sock, authKey string) (*Client, error) {
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil, err
	}

	c := &Client{conn: conn}

	// Send auth key.
	if err := writeString(c.conn, authKey); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return c, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) SetEnv(key, value string) error {
	return writeSetEnv(c.conn, key, value)
}
