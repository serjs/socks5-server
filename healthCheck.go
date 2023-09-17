package main

import (
	"log"
	"net/http"
	"os/exec"
)

func startHealthCheck(healthCheckPort string, proxyPort string) {
	log.Printf("Start listening healthcheck on port %s\n", healthCheckPort)
	addr := ":" + healthCheckPort
	err := http.ListenAndServe(addr, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var header int
		var textResponse string
		if healthCheck(proxyPort) {
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

func healthCheck(port string) bool {
	curl := exec.Command("curl", "--socks5", "localhost:"+port, "https://ifcfg.co", "-m", "2")
	err := curl.Run()
	if err != nil {
		log.Printf("Error while running curl: %v", err)
		return false
	} else {
		return true
	}
}
