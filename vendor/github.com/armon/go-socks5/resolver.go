package socks5

import (
	"context"
	"net"
	"net/netip"
)

// NameResolver is used to implement custom name resolution
type NameResolver interface {
	Resolve(ctx context.Context, name string) (context.Context, netip.Addr, error)
}

// DNSResolver uses the system DNS to resolve host names
type DNSResolver struct{}

func (d DNSResolver) Resolve(ctx context.Context, name string) (context.Context, netip.Addr, error) {
	ipAddr, err := net.ResolveIPAddr("ip", name)

	if err != nil {
		return ctx, netip.Addr{}, err
	}

	addr := netip.MustParseAddr(ipAddr.String())
	return ctx, addr, err
}
