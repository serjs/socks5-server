package main

import (
	"errors"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	opt "github.com/akamensky/argparse"
	"github.com/armon/go-socks5"
	"github.com/sirupsen/logrus"
)

const USER_PASSWORD_SEPARATOR = "="

var (
	args = opt.NewParser("socks5",
		"socks5 is a simple socks5 server written in go")

	users = args.StringList("u", "user",
		&opt.Options{
			Help: "Add proxy user (format: username" + USER_PASSWORD_SEPARATOR + "password)",
			Validate: func(users []string) error {
				for _, user := range users {
					if len(strings.Split(user, USER_PASSWORD_SEPARATOR)) != 2 {
						return errors.New("wrong user format")
					}
				}
				return nil
			},
			Required: false,
		})

	port = args.Int("p", "port",
		&opt.Options{
			Help:    "Proxy port",
			Default: 1080,
		})

	network = "tcp"
)

func main() {
	// Load configuration
	err := args.Parse(os.Args)
	if err != nil {
		logrus.Fatal(args.Usage(err))
	}

	extractedUsers := make(map[string]string, 0)
	for _, u := range *users {
		values := strings.Split(u, USER_PASSWORD_SEPARATOR)
		extractedUsers[values[0]] = values[1]
	}

	// Socks5 configuration
	conf := &socks5.Config{
		Logger: log.New(logrus.New().Out, "[server]", log.Lmsgprefix),
	}

	credentials := socks5.StaticCredentials{}
	for user, password := range extractedUsers {
		logrus.WithField("user", user).Info("Adding user")
		credentials[user] = password
	}

	if len(credentials) > 0 {
		conf.Credentials = credentials
	}

	// Start server
	server, err := socks5.New(conf)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.WithField("port", *port).
		WithField("network", "tcp").
		Infof("Start listening proxy service")

	addr := net.JoinHostPort("", strconv.Itoa(*port))
	err = server.ListenAndServe(network, addr)
	if err != nil {
		logrus.Fatal(err)
	}
}
