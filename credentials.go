package main

import (
	"encoding/json"

	"github.com/armon/go-socks5"
)

type credentials struct {
	username string `json:"username"`
	password string `json:"password"`
}

func parseCredentials(credsString string) (socks5.StaticCredentials, error) {
	var creds []credentials
	err := json.Unmarshal([]byte(credsString), &creds)
	if err != nil {
		return nil, err
	}

	var credsMap socks5.StaticCredentials
	for _, cred := range creds {
		credsMap[cred.username] = cred.password
	}

	return credsMap, nil
}
