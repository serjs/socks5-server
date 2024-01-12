package main

import (
	"log"
	"net"
	"os"

	"github.com/armon/go-socks5"
	"github.com/caarlos0/env/v6"
)

type params struct {
	Creds           string   `env:"PROXY_CREDENTIALS" envDefault:""`
	User            string   `env:"PROXY_USER" envDefault:""`
	Password        string   `env:"PROXY_PASSWORD" envDefault:""`
	Port            string   `env:"PROXY_PORT" envDefault:"1080"`
	AllowedDestFqdn string   `env:"ALLOWED_DEST_FQDN" envDefault:""`
	AllowedIPs      []string `env:"ALLOWED_IPS" envSeparator:"," envDefault:""`
}

func main() {
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

	var creds socks5.StaticCredentials
	if cfg.Creds != "" {
		creds, err = parseCredentials(cfg.Creds)
		if err != nil {
			log.Printf("%+v\n", err)
		}
	}
	if cfg.User+cfg.Password != "" {
		creds[cfg.User] = cfg.Password
	}
	if len(creds) > 0 {
		cator := socks5.UserPassAuthenticator{Credentials: creds}
		socks5conf.AuthMethods = []socks5.Authenticator{cator}
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

	log.Printf("Start listening proxy service on port %s\n", cfg.Port)
	if err := server.ListenAndServe("tcp", ":"+cfg.Port); err != nil {
		log.Fatal(err)
	}
}
