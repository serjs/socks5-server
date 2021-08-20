# go-socks5-proxy

![Latest tag from master branch](https://github.com/serjs/socks5-server/workflows/Latest%20tag%20from%20master%20branch/badge.svg)
![Release tag](https://github.com/serjs/socks5-server/workflows/Release%20tag/badge.svg)

Simple socks5 server using go-socks5 with authentication options

## How to use

### Docker (recommended)

```bash
docker run -d \
    -p '1080:1080' \
    serjs/go-socks5-proxy \
    --user 'someuser=somepass' \
    --user 'someuser2=somepass2'
```

Remove `--user` parameters to disable authentication on your socks5 server.

or

```yaml
version: '3'
services:
  socks5:
    image: serjs/go-socks5-proxy
    ports:
      - '1080:1080'
    command: |
      --user 'someuser=somepass'
      --user 'someuser2=somepass2'
```

Remove `command` field to disable authentication on your socks5 server.

## Usage

```bash
usage: socks5 [-h|--help] [-u|--user "<value>" [-u|--user "<value>" ...]]
              [-p|--port <integer>]

              socks5 is a simple socks5 server written in go

Arguments:

  -h  --help  Print help information
  -u  --user  Add proxy user (format: username=password)
  -p  --port  Proxy port. Default: 1080
```

## Test running service

### Without authentication

Result must show docker host ip (for bridged network)

```bash
curl --socks5 <docker host ip>:1080 http://ifcfg.co
```

or

```bash
docker run --rm curlimages/curl:7.65.3 -s --socks5 <docker host ip>:1080 http://ifcfg.co
```

### With authentication

Result must show docker host ip (for bridged network)

```bash
curl --socks5 <docker host ip>:1080 -U <PROXY_USER>:<PROXY_PASSWORD> http://ifcfg.co
```

or

```bash
docker run --rm curlimages/curl:7.65.3 -s --socks5 <PROXY_USER>:<PROXY_PASSWORD>@<docker host ip>:1080 http://ifcfg.co
```

## Authors

- **Sergey Bogayrets**

See also the list of [contributors](https://github.com/serjs/socks5-server/graphs/contributors) who participated in this project.
