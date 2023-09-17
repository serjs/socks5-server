package main

import (
	"log"
	"net/http"
	"os/exec"
)

func startHealthCheck(healthCheckPort string, proxyPort string, user string, password string) {
	log.Printf("Start listening healthcheck on port %s\n", healthCheckPort)
	addr := ":" + healthCheckPort
	err := http.ListenAndServe(addr, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var header int
		var textResponse string
		if healthCheck(proxyPort, user, password) {
			header = http.StatusOK
			textResponse = "OK"
		} else {
			header = http.StatusServiceUnavailable
			textResponse = "FAIL"
		}
		writer.WriteHeader(header)
		_, err := writer.Write([]byte(textResponse))
		if err != nil {
			log.Printf("Error while writing healthcheck response: %v", err)
		}
	}))
	if err != nil {
		log.Fatal(err)
	}
}

func healthCheck(port string, user string, password string) bool {
	var curl *exec.Cmd
	if user == "" || password == "" {
		curl = exec.Command("curl", "--socks5", "localhost:"+port, "https://ifcfg.co", "-m", "2")
	} else {
		curl = exec.Command("curl", "--socks5", "localhost:"+port, "https://ifcfg.co", "-m", "2", "-u", user+":"+password)
	}
	err := curl.Run()
	if err != nil {
		log.Printf("Error while running curl: %v", err)
		return false
	} else {
		return true
	}
}
