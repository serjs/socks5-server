# socks5-server
Simple socks5 server using go-socks5 with hardcoded auth

Run

```docker build -t <image_name> .```

Then
```docker run -d -p 1080:1080 -e PROXY_USER=<PROXY_USER -e PROXY_PASSWORD=<PROXY_PASSWORD <image_name>```