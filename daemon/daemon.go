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
	"path/filepath"
	"strings"
	"sync"

	"github.com/cpendery/wock/cert"
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
		data, err := json.Marshal(d.mockedHosts)
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
		if err := d.serverHttps.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown https server", slog.String("error", err.Error()))
			return
		}
		slog.Debug("begin creation of http/s servers")
		if err := d.generateNewCerts(); err != nil {
			slog.Error("failed to generate new certs", slog.String("error", err.Error()))
			os.Exit(1)
		}
		go d.httpServer()
		go d.httpsServer()
	case model.ClearMessage:
		slog.Debug("received clear message")
		hosts.ClearHosts()
		d.mockedHosts = []model.MockedHost{}
		if err := d.sendMessage(
			model.Message{MsgType: model.SuccessMessage},
			msg.ClientId,
			conn,
		); err != nil {
			slog.Error("failed to response to a clear message", slog.String("clientId", msg.ClientId), slog.String("error", err.Error()))
			return
		}
	case model.StopMessage:
		slog.Debug("received stop message")
		if err := d.sendMessage(
			model.Message{MsgType: model.SuccessMessage},
			msg.ClientId,
			conn,
		); err != nil {
			slog.Error("failed to response to a stop message", slog.String("clientId", msg.ClientId), slog.String("error", err.Error()))
		}
		os.Exit(0)
	}
}

func create(p string) error {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func (d *Daemon) generateNewCerts() error {
	var hosts []string
	for _, mockedHost := range d.mockedHosts {
		hosts = append(hosts, mockedHost.Host)
	}
	if err := create(cert.WockCertFile); err != nil {
		return fmt.Errorf("failed to create cert file: %w", err)
	}
	if err := create(cert.WockKeyFile); err != nil {
		return fmt.Errorf("failed to create key file: %w", err)
	}
	return cert.CreateCert(hosts)
}

func (d *Daemon) httpsServer() {
	mux := http.NewServeMux()
	d.serverHttps = http.Server{
		Handler: mux,
		Addr:    ":443",
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
	slog.Debug("starting new https server")
	if err := d.serverHttps.ListenAndServeTLS(cert.WockCertFile, cert.WockKeyFile); err != nil {
		if err != http.ErrServerClosed {
			slog.Error("https server failed", slog.String("error", err.Error()))
		} else {
			slog.Debug("https server shutdown", slog.String("error", err.Error()))
		}
	}
}

func (d *Daemon) httpServer() {
	mux := http.NewServeMux()
	d.serverHttp = http.Server{
		Handler: mux,
		Addr:    ":80",
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
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
