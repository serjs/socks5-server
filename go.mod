module github.com/serjs/socks5-server

go 1.24.0

require (
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/caarlos0/env/v6 v6.10.1
)

require golang.org/x/net v0.45.0

replace github.com/armon/go-socks5 => github.com/serjs/go-socks5 v0.0.0-20250923183437-3920b97ee0d2
