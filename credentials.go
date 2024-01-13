package main

import (
	"encoding/json"

	"github.com/armon/go-socks5"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func getCredentials(params params) (socks5.StaticCredentials, error) {
	var creds socks5.StaticCredentials = make(socks5.StaticCredentials)

	if params.Creds != "" {
		var parsed_env_creds []Credentials
		err := json.Unmarshal([]byte(params.Creds), &parsed_env_creds)
		if err != nil {
			return nil, err
		}

		for _, kv := range parsed_env_creds {
			creds[kv.Username] = kv.Password
		}
	}

	if params.User+params.Password != "" {
		creds[params.User] = params.Password
	}

	return creds, nil
}
