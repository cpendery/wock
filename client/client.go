package client

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/cpendery/wock/model"
	"github.com/cpendery/wock/pipe"
	"github.com/google/uuid"
)

const (
	clientTimeout = 5 * time.Second
)

var (
	ErrUnableToDialDaemon = errors.New("unable to dial daemon")
)

type Client struct {
	server   net.Conn
	incoming net.Listener
	clientId string
	received chan model.Message
}

func NewClient() (*Client, error) {
	clientId := uuid.NewString()
	serverConnection, err := pipe.DialServer()
	if err != nil {
		slog.Debug("unable to dial daemon", slog.String("error", err.Error()))
		return nil, ErrUnableToDialDaemon
	}
	incomingListener, err := pipe.ClientListen(clientId)
	if err != nil {
		slog.Debug("unable to listen to client pipe", slog.String("error", err.Error()))
		return nil, errors.New("unable to listen to client pipe")
	}
	client := Client{
		server:   serverConnection,
		incoming: incomingListener,
		clientId: clientId,
		received: make(chan model.Message, 2),
	}
	go client.readIncomingMessages()
	return &client, nil
}

func (c *Client) Close() error {
	if err := c.server.Close(); err != nil {
		return fmt.Errorf("failed to close daemon pipe: %w", err)
	}
	if err := c.incoming.Close(); err != nil {
		return fmt.Errorf("failed to close daemon pipe: %w", err)
	}
	return nil
}

func (c *Client) SendMessage(msgType model.MessageType, data []byte) error {
	msg, err := json.Marshal(&model.Message{MsgType: msgType, Data: data, ClientId: c.clientId})
	if err != nil {
		return fmt.Errorf("unable to marshal message: %w", err)
	}
	if _, err := c.server.Write(msg); err != nil {
		return fmt.Errorf("unable to write message %s: %w", string(msg), err)
	}
	if _, err := c.server.Write([]byte("\n")); err != nil {
		return fmt.Errorf("unable to write message delimiter: %w", err)
	}
	go func() {
		time.Sleep(clientTimeout)
		c.received <- model.Message{MsgType: model.TimeoutMessage}
	}()
	return nil
}

func (c *Client) readIncomingMessages() {
	conn, err := c.incoming.Accept()
	if err != nil {
		slog.Error("failed to accept incoming connection", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				slog.Error("failed to read incoming message", slog.String("error", err.Error()))
			}
			break
		}
		var msg model.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			slog.Error("failed to unmarshal message", slog.String("error", err.Error()), slog.String("message", string(data)))
		}
		c.received <- msg
	}
}

func (c *Client) CheckStatus() (*[]model.MockedHost, error) {
	if err := c.SendMessage(model.StatusMessage, []byte{}); err != nil {
		return nil, fmt.Errorf("unable to send status message: %w", err)
	}
	resp := <-c.received

	switch resp.MsgType {
	case model.SuccessMessage:
		var hosts []model.MockedHost
		err := json.Unmarshal(resp.Data, &hosts)
		if err != nil {
			return nil, fmt.Errorf("unable to read status response: %w", err)
		}
		return &hosts, nil
	default:
		return nil, errors.New("status request failed")
	}
}

func (c *Client) Mock(host string, directory string) error {
	mockMessage, err := json.Marshal(model.MockMessageData{Host: host, Directory: directory})
	if err != nil {
		return fmt.Errorf("unable to create mock message: %w", err)
	}
	if err := c.SendMessage(model.MockMessage, mockMessage); err != nil {
		return fmt.Errorf("unable to send mock message: %w", err)
	}
	resp := <-c.received

	switch resp.MsgType {
	case model.SuccessMessage:
		return nil
	default:
		return errors.New("mock request failed")
	}
}

func (c *Client) Clear() error {
	if err := c.SendMessage(model.ClearMessage, []byte{}); err != nil {
		return fmt.Errorf("unable to send clear message: %w", err)
	}
	resp := <-c.received

	switch resp.MsgType {
	case model.SuccessMessage:
		return nil
	default:
		return errors.New("clear request failed")
	}
}

func (c *Client) Stop() error {
	if err := c.SendMessage(model.StopMessage, []byte{}); err != nil {
		return fmt.Errorf("unable to send stop message: %w", err)
	}
	resp := <-c.received

	switch resp.MsgType {
	case model.SuccessMessage:
		return nil
	default:
		return errors.New("stop request failed")
	}
}

func (c *Client) Remove(host string) error {
	if err := c.SendMessage(model.UnmockMessage, []byte(host)); err != nil {
		return fmt.Errorf("unable to send remove message: %w", err)
	}
	resp := <-c.received

	switch resp.MsgType {
	case model.SuccessMessage:
		return nil
	case model.ErrorMessage:
		return errors.New(string(resp.Data))
	default:
		return errors.New("remove request failed")
	}
}
