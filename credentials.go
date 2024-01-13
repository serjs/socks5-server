package main

import (
	"encoding/json"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func parseCredentials(credsString string) ([]Credentials, error) {
	var creds []Credentials
	err := json.Unmarshal([]byte(credsString), &creds)
	if err != nil {
		return nil, err
	}

	return creds, nil
}
