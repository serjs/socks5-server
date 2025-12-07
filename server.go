package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/armon/go-socks5"
	"github.com/caarlos0/env/v6"
)

type params struct {
	User            string   `env:"PROXY_USER" envDefault:""`
	Password        string   `env:"PROXY_PASSWORD" envDefault:""`
	Port            string   `env:"PROXY_PORT" envDefault:"1080"`
	AllowedDestFqdn string   `env:"ALLOWED_DEST_FQDN" envDefault:""`
	AllowedIPs      []string `env:"ALLOWED_IPS" envSeparator:"," envDefault:""`
	ListenIP        string   `env:"PROXY_LISTEN_IP" envDefault:"0.0.0.0"`
	RequireAuth     bool     `env:"REQUIRE_AUTH" envDefault:"true"`
	MaxConns        int      `env:"MAX_CONNECTIONS" envDefault:"100"`
	TimeoutSec      int      `env:"TIMEOUT" envDefault:"300"`
}

type timeoutConn struct {
	net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func (t timeoutConn) Read(b []byte) (int, error) {
	t.Conn.SetReadDeadline(time.Now().Add(t.readTimeout))

	return t.Conn.Read(b)
}

func (t timeoutConn) Write(b []byte) (int, error) {
	t.Conn.SetWriteDeadline(time.Now().Add(t.writeTimeout))

	return t.Conn.Write(b)
}

type limitListener struct {
	net.Listener
	sem         chan struct{}
	readTimeout time.Duration
	writeTimeout time.Duration
}

func (l *limitListener) Accept() (net.Conn, error) {
	l.sem <- struct{}{}
	c, err := l.Listener.Accept()
	if err != nil {
		<-l.sem
		return nil, err
	}

	return &limitConn{
		Conn:        timeoutConn{Conn: c, readTimeout: l.readTimeout, writeTimeout: l.writeTimeout},
		release:     func() { <-l.sem },
	}, nil
}

type limitConn struct {
	net.Conn
	release func()
}

func (l *limitConn) Close() error {
	l.release()
	return l.Conn.Close()
}

func main() {
	cfg := params{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse env: %+v", err)
	}

	socksConf := &socks5.Config{
		Logger: log.New(os.Stdout, "[SOCKS5] ", log.LstdFlags),
	}

	if cfg.RequireAuth {
		if cfg.User == "" || cfg.Password == "" {
			log.Fatalln("REQUIRE_AUTH=true, но PROXY_USER/PROXY_PASSWORD не заданы")
		}
		creds := socks5.StaticCredentials{cfg.User: cfg.Password}
		auth := socks5.UserPassAuthenticator{Credentials: creds}
		socksConf.AuthMethods = []socks5.Authenticator{auth}
	} else {
		log.Println("Warning: proxy без авторизации! Это небезопасно.")
	}

	if cfg.AllowedDestFqdn != "" {
		socksConf.Rules = PermitDestAddrPattern(cfg.AllowedDestFqdn)
	}

	server, err := socks5.New(socksConf)
	if err != nil {
		log.Fatalf("Ошибка создания socks5 сервера: %v", err)
	}

	if len(cfg.AllowedIPs) > 0 {
		ips := make([]net.IP, len(cfg.AllowedIPs))
		for i, ip := range cfg.AllowedIPs {
			ips[i] = net.ParseIP(ip)
		}

		server.SetIPWhitelist(ips)
	}

	listenAddr := ":" + cfg.Port
	if cfg.ListenIP != "" {
		listenAddr = cfg.ListenIP + ":" + cfg.Port
	}

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Не удалось слушать %s: %v", listenAddr, err)
	}

	limitedLn := &limitListener{
		Listener:    ln,
		sem:         make(chan struct{}, cfg.MaxConns),
		readTimeout:  time.Duration(cfg.TimeoutSec) * time.Second,
		writeTimeout: time.Duration(cfg.TimeoutSec) * time.Second,
	}

	log.Printf("SOCKS5 proxy слушает на %s", listenAddr)

	// Graceful shutdown обработка
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan
		log.Println("Получен сигнал завершения, закрываем слушатель...")
		limitedLn.Close()
	}()

	if err := server.Serve(limitedLn); err != nil {
		// Ошибка Serve может быть от Close() при graceful shutdown
		if err == net.ErrClosed {
			log.Println("Сервер корректно остановлен")
		} else {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}
}
