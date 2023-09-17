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
		_, err := writer.Write([]byte(healthCheck(proxyPort)))
		if err != nil {
			log.Printf("Error while writing healthcheck response: %v", err)
		}
	}))
	if err != nil {
		log.Fatal(err)
	}
}

func healthCheck(port string) string {
	curl := exec.Command("curl", "--socks5", "localhost:"+port, "https://ifcfg.co", "-m", "2")
	err := curl.Run()
	if err != nil {
		log.Printf("Error while running curl: %v", err)
		return "FAIL"
	} else {
		return "OK"
	}
}
