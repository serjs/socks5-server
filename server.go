package main

import (
	"github.com/armon/go-socks5"
	"os"
	"log"
)

func main() {
	creds := socks5.StaticCredentials{
		os.Getenv("PROXY_USER"):os.Getenv("PROXY_PASSWORD"),
	}
	cator := socks5.UserPassAuthenticator{Credentials: creds}
	// Create a SOCKS5 server
	conf := &socks5.Config{
		AuthMethods: []socks5.Authenticator{cator},
		Logger:      log.New(os.Stdout, "", log.LstdFlags),
	}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	// Create SOCKS5 proxy on localhost port 8000
	if err := server.ListenAndServe("tcp", "0.0.0.0:1080"); err != nil {
		panic(err)
	}
}