package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/armon/go-socks5"
)

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestAuthFailureLogsIncludeFail2banFields(t *testing.T) {
	logSink := &lockedBuffer{}

	creds := socks5.StaticCredentials{
		"testuser": "testpass",
	}
	auth := socks5.UserPassAuthenticator{Credentials: creds}

	server, err := socks5.New(&socks5.Config{
		Logger:      log.New(logSink, "", 0),
		AuthMethods: []socks5.Authenticator{auth},
	})
	if err != nil {
		t.Fatalf("failed to create socks5 server: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	serveErr := make(chan error, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		err := server.Serve(listener)
		if err != nil && !isClosedListenerError(err) {
			serveErr <- err
		}
	}()

	t.Cleanup(func() {
		_ = listener.Close()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for socks5 server to stop")
		}

		select {
		case err := <-serveErr:
			t.Fatalf("server returned unexpected error: %v", err)
		default:
		}
	})

	sendInvalidAuthHandshake(t, listener.Addr().String(), "testuser", "wrongpass")

	waitForLogParts(t, logSink, []string{
		"auth_failed",
		"remote_ip=127.0.0.1",
		`username="testuser"`,
		"method=2",
		"reason=invalid_credentials",
	})
}

func sendInvalidAuthHandshake(t *testing.T, addr, username, password string) {
	t.Helper()

	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatalf("failed to connect to socks5 server: %v", err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("failed to set connection deadline: %v", err)
	}

	// Client greeting: SOCKS5 with username/password auth support.
	if _, err := conn.Write([]byte{0x05, 0x01, 0x02}); err != nil {
		t.Fatalf("failed to write client greeting: %v", err)
	}

	resp := make([]byte, 2)
	if _, err := io.ReadFull(conn, resp); err != nil {
		t.Fatalf("failed to read auth method selection: %v", err)
	}
	if resp[0] != 0x05 || resp[1] != 0x02 {
		t.Fatalf("unexpected method selection response: %#v", resp)
	}

	if len(username) > 255 || len(password) > 255 {
		t.Fatal("test username/password exceed RFC1929 limits")
	}

	authReq := append([]byte{0x01, byte(len(username))}, []byte(username)...)
	authReq = append(authReq, byte(len(password)))
	authReq = append(authReq, []byte(password)...)
	if _, err := conn.Write(authReq); err != nil {
		t.Fatalf("failed to write auth request: %v", err)
	}

	if _, err := io.ReadFull(conn, resp); err != nil {
		t.Fatalf("failed to read auth response: %v", err)
	}
	if resp[0] != 0x01 || resp[1] != 0x01 {
		t.Fatalf("expected auth failure response [0x01 0x01], got %#v", resp)
	}
}

func waitForLogParts(t *testing.T, sink *lockedBuffer, parts []string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for {
		logs := sink.String()
		if containsAll(logs, parts) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected log to contain all parts %q, logs:\n%s", parts, logs)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func containsAll(s string, parts []string) bool {
	for _, part := range parts {
		if !strings.Contains(s, part) {
			return false
		}
	}
	return true
}

func isClosedListenerError(err error) bool {
	return errors.Is(err, net.ErrClosed) || strings.Contains(err.Error(), "use of closed network connection")
}
