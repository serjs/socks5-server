package main

import (
	"log"
	"net/netip"
	"os"

	"github.com/armon/go-socks5"
	env "github.com/caarlos0/env/v6"
)

type params struct {
	User            string   `env:"PROXY_USER" envDefault:""`
	Password        string   `env:"PROXY_PASSWORD" envDefault:""`
	Port            string   `env:"PROXY_PORT" envDefault:"1080"`
	AllowedDestFqdn string   `env:"ALLOWED_DEST_FQDN" envDefault:""`
	AllowedIPs      []string `env:"ALLOWED_IPS" envSeparator:"," envDefault:""`
	AllowedNets     []string `env:"ALLOWED_NETS" envSeparator:"," envDefault:""`
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

	if cfg.User+cfg.Password != "" {
		creds := socks5.StaticCredentials{
			os.Getenv("PROXY_USER"): os.Getenv("PROXY_PASSWORD"),
		}
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
	if len(cfg.AllowedIPs) > 0 || len(cfg.AllowedNets) > 0 {
		whitelist := make([]netip.Addr, len(cfg.AllowedIPs))
		whitelistnet := make([]netip.Prefix, len(cfg.AllowedNets))

		if len(cfg.AllowedIPs) > 0 {
			for i, ip := range cfg.AllowedIPs {
				whitelist[i], _ = netip.ParseAddr(ip)
			}
		}
		if len(cfg.AllowedNets) > 0 {
			for i, inet := range cfg.AllowedNets {
				whitelistnet[i], _ = netip.ParsePrefix(inet)
			}
		}

		server.SetIPWhitelist(whitelist, whitelistnet)
	}

	log.Printf("Started proxy service listening on port %s\n", cfg.Port)
	if err := server.ListenAndServe("tcp", ":"+cfg.Port); err != nil {
		log.Fatal(err)
	}
}
