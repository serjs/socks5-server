# go-socks5-proxy

![Latest tag from master branch](https://github.com/serjs/socks5-server/workflows/Latest%20tag%20from%20master%20branch/badge.svg)

Simple socks5 server using go-socks5 with authentication, allowed ips list and destination FQDNs filtering

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
|PROXY_LISTEN_IP|String|0.0.0.0|Set listen IP for application inside docker container|
|ALLOWED_DEST_FQDN|String|EMPTY|Allowed destination address regular expression pattern. Default allows all.|
|ALLOWED_IPS|String|Empty|Set allowed IP's that can connect to proxy, separator `,`|

# Health Check

The application includes built-in health check functionality via the `--healthcheck` flag. This performs a full SOCKS5 protocol handshake to verify the server is running and accepting connections.

## Usage

**Docker Compose:**
```yaml
services:
  socks5-proxy:
    image: serjs/go-socks5-proxy
    healthcheck:
      test: ["/app/socks5", "--healthcheck"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 10s
```

## How it works

- Connects to the SOCKS5 port and performs protocol handshake
- Automatically detects if authentication is required
- Uses `PROXY_USER` and `PROXY_PASSWORD` environment variables for authentication
- Returns exit code 0 (success) or 1 (failure)
- Works with both authenticated (`REQUIRE_AUTH=true`) and non-authenticated modes
- No additional ports or HTTP endpoints required

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

# Authors

* **Sergey Bogayrets**

See also the list of [contributors](https://github.com/serjs/socks5-server/graphs/contributors) who participated in this project.