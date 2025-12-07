package main

import (
	"context"
	"testing"

	"github.com/armon/go-socks5"
)

// TestPermitDestAddrPattern_ValidPatterns тестирует валидные паттерны
func TestPermitDestAddrPattern_ValidPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		fqdn    string
		want    bool
	}{
		{
			name:    "Exact domain match",
			pattern: "^example\\.com$",
			fqdn:    "example.com",
			want:    true,
		},
		{
			name:    "Subdomain match",
			pattern: "^example\\.com$",
			fqdn:    "sub.example.com",
			want:    false,
		},
		{
			name:    "Wildcard subdomain",
			pattern: ".*\\.example\\.com$",
			fqdn:    "api.example.com",
			want:    true,
		},
		{
			name:    "Multiple domain alternatives",
			pattern: "^(google\\.com|facebook\\.com)$",
			fqdn:    "google.com",
			want:    true,
		},
		{
			name:    "Multiple domain alternatives - no match",
			pattern: "^(google\\.com|facebook\\.com)$",
			fqdn:    "twitter.com",
			want:    false,
		},
		{
			name:    "Allow all pattern",
			pattern: ".*",
			fqdn:    "anything.example.com",
			want:    true,
		},
		{
			name:    "Empty FQDN",
			pattern: "^example\\.com$",
			fqdn:    "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := PermitDestAddrPattern(tt.pattern)
			ctx := context.Background()
			req := &socks5.Request{
				DestAddr: &socks5.AddrSpec{
					FQDN: tt.fqdn,
				},
			}

			_, got := rs.Allow(ctx, req)
			if got != tt.want {
				t.Errorf("Allow() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPermitDestAddrPattern_ComplexPatterns тестирует сложные паттерны
func TestPermitDestAddrPattern_ComplexPatterns(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		fqdn    string
		want    bool
	}{
		{
			name:    "Nested domain",
			pattern: "^api\\..*\\.example\\.com$",
			fqdn:    "api.v2.example.com",
			want:    true,
		},
		{
			name:    "Character class - letters only",
			pattern: "[a-z]+\\.example\\.com",
			fqdn:    "api.example.com",
			want:    true,
		},
		{
			name:    "Character class - with digits (no match)",
			pattern: "[a-z]+\\.example\\.com",
			fqdn:    "api123.example.com",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := PermitDestAddrPattern(tt.pattern)
			ctx := context.Background()
			req := &socks5.Request{
				DestAddr: &socks5.AddrSpec{
					FQDN: tt.fqdn,
				},
			}

			_, got := rs.Allow(ctx, req)
			if got != tt.want {
				t.Errorf("Allow() = %v, want %v", got, tt.want)
			}
		})
	}
}

// BenchmarkPermitDestAddrPattern_Allow проверяет производительность фильтрации
func BenchmarkPermitDestAddrPattern_Allow(b *testing.B) {
	rs := PermitDestAddrPattern(".*\\.example\\.com$")
	ctx := context.Background()
	req := &socks5.Request{
		DestAddr: &socks5.AddrSpec{
			FQDN: "api.v2.example.com",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rs.Allow(ctx, req)
	}
}
