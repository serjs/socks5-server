# go-socks5-proxy

![Latest tag from master branch](https://github.com/serjs/socks5-server/workflows/Latest%20tag%20from%20master%20branch/badge.svg)

Simple socks5 server using go-socks5 with authentication, allowed ips list and destination FQDNs filtering

## ℹ️ About this fork

This is a **maintained fork** of the original [serjs/socks5-server](https://github.com/serjs/socks5-server) project with focus on **reliability and production readiness**. 

### 🔧 Key improvements in this fork

- ✅ **Graceful shutdown** — Properly handles `SIGTERM`/`SIGINT` signals for safe container termination
- ✅ **Regex validation** — FQDN patterns are validated at startup (fail-fast on config errors)
- ✅ **Enhanced logging** — Rejected connections are logged with details for easier troubleshooting
- ✅ **Unit tests** — Comprehensive test suite covering:
  - FQDN filtering logic with regex patterns
  - IP address validation and parsing
  - Connection limiting and semaphore behavior
  - Configuration parsing from environment variables
- ✅ **Bug fixes** — Semaphore leak prevention in error scenarios
- ✅ **Performance** — Compiled regex patterns are cached for optimal performance

### 📈 Test coverage

- 17+ unit tests, 30+ subtests
- All tests pass in < 1 second
- Ready for CI/CD pipelines

# Examples

- Run docker container using default container port 1080 and expose it to world using host port 1080, with auth creds

    ```docker run -d --name socks5 -p 1080:1080 -e PROXY_USER=<PROXY_USER> -e PROXY_PASSWORD=<PROXY_PASSWORD>  serjs/go-socks5-proxy```

- Run docker container using specific container port and expose it to host port 1090

    ```docker run -d --name socks5 -p 1090:9090 -e PROXY_USER=<PROXY_USER> -e PROXY_PASSWORD=<PROXY_PASSWORD> -e PROXY_PORT=9090 serjs/go-socks5-proxy```

# List of supported config parameters

|ENV variable|Type|Default|Description|
|------------|----|-------|-----------|
|REQUIRE_AUTH|String|true|Allow accepting socks5 connections without auth creds. Not recommended untill you use other protections mechanisms like Whitelists Subnets using Firewall or Proxy itself|
|PROXY_USER|String|EMPTY|Set proxy user (also required existed PROXY_PASS)|
|PROXY_PASSWORD|String|EMPTY|Set proxy password for auth, used with PROXY_USER|
|PROXY_PORT|String|1080|Set listen port for application inside docker container|
|PROXY_LISTEN_IP|String|0.0.0.0|Listening address. Change to 127.0.0.1 if you want to restrict proxy access to localhost only.|
|ALLOWED_DEST_FQDN|String|EMPTY|Allowed destination address regular expression pattern. Default allows all.|
|ALLOWED_IPS|String|Empty|Set allowed IP's that can connect to proxy, separator `,`|
|MAX_CONNECTIONS|Int|100|Maximum number of concurrent client connections. New connections are queued until slots free up.|
|TIMEOUT|Int|300|Connection read/write timeout in seconds. Prevents stuck or idle connections from hanging forever.|
|**TLS_ENABLED**|Boolean|false|Enable SOCKS5 over TLS. Requires TLS_CERT_FILE and TLS_KEY_FILE.|
|**TLS_PORT**|String|10443|Listen port for TLS connections.|
|**TLS_CERT_FILE**|String|EMPTY|Path to TLS certificate file (PEM format). Required if TLS_ENABLED=true.|
|**TLS_KEY_FILE**|String|EMPTY|Path to TLS private key file (PEM format). Required if TLS_ENABLED=true.|
|**TLS_MIN_VERSION**|String|1.2|Minimum TLS version (1.0, 1.1, 1.2, 1.3).|

### Configuration Examples

**Restrict to specific domains:**
```bash
ALLOWED_DEST_FQDN='^.*\.(example\.com|internal\.local)$'
```

**Allow only specific IPs:**
```bash
ALLOWED_IPS=192.168.1.10,10.0.0.5
```

**Disable authentication (not recommended):**
```bash
REQUIRE_AUTH=false
```

**Enable TLS (SOCKS5 over TLS):**
```bash
TLS_ENABLED=true
TLS_PORT=10443
TLS_CERT_FILE=/etc/ssl/certs/proxy.crt
TLS_KEY_FILE=/etc/ssl/private/proxy.key
TLS_MIN_VERSION=1.2
```

### Generate Self-Signed Certificate

For testing purposes, generate a self-signed certificate:

```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
```

For production, use a certificate from Let's Encrypt or another CA.

### Running with TLS

```bash
docker run -d --name socks5-tls \
  -p 1080:1080 -p 10443:10443 \
  -e PROXY_USER=<user> \
  -e PROXY_PASSWORD=<pass> \
  -e TLS_ENABLED=true \
  -e TLS_CERT_FILE=/certs/cert.pem \
  -e TLS_KEY_FILE=/certs/key.pem \
  -v /path/to/certs:/certs:ro \
  serjs/go-socks5-proxy
```

# Build your own image:
`docker-compose -f docker-compose.build.yml up -d`\
Just don't forget to set parameters in the `.env` file (`cp .env.example .env)` and edit it with your config parameters

# Test running service

Assuming that you are using container on 1080 host docker port

## Without authentication

```curl --socks5 <docker host ip>:1080  https://ipinfo.io``` - result must show docker host ip (for bridged network)

or

```docker run --rm curlimages/curl:7.65.3 -s --socks5 <docker host ip>:1080 https://ipinfo.io```

## With authentication

```curl --socks5 <docker host ip>:1080 -U <PROXY_USER>:<PROXY_PASSWORD> https://ipinfo.io```

or

```docker run --rm curlimages/curl:7.65.3 -s --socks5 <PROXY_USER>:<PROXY_PASSWORD>@<docker host ip>:1080 https://ipinfo.io```

## With TLS (SOCKS5 over TLS)

To connect via TLS, use `socat` or a SOCKS5 client that supports TLS:

**Using socat as a TLS wrapper:**
```bash
# Create a local tunnel that wraps SOCKS5 in TLS
socat TCP-LISTEN:1080,reuseaddr,fork OPENSSL:<server-ip>:10443,verify=0

# Then use curl through the tunnel
curl --socks5 localhost:1080 -U <user>:<pass> https://ipinfo.io
```

**Using proxychains with TLS:**
```bash
# First create SSH tunnel or use stunnel
stunnel -c -s -d 1080 -r <server-ip>:10443

# Then configure proxychains to use localhost:1080
```

**Note:** The client must establish a TLS connection first, then speak SOCKS5 protocol over the encrypted channel.

# Running tests

```bash
# Run all unit tests
go test -v ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestPermitDestAddrPattern ./...
```

# Authors

* **Sergey Bogayrets** (original project)
* Maintained fork with improvements

See also the list of [contributors](https://github.com/serjs/socks5-server/graphs/contributors) who participated in this project.
