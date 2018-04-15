FROM golang:latest as builder
WORKDIR /go/src/github.com/olebedev/socks5
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-s' -o ./socks5 

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/github.com/olebedev/socks5/socks5 /
ENTRYPOINT ["/socks5"]
