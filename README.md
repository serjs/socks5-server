# go-socks5-proxy

![Latest tag from master branch](https://github.com/serjs/socks5-server/workflows/Latest%20tag%20from%20master%20branch/badge.svg)
![Release tag](https://github.com/serjs/socks5-server/workflows/Release%20tag/badge.svg)

Simple socks5 server using go-socks5 with authentication options

## Start container with proxy

```docker run -d --name socks5 -p 1080:1080 -e PROXY_USER=<PROXY_USER> -e PROXY_PASSWORD=<PROXY_PASSWORD>  serjs/go-socks5-proxy```

Leave `PROXY_USER` and `PROXY_PASSWORD` empty for skip authentication options while running socks5 server.

## List of all supported config parameters

|ENV variable|Type|Default|Description|
|------------|----|-------|-----------|
|PROXY_USER|String|EMPTY|Set proxy user (also required existed PROXY_PASS)|
|PROXY_PASSWORD|String|EMPTY|Set proxy password for auth, used with PROXY_USER|
|PROXY_PORT|String|1080|Set listen port for application inside docker container|
|TZ|String|UTC|Set Timezone like in many common Operation Systems|
|ALLOWED_IPS|String|Empty|Set allowed IP's that can connect to proxy, separator `,`|

`ALLOWED_IPS` parameter is not included in `serjs/go-socks5-proxy` image

## Test running service

Without authentication

```curl --socks5 <docker host ip>:1080  http://ifcfg.co``` - result must show docker host ip (for bridged network)

or

```docker run --rm curlimages/curl:7.65.3 -s --socks5 <docker host ip>:1080 http://ifcfg.co```

With authentication - result must show docker host ip (for bridged network)

```curl --socks5 <docker host ip>:1080 -U <PROXY_USER>:<PROXY_PASSWORD> http://ifcfg.co```

or

```docker run --rm curlimages/curl:7.65.3 -s --socks5 <PROXY_USER>:<PROXY_PASSWORD>@<docker host ip>:1080 http://ifcfg.co```

## Authors

* **Sergey Bogayrets**

See also the list of [contributors](https://github.com/serjs/socks5-server/graphs/contributors) who participated in this project.
