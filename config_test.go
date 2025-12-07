package main

import (
	"net"
	"os"
	"testing"

	"github.com/caarlos0/env/v6"
)

// TestParseConfig_Defaults проверяет значения по умолчанию
func TestParseConfig_Defaults(t *testing.T) {
	// Очищаем переменные окружения
	os.Clearenv()

	cfg := params{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %+v", err)
	}

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Port", cfg.Port, "1080"},
		{"ListenIP", cfg.ListenIP, "0.0.0.0"},
		{"RequireAuth", cfg.RequireAuth, true},
		{"MaxConns", cfg.MaxConns, 100},
		{"TimeoutSec", cfg.TimeoutSec, 300},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s: got %v, expected %v", tt.name, tt.got, tt.expected)
		}
	}
}

// TestParseConfig_CustomValues проверяет парсинг кастомных значений
func TestParseConfig_CustomValues(t *testing.T) {
	// Устанавливаем переменные окружения
	os.Setenv("PROXY_USER", "testuser")
	os.Setenv("PROXY_PASSWORD", "testpass")
	os.Setenv("PROXY_PORT", "9090")
	os.Setenv("PROXY_LISTEN_IP", "127.0.0.1")
	os.Setenv("MAX_CONNECTIONS", "50")
	os.Setenv("TIMEOUT", "120")
	os.Setenv("REQUIRE_AUTH", "false")
	defer os.Clearenv()

	cfg := params{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %+v", err)
	}

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"User", cfg.User, "testuser"},
		{"Password", cfg.Password, "testpass"},
		{"Port", cfg.Port, "9090"},
		{"ListenIP", cfg.ListenIP, "127.0.0.1"},
		{"MaxConns", cfg.MaxConns, 50},
		{"TimeoutSec", cfg.TimeoutSec, 120},
		{"RequireAuth", cfg.RequireAuth, false},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s: got %v, expected %v", tt.name, tt.got, tt.expected)
		}
	}
}

// TestParseConfig_AllowedIPs проверяет парсинг списка IP адресов
func TestParseConfig_AllowedIPs(t *testing.T) {
	os.Setenv("ALLOWED_IPS", "192.168.1.1,10.0.0.1,172.16.0.1")
	defer os.Clearenv()

	cfg := params{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %+v", err)
	}

	expectedIPs := []string{"192.168.1.1", "10.0.0.1", "172.16.0.1"}
	if len(cfg.AllowedIPs) != len(expectedIPs) {
		t.Errorf("Expected %d IPs, got %d", len(expectedIPs), len(cfg.AllowedIPs))
	}

	for i, expectedIP := range expectedIPs {
		if cfg.AllowedIPs[i] != expectedIP {
			t.Errorf("IP[%d]: got %s, expected %s", i, cfg.AllowedIPs[i], expectedIP)
		}
	}
}

// TestParseIP_Valid проверяет парсинг валидных IP адресов
func TestParseIP_Valid(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{"IPv4 standard", "192.168.1.1"},
		{"IPv4 localhost", "127.0.0.1"},
		{"IPv4 zero", "0.0.0.0"},
		{"IPv4 broadcast", "255.255.255.255"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := net.ParseIP(tt.ip)
			if parsed == nil {
				t.Errorf("Failed to parse valid IP: %s", tt.ip)
			}
		})
	}
}

// TestParseIP_Invalid проверяет обработку невалидных IP адресов
func TestParseIP_Invalid(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{"Invalid octet", "256.1.1.1"},
		{"Incomplete", "192.168.1"},
		{"Too many octets", "192.168.1.1.1"},
		{"Non-numeric", "192.168.a.1"},
		{"Empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := net.ParseIP(tt.ip)
			if parsed != nil {
				t.Errorf("Expected nil for invalid IP: %s", tt.ip)
			}
		})
	}
}

// TestParseConfig_AllowedDestFqdn проверяет парсинг regex для FQDN
func TestParseConfig_AllowedDestFqdn(t *testing.T) {
	os.Setenv("ALLOWED_DEST_FQDN", "^.*\\.example\\.com$")
	defer os.Clearenv()

	cfg := params{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %+v", err)
	}

	if cfg.AllowedDestFqdn != "^.*\\.example\\.com$" {
		t.Errorf("AllowedDestFqdn: got %s, expected %s",
			cfg.AllowedDestFqdn, "^.*\\.example\\.com$")
	}
}

// TestParseConfig_EmptyAllowedLists проверяет пустые списки разрешений
func TestParseConfig_EmptyAllowedLists(t *testing.T) {
	os.Clearenv()
	os.Setenv("ALLOWED_IPS", "")
	os.Setenv("ALLOWED_DEST_FQDN", "")

	cfg := params{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %+v", err)
	}

	if len(cfg.AllowedIPs) != 0 {
		t.Errorf("Expected empty AllowedIPs, got %v", cfg.AllowedIPs)
	}

	if cfg.AllowedDestFqdn != "" {
		t.Errorf("Expected empty AllowedDestFqdn, got %s", cfg.AllowedDestFqdn)
	}
}
