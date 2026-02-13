package main

import (
	"fmt"
	"github.com/armon/go-socks5"
	"github.com/caarlos0/env/v6"
	"log"
	"net"
	"os"
	"time"
)

type params struct {
	User            string   `env:"PROXY_USER" envDefault:""`
	Password        string   `env:"PROXY_PASSWORD" envDefault:""`
	Port            string   `env:"PROXY_PORT" envDefault:"1080"`
	AllowedDestFqdn string   `env:"ALLOWED_DEST_FQDN" envDefault:""`
	AllowedIPs      []string `env:"ALLOWED_IPS" envSeparator:"," envDefault:""`
	ListenIP        string   `env:"PROXY_LISTEN_IP" envDefault:"0.0.0.0"`
	RequireAuth     bool     `env:"REQUIRE_AUTH" envDefault:"true"`
}

func main() {
	// Check if running in healthcheck mode
	if len(os.Args) > 1 && os.Args[1] == "--healthcheck" {
		runHealthCheck()
		return
	}
	// Working with app params
	cfg := params{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Printf("%+v\n", err)
	}

	//Initialize socks5 config
	socks5conf := &socks5.Config{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}

	if cfg.RequireAuth {
		if cfg.User == "" || cfg.Password == "" {
			log.Fatalln("Error: REQUIRE_AUTH is true, but PROXY_USER and PROXY_PASSWORD are not set.  The application will now exit.")
		}
		creds := socks5.StaticCredentials{
			cfg.User: cfg.Password,
		}
		cator := socks5.UserPassAuthenticator{Credentials: creds}
		socks5conf.AuthMethods = []socks5.Authenticator{cator}
	} else {
		log.Println("Warning: Running the proxy server without authentication. This is NOT recommended for public servers.")
	}

	if cfg.AllowedDestFqdn != "" {
		socks5conf.Rules = PermitDestAddrPattern(cfg.AllowedDestFqdn)
	}

	server, err := socks5.New(socks5conf)
	if err != nil {
		log.Fatal(err)
	}

	// Set IP whitelist
	if len(cfg.AllowedIPs) > 0 {
		whitelist := make([]net.IP, len(cfg.AllowedIPs))
		for i, ip := range cfg.AllowedIPs {
			whitelist[i] = net.ParseIP(ip)
		}
		server.SetIPWhitelist(whitelist)
	}

	listenAddr := ":" + cfg.Port
	if cfg.ListenIP != "" {
		listenAddr = cfg.ListenIP + ":" + cfg.Port
	}

	log.Printf("Start listening proxy service on %s\n", listenAddr)
	if err := server.ListenAndServe("tcp", listenAddr); err != nil {
		log.Fatal(err)
	}
}

func runHealthCheck() {
	port := os.Getenv("PROXY_PORT")
	if port == "" {
		port = "1080"
	}

	listenIP := os.Getenv("PROXY_LISTEN_IP")
	if listenIP == "" {
		listenIP = "127.0.0.1"
	}

	user := os.Getenv("PROXY_USER")
	password := os.Getenv("PROXY_PASSWORD")

	addr := listenIP + ":" + port

	// Try to connect to the SOCKS5 port
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Health check failed: cannot connect to %s: %v\n", addr, err)
		os.Exit(1)
	}
	defer conn.Close()

	// Send SOCKS5 greeting - always offer both no-auth and username/password
	// Format: [VERSION, NUM_METHODS, METHOD1, METHOD2]
	greeting := []byte{0x05, 0x02, 0x00, 0x02} // Version 5, 2 methods: No auth (0x00) and Username/Password (0x02)

	_, err = conn.Write(greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Health check failed: cannot write greeting: %v\n", err)
		os.Exit(1)
	}

	// Read server response
	// Format: [VERSION, CHOSEN_METHOD]
	response := make([]byte, 2)
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, err = conn.Read(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Health check failed: cannot read response: %v\n", err)
		os.Exit(1)
	}

	// Check if server accepted (version should be 5)
	if response[0] != 0x05 {
		fmt.Fprintf(os.Stderr, "Health check failed: invalid SOCKS version: %d\n", response[0])
		os.Exit(1)
	}

	// If server chose username/password authentication (0x02), perform auth
	if response[1] == 0x02 {
		if user == "" || password == "" {
			fmt.Fprintf(os.Stderr, "Health check failed: server requires auth but PROXY_USER/PROXY_PASSWORD not set\n")
			os.Exit(1)
		}

		// Send username/password
		// Format: [VERSION, USER_LEN, USERNAME, PASS_LEN, PASSWORD]
		authRequest := []byte{0x01} // Auth version
		authRequest = append(authRequest, byte(len(user)))
		authRequest = append(authRequest, []byte(user)...)
		authRequest = append(authRequest, byte(len(password)))
		authRequest = append(authRequest, []byte(password)...)

		_, err = conn.Write(authRequest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Health check failed: cannot write auth: %v\n", err)
			os.Exit(1)
		}

		// Read auth response
		// Format: [VERSION, STATUS]
		authResponse := make([]byte, 2)
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, err = conn.Read(authResponse)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Health check failed: cannot read auth response: %v\n", err)
			os.Exit(1)
		}

		// Check if authentication was successful (status should be 0)
		if authResponse[1] != 0x00 {
			fmt.Fprintf(os.Stderr, "Health check failed: authentication failed\n")
			os.Exit(1)
		}
	}

	fmt.Println("Health check passed")
	os.Exit(0)
}
