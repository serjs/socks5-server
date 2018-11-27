# go-socks5-proxy

Simple socks5 server using go-socks5 with or without auth

## Start container with proxy

```docker run -d --name socks5 -p 1080:1080 -e PROXY_USER=<PROXY_USER> -e PROXY_PASSWORD=<PROXY_PASSWORD>  serjs/go-socks5-proxy```

For auth-less mode just do not pass envrionment variables `PROXY_USER` and `PROXY_PASSWORD`.

## List of all supported config parameters

|ENV variable|Type|Default|Description|
|------------|----|-------|-----------|
|PROXY_USER|String|EMPTY|Set proxy user (also required existed PROXY_PASS)|
|PROXY_PASSWORD|String|EMPTY|Set proxy password for auth, used with PROXY_USER|
|PROXY_PORT|String|1080|Set listen port for application|

## Test running service

Without auth

```curl --socks5 <docker host ip>:1080  http://ifcfg.co``` - result must show docker host ip (for bridged network)

With auth

```curl --proxy-user <PROXY_USER>:<PROXY_PASSWORD> --socks5 <docker host ip>:1080  http://ifcfg.co``` - result must show docker host ip (for bridged network)

## Authors

* **Sergey Bogayrets**

See also the list of [contributors](https://github.com/serjs/socks5-server/graphs/contributors) who participated in this project.
