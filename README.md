# go-socks5-proxy

Simple socks5 server using go-socks5 with or without auth.

# Start container with proxy
```docker run -d --name socks5 -p 1080:1080 -e PROXY_USER=<PROXY_USER> -e PROXY_PASSWORD=<PROXY_PASSWORD>  olebedev/socks5```

For auth-less mode just do not pass `PROXY_USER` and `PROXY_PASSWORD`.

# Test running service
```curl --socks5 <docker machine ip>:1080  https://ya.ru``` - result must show docker host ip (for bridged network)
