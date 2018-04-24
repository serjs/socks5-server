package main

import (
	"fmt"
	"log"
	"os"

	"github.com/armon/go-socks5"
)

func main() {

	conf := &socks5.Config{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}

	if os.Getenv("PROXY_USER")+os.Getenv("PROXY_PASSWORD") != "" {
		creds := socks5.StaticCredentials{
			os.Getenv("PROXY_USER"): os.Getenv("PROXY_PASSWORD"),
		}
		cator := socks5.UserPassAuthenticator{Credentials: creds}
		conf.AuthMethods = []socks5.Authenticator{cator}
	}

	server, err := socks5.New(conf)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Start listening ...")
	if err := server.ListenAndServe("tcp", ":1080"); err != nil {
		log.Fatal(err)
	}
}
