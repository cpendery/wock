package daemon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"sync"

	"github.com/cpendery/wock/hosts"
	"github.com/cpendery/wock/model"
	"github.com/cpendery/wock/pipe"
)

type Daemon struct {
	mockedHosts []string
	lock        sync.RWMutex
}

func (d *Daemon) sendMessage(msg model.Message, clientId string, conn net.Conn) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("unable to marshal message: %w", err)
	}
	conn.Write(data)
	conn.Write([]byte("\n"))
	return nil
}

func (d *Daemon) handleMessage(msg model.Message) {
	d.lock.Lock()
	defer d.lock.Unlock()
	conn, err := pipe.DialClient(msg.ClientId)
	if err != nil {
		slog.Error("failed to dial client", slog.String("clientId", msg.ClientId), slog.String("error", err.Error()))
		return
	}
	defer conn.Close()
	switch msg.MsgType {
	case model.StatusMessage:
		data, _ := json.Marshal(d.mockedHosts)
		if err := d.sendMessage(
			model.Message{MsgType: model.SuccessMessage, Data: data},
			msg.ClientId,
			conn,
		); err != nil {
			slog.Error("failed to response to a status message", slog.String("clientId", msg.ClientId), slog.String("error", err.Error()))
		}
	case model.MockMessage:
		host := strings.ToLower(strings.TrimSpace(string(msg.Data)))
		err := hosts.UpdateHosts(host, true)
		if err != nil {
			slog.Error("failed to update hosts file", slog.String("error", err.Error()))
		}
		if err := d.sendMessage(
			model.Message{MsgType: model.SuccessMessage},
			msg.ClientId,
			conn,
		); err != nil {
			slog.Error("failed to response to a mock message", slog.String("clientId", msg.ClientId), slog.String("error", err.Error()))
		}
		d.mockedHosts = append(d.mockedHosts, host)
	}
}

func (d *Daemon) handleClient(c net.Conn) {
	defer c.Close()
	reader := bufio.NewReader(c)

	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				// log error here
			}
			break
		}
		var msg model.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			// log error here
		}
		d.handleMessage(msg)
	}
}

func NewDaemon() *Daemon {
	return &Daemon{
		mockedHosts: []string{},
		lock:        sync.RWMutex{},
	}
}

func (d *Daemon) Start() {
	l, err := pipe.ServerListen()
	if err != nil {
		fmt.Println(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go d.handleClient(conn)
	}
}
