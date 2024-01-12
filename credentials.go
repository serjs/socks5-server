package main

import (
	"encoding/json"

	"github.com/armon/go-socks5"
)

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func parseCredentials(credsString string) (socks5.StaticCredentials, error) {
	var creds []credentials
	err := json.Unmarshal([]byte(credsString), &creds)
	if err != nil {
		return nil, err
	}

	var credsMap socks5.StaticCredentials
	for _, cred := range creds {
		credsMap[cred.Username] = cred.Password
	}

	return credsMap, nil
}
