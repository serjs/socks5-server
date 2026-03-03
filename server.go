package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
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
	// TLS settings
	TLSEnabled   bool   `env:"TLS_ENABLED" envDefault:"false"`
	TLSPort      string `env:"TLS_PORT" envDefault:"10443"`
	TLSCertFile  string `env:"TLS_CERT_FILE" envDefault:""`
	TLSKeyFile   string `env:"TLS_KEY_FILE" envDefault:""`
	TLSMinVersion string `env:"TLS_MIN_VERSION" envDefault:"1.2"`
}

type timeoutConn struct {
	net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func (t timeoutConn) Read(b []byte) (int, error) {
	if t.readTimeout > 0 {
		t.Conn.SetReadDeadline(time.Now().Add(t.readTimeout))
	}
	return t.Conn.Read(b)
}

func (t timeoutConn) Write(b []byte) (int, error) {
	if t.writeTimeout > 0 {
		t.Conn.SetWriteDeadline(time.Now().Add(t.writeTimeout))
	}
	return t.Conn.Write(b)
}

type limitListener struct {
	net.Listener
	sem          chan struct{}
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func (l *limitListener) Accept() (net.Conn, error) {
	l.sem <- struct{}{}
	defer func() {
		// Release semaphore slot if Accept returns an error
		// (connection won't be created, so release() won't be called)
		if err := recover(); err != nil {
			<-l.sem
			panic(err)
		}
	}()

	c, err := l.Listener.Accept()
	if err != nil {
		<-l.sem
		return nil, err
	}

	return &limitConn{
		Conn:    timeoutConn{Conn: c, readTimeout: l.readTimeout, writeTimeout: l.writeTimeout},
		release: func() { <-l.sem },
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

	// Validate TLS configuration
	if cfg.TLSEnabled {
		if cfg.TLSCertFile == "" || cfg.TLSKeyFile == "" {
			log.Fatalln("TLS_ENABLED=true but TLS_CERT_FILE/TLS_KEY_FILE not set")
		}
		// Check if cert files exist
		if _, err := os.Stat(cfg.TLSCertFile); os.IsNotExist(err) {
			log.Fatalf("TLS certificate file not found: %s", cfg.TLSCertFile)
		}
		if _, err := os.Stat(cfg.TLSKeyFile); os.IsNotExist(err) {
			log.Fatalf("TLS key file not found: %s", cfg.TLSKeyFile)
		}
	}

	socksConf := &socks5.Config{
		Logger: log.New(os.Stdout, "[SOCKS5] ", log.LstdFlags),
	}

	if cfg.RequireAuth {
		if cfg.User == "" || cfg.Password == "" {
			log.Fatalln("REQUIRE_AUTH=true but PROXY_USER/PROXY_PASSWORD not set")
		}
		creds := socks5.StaticCredentials{cfg.User: cfg.Password}
		auth := socks5.UserPassAuthenticator{Credentials: creds}
		socksConf.AuthMethods = []socks5.Authenticator{auth}
	} else {
		log.Println("Warning: proxy running without authentication! This is insecure.")
	}

	if cfg.AllowedDestFqdn != "" {
		socksConf.Rules = PermitDestAddrPattern(cfg.AllowedDestFqdn)
	}

	server, err := socks5.New(socksConf)
	if err != nil {
		log.Fatalf("Failed to create socks5 server: %v", err)
	}

	if len(cfg.AllowedIPs) > 0 {
		ips := make([]net.IP, len(cfg.AllowedIPs))
		for i, ip := range cfg.AllowedIPs {
			parsed := net.ParseIP(ip)
			if parsed == nil {
				log.Fatalf("Invalid IP address in ALLOWED_IPS: %s", ip)
			}
			ips[i] = parsed
		}

		server.SetIPWhitelist(ips)
	}

	// WaitGroup for graceful shutdown
	var wg sync.WaitGroup

	// Start plain TCP listener
	listenAddr := ":" + cfg.Port
	if cfg.ListenIP != "" {
		listenAddr = cfg.ListenIP + ":" + cfg.Port
	}

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", listenAddr, err)
	}

	plainLimitedLn := &limitListener{
		Listener:    ln,
		sem:         make(chan struct{}, cfg.MaxConns),
		readTimeout:  time.Duration(cfg.TimeoutSec) * time.Second,
		writeTimeout: time.Duration(cfg.TimeoutSec) * time.Second,
	}

	log.Printf("SOCKS5 proxy listening on %s (plain)", listenAddr)

	// Start TLS listener if enabled
	var tlsLimitedLn *limitListener
	if cfg.TLSEnabled {
		tlsListenAddr := ":" + cfg.TLSPort
		if cfg.ListenIP != "" {
			tlsListenAddr = cfg.ListenIP + ":" + cfg.TLSPort
		}

		// Parse TLS version
		tlsMinVersion, err := parseTLSVersion(cfg.TLSMinVersion)
		if err != nil {
			log.Fatalf("Invalid TLS_MIN_VERSION: %v", err)
		}

		tlsConfig := &tls.Config{
			MinVersion: tlsMinVersion,
			ClientAuth: tls.NoClientCert,
		}

		tlsLn, err := tls.Listen("tcp", tlsListenAddr, tlsConfig)
		if err != nil {
			log.Fatalf("Failed to listen on TLS %s: %v", tlsListenAddr, err)
		}

		tlsLimitedLn = &limitListener{
			Listener:    tlsLn,
			sem:         make(chan struct{}, cfg.MaxConns),
			readTimeout:  time.Duration(cfg.TimeoutSec) * time.Second,
			writeTimeout: time.Duration(cfg.TimeoutSec) * time.Second,
		}

		log.Printf("SOCKS5 proxy listening on %s (TLS %s)", tlsListenAddr, cfg.TLSMinVersion)
	}

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, closing listeners...")
		plainLimitedLn.Close()
		if tlsLimitedLn != nil {
			tlsLimitedLn.Close()
		}
	}()

	// Serve plain connections
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Serve(plainLimitedLn); err != nil {
			if err == net.ErrClosed {
				log.Println("Plain server stopped gracefully")
			} else {
				log.Fatalf("Plain server error: %v", err)
			}
		}
	}()

	// Serve TLS connections if enabled
	if tlsLimitedLn != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := server.Serve(tlsLimitedLn); err != nil {
				if err == net.ErrClosed {
					log.Println("TLS server stopped gracefully")
				} else {
					log.Fatalf("TLS server error: %v", err)
				}
			}
		}()
	}

	wg.Wait()
	log.Println("All servers stopped")
}

// parseTLSVersion converts TLS version string to tls.Version constant
func parseTLSVersion(version string) (uint16, error) {
	versions := map[string]uint16{
		"1.0": tls.VersionTLS10,
		"1.1": tls.VersionTLS11,
		"1.2": tls.VersionTLS12,
		"1.3": tls.VersionTLS13,
	}
	v, ok := versions[version]
	if !ok {
		return 0, fmt.Errorf("unsupported TLS version: %s (supported: 1.0, 1.1, 1.2, 1.3)", version)
	}
	return v, nil
}
