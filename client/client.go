package client

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

type ClientConfig struct {
	dialInterval time.Duration
	dialAttempts int
	remoteAddr   string
}

func NewClientConfig(params ...any) (*ClientConfig, error) {
	// Check the number of parameters
	if len(params) != 3 {
		msg := fmt.Sprintf("error, params length mismatch; expected 3, found: %d", len(params))
		err := errors.New(msg)
		return nil, err
	}

	// Format the configuration type
	config := new(ClientConfig)
	for _, param := range params {

		switch t := param.(type) {
		case time.Duration:
			config.dialInterval = t
		case int:
			config.dialAttempts = t
		case string:
			config.remoteAddr = t
		default:
			msg := fmt.Sprintf("error, invalid type: %s", t)
			err := errors.New(msg)
			return nil, err
		}
	}
	return config, nil
}

type Client struct {
	remoteAddr   string
	dialInterval time.Duration
	dialAttepmts int
	conn         net.Conn

	// client also has a file reader that reads the files and
	// pipes the messages to the client
	msgch   chan string
	closech chan struct{}
}

func NewClient(c *ClientConfig) *Client {
	return &Client{
		remoteAddr:   c.remoteAddr,
		dialInterval: c.dialInterval,
		dialAttepmts: c.dialAttempts,
		msgch:        make(chan string),
		closech:      make(chan struct{}),
	}
}

func (c *Client) Run() error {
	// Dial the server
	if err := c.dial(); err != nil {
		return err
	}
	// Wait for handshake
	if err := c.handleHandshake(); err != nil {
		logrus.Errorf("handshake error: %v", err)
	}

	go c.handleRead()
	go c.handleWrite()

	for {
		select {
		case msg := <-c.msgch:
			log.Println(msg)
		case <-c.closech:
			return c.quit()
		}
	}
}

func (c *Client) dial() error {
	dialer := &net.Dialer{
		Timeout:   2 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	var attempts int
	for {
		// Check the number of missed dialed attempts
		if attempts > c.dialAttepmts {
			msg := fmt.Sprintf("error, to many dial attempts: %d", attempts)
			err := errors.New(msg)
			return err
		}

		conn, err := dialer.Dial("tcp", c.remoteAddr)
		if err != nil {
			log.Printf("dial error (%s), retrying in (%d):", err, c.dialInterval)
			time.Sleep(c.dialInterval)
			attempts += 1
			continue
		}

		c.conn = conn
		break
	}
	return nil
}

func (c *Client) handleHandshake() error {
	buf := make([]byte, 1024)
	n, err := c.conn.Read(buf)
	if err != nil {
		return err
	}

	msg := &MessageHandshakeRequest{}

	r := bytes.NewBuffer(buf[:n])
	if err := binary.Read(r, binary.LittleEndian, msg); err != nil {
		return err
	}

	// We have the handshake message, interpret it and send
	// back the response
	if msg.ProtocolVersion != PROTOCOL_VERSION {
		errStr := fmt.Sprintf("handshake error, expected protocol version: %v", PROTOCOL_VERSION)
		return errors.New(errStr)
	}

	if msg.HandshakeRequestByte != HANDSHAKE_REQUEST_BYTE {
		return errors.New("handshake error, invalid request byte")
	}

	// Write response back to the server
	respMsg := MessageHandshakeResponse{
		ProtocolVersion:       1,
		HandshakeResponseByte: 0x06,
	}
	respBuf := new(bytes.Buffer)

	binary.Write(respBuf, binary.LittleEndian, respMsg)
	if _, err := c.conn.Write(respBuf.Bytes()); err != nil {
		return err
	}

	return nil
}

func (c *Client) handleRead() {
	defer c.conn.Close()
	buf := make([]byte, 2048)

	for {
		n, err := c.conn.Read(buf)
		if err == io.EOF {
			// lost connection to the server
			log.Println(err)
			return
		}

		c.msgch <- string(buf[:n])
	}
}

func (c *Client) handleWrite() {
	tick := time.NewTicker(2 * time.Second)
	for range tick.C {
		c.write([]byte("Hello"))
	}
}

func (c *Client) write(b []byte) error {
	// Read the data in memory
	if _, err := c.conn.Write(b); err != nil {
		return err
	}
	return nil
}

func (c *Client) quit() error {
	close(c.msgch)
	return c.conn.Close()
}
