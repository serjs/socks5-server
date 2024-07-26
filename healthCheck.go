package main

import (
	"log"
	"net/http"
	"os/exec"
)

func startHealthCheck(healthCheckPort string, proxyPort string, user string, password string, allowedOrigins []string) {
	log.Printf("Start listening healthcheck on port %s\n", healthCheckPort)
	addr := ":" + healthCheckPort
	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
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
	})

	corsHandler := corsMiddleware(handler, allowedOrigins)

	err := http.ListenAndServe(addr, corsHandler)
	if err != nil {
		log.Fatal(err)
	}
}

func healthCheck(port string, user string, password string) bool {
	var curl *exec.Cmd
	if user == "" || password == "" {
		curl = exec.Command("curl", "--socks5", "localhost:"+port, "https://ifcfg.co", "-m", "2")
	} else {
		curl = exec.Command("curl", "--socks5", user+":"+password+"@localhost:"+port, "https://ifcfg.co")
	}
	err := curl.Run()
	if err != nil {
		log.Printf("Error while running curl: %v", err)
		return false
	} else {
		return true
	}
}

func corsMiddleware(next http.Handler, allowedOrigins []string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Allow origins from the allowedOrigins list
        origin := r.Header.Get("Origin")
        if contains(allowedOrigins, origin) {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
        }
        next.ServeHTTP(w, r)
    })
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
