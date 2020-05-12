package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/armon/go-socks5"
	"github.com/caarlos0/env"
)

type params struct {
	User     string        `env:"PROXY_USER" envDefault:""`
	Password string        `env:"PROXY_PASSWORD" envDefault:""`
	Port     string        `env:"PROXY_PORT" envDefault:"1080"`
	Timeout  time.Duration `env:"DIAL_TIMEOUT" envDefault:"3s"`
}

func main() {
	// Working with app params
	cfg := params{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Printf("%+v\n", err)
	}

	//Initialize socks5 config
	socsk5conf := &socks5.Config{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: cfg.Timeout,
			}
			return d.DialContext(ctx, network, addr)
		},
	}

	if cfg.User+cfg.Password != "" {
		creds := socks5.StaticCredentials{
			os.Getenv("PROXY_USER"): os.Getenv("PROXY_PASSWORD"),
		}
		cator := socks5.UserPassAuthenticator{Credentials: creds}
		socsk5conf.AuthMethods = []socks5.Authenticator{cator}
	}

	server, err := socks5.New(socsk5conf)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Start listening proxy service on port %s\n", cfg.Port)
	if err := server.ListenAndServe("tcp", ":"+cfg.Port); err != nil {
		log.Fatal(err)
	}
}
