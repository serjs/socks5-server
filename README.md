# go-socks5-proxy

![Latest tag from master branch](https://github.com/serjs/socks5-server/workflows/Latest%20tag%20from%20master%20branch/badge.svg)
![Release tag](https://github.com/serjs/socks5-server/workflows/Release%20tag/badge.svg)

Simple socks5 server using go-socks5 with authentication, allowed ips list and destination FQDNs filtering

# Examples

 - Clone this repo
```shell
git clone https://github.com/STmihan/socks5-server.git
cd socks5-server
```
 - Run docker compose
```shell
docker compose up -d
```
 - Test it
```shell
cmd # ony if you are using windows powershell
curl --socks5 localhost:1080 -U someuser:somepass http://ifcfg.co
```
 - If you want to use it without authentication, just remove `PROXY_USER` and `PROXY_PASSWORD` from `.env` file and restart container

# List of supported config parameters

|ENV variable|Type|Default|Description|
|------------|----|-------|-----------|
|PROXY_USER|String|EMPTY|Set proxy user (also required existed PROXY_PASS)|
|PROXY_PASSWORD|String|EMPTY|Set proxy password for auth, used with PROXY_USER|
|PROXY_PORT|String|1080|Set listen port for application inside docker container|
|PROXY_HEALTHCHECK_PORT|String|1081|Set listen port for healthcheck endpoint inside docker container|
|ALLOWED_DEST_FQDN|String|EMPTY|Allowed destination address regular expression pattern. Default allows all.|
|ALLOWED_IPS|String|Empty|Set allowed IP's that can connect to proxy, separator `,`|


# Build your own image:
`docker-compose -f docker-compose.build.yml up -d`\
Just don't forget to set parameters in the `.env` file.

# Test running service

Assuming that you are using container on 1080 host docker port

## Without authentication

```curl --socks5 <docker host ip>:1080  https://ifcfg.co``` - result must show docker host ip (for bridged network)

or

```docker run --rm curlimages/curl:7.65.3 -s --socks5 <docker host ip>:1080 https://ifcfg.co```

## With authentication

```curl --socks5 <docker host ip>:1080 -U <PROXY_USER>:<PROXY_PASSWORD> http://ifcfg.co```

or

```docker run --rm curlimages/curl:7.65.3 -s --socks5 <PROXY_USER>:<PROXY_PASSWORD>@<docker host ip>:1080 http://ifcfg.co```

# Authors

* **Sergey Bogayrets**
* **Mikhail Dunaev** (fork)

See also the list of [contributors](https://github.com/serjs/socks5-server/graphs/contributors) who participated in this project.
