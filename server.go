package main

import (
	"log"
	"net"
	"os"
	"github.com/armon/go-socks5"
	"github.com/caarlos0/env/v6"
)

type params struct {
	User            string    `env:"PROXY_USER" envDefault:""`
	Password        string    `env:"PROXY_PASSWORD" envDefault:""`
	Port            string    `env:"PROXY_PORT" envDefault:"1080"`
	AllowedDestFqdn string    `env:"ALLOWED_DEST_FQDN" envDefault:""`
	AllowedDestIP   []string  `env:"ALLOWED_DEST_IP" envSeparator:"," envDefault:""`
	AllowedIPs      []string  `env:"ALLOWED_IPS" envSeparator:"," envDefault:""`
	ListenIP 		string 	  `env:"PROXY_LISTEN_IP" envDefault:"0.0.0.0"`
	RequireAuth		bool      `env:"REQUIRE_AUTH" envDefault:"true"`
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

	var rule socks5.RuleSet
	if cfg.AllowedDestFqdn != "" {
		fqdnRule, err := PermitDestAddrPatternCompiled(cfg.AllowedDestFqdn)
		if err != nil {
			log.Fatalf("Invalid ALLOWED_DEST_FQDN regex: %v\n", err)
		}
		rule = fqdnRule
	}

	if len(cfg.AllowedDestIP) > 0 {
		ipRule, err := PermitDestAddrIP(cfg.AllowedDestIP)
		if err != nil {
			log.Fatalf("Invalid ALLOWED_DEST_IP list: %v\n", err)
		}
		if rule == nil {
			rule = ipRule
		} else {
			rule = CombineRuleSets(rule, ipRule)
		}
	}

	if rule != nil {
		socks5conf.Rules = rule
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
