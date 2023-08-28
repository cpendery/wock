package daemon

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/cpendery/wock/hosts"
	"github.com/cpendery/wock/model"
	"github.com/cpendery/wock/pipe"
)

type Daemon struct {
	mockedHosts []model.MockedHost
	lock        sync.RWMutex
	serverHttp  http.Server
	serverHttps http.Server
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
		var mockedHosts []string
		for _, mockedHost := range d.mockedHosts {
			mockedHosts = append(mockedHosts, mockedHost.Host)
		}
		data, err := json.Marshal(mockedHosts)
		if err != nil {
			slog.Error("failed to marshal status response", slog.String("error", err.Error()))
			return
		}
		if err := d.sendMessage(
			model.Message{MsgType: model.SuccessMessage, Data: data},
			msg.ClientId,
			conn,
		); err != nil {
			slog.Error("failed to response to a status message", slog.String("clientId", msg.ClientId), slog.String("error", err.Error()))
			return
		}
	case model.MockMessage:
		slog.Debug("received mock message")
		var mockMessageData model.MockMessageData
		if err := json.Unmarshal(msg.Data, &mockMessageData); err != nil {
			slog.Error("invalid mock message", slog.String("error", err.Error()), slog.String("data", string(msg.Data)))
			return
		}
		host := strings.ToLower(strings.TrimSpace(mockMessageData.Host))

		if err := hosts.UpdateHosts(host, true); err != nil {
			slog.Error("failed to update hosts file", slog.String("error", err.Error()))
			return
		}
		slog.Debug("updated mocked hosts")
		if err := d.sendMessage(
			model.Message{MsgType: model.SuccessMessage},
			msg.ClientId,
			conn,
		); err != nil {
			slog.Error("failed to response to a mock message", slog.String("clientId", msg.ClientId), slog.String("error", err.Error()))
			return
		}
		d.mockedHosts = append(d.mockedHosts, model.MockedHost{Host: host, Directory: mockMessageData.Directory})

		slog.Debug("starting server http/s shutdowns")
		if err := d.serverHttp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown http server", slog.String("error", err.Error()))
			return
		}
		slog.Debug("begin creation of http/s servers")
		go d.httpServer()
	}
}

func (d *Daemon) httpServer() {
	defer func() {
		if r := recover(); r != nil {
			slog.Debug("recovered from panic", "r", r)
		}
	}()
	mux := http.NewServeMux()
	d.serverHttp = http.Server{
		Handler: mux,
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		host := strings.Split(r.Host, ":")[0]
		for _, mockedHost := range d.mockedHosts {
			slog.Debug(host, "mocked", mockedHost.Host, "dir", mockedHost.Directory)
			if mockedHost.Host == host {
				server := http.FileServer(http.Dir(mockedHost.Directory))
				server.ServeHTTP(w, r)
			}
		}
	})
	slog.Debug("starting new http/s server")
	if err := d.serverHttp.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			slog.Error("http server failed", slog.String("error", err.Error()))
		} else {
			slog.Debug("http server shutdown", slog.String("error", err.Error()))
		}
	}
}

func (d *Daemon) handleClient(c net.Conn) {
	defer c.Close()
	reader := bufio.NewReader(c)

	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				slog.Error("failed to read message", slog.String("error", err.Error()))
			}
			break
		}
		var msg model.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			slog.Error("failed to unmarshal message", slog.String("error", err.Error()))
		}
		d.handleMessage(msg)
	}
}

func NewDaemon() *Daemon {
	logFile, err := os.Create("")
	if err != nil {
		fmt.Println(err.Error())
	}
	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	return &Daemon{
		mockedHosts: []model.MockedHost{},
		lock:        sync.RWMutex{},
		serverHttp: http.Server{
			Addr: ":80",
		},
		serverHttps: http.Server{
			Addr: ":443",
		},
	}
}

func (d *Daemon) Start() {
	slog.Debug("starting daemon")
	l, err := pipe.ServerListen()
	if err != nil {
		slog.Error("failed to listen to daemon pipe", slog.String("error", err.Error()))
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			slog.Error("failed to accept new connection", slog.String("error", err.Error()))
		}
		go d.handleClient(conn)
	}
}
