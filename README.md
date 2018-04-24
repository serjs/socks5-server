# go-socks5-proxy

Simple socks5 server using go-socks5 with or without auth.

## Start container with proxy

```docker run -d --name socks5 -p 1080:1080 -e PROXY_USER=<PROXY_USER> -e PROXY_PASSWORD=<PROXY_PASSWORD>  serjs/go-socks5-proxy```

For auth-less mode just do not pass envrionment variables `PROXY_USER` and `PROXY_PASSWORD`.

## Test running service

Without auth

```curl --socks5 <docker host ip>:1080  http://ifcfg.co``` - result must show docker host ip (for bridged network)

With auth

```curl --socks5 --user <PROXY_USER:<PROXY_PASSWORD> <docker host ip>:1080  http://ifcfg.co``` - result must show docker host ip (for bridged network)

## Authors

* **Sergey Bogayrets**

See also the list of [contributors](https://github.com/serjs/socks5-server/graphs/contributors) who participated in this project.
