package socks5

import (
	"context"
	"net/netip"
)

// NameResolver is used to implement custom name resolution
type NameResolver interface {
	Resolve(ctx context.Context, name string) (context.Context, netip.Addr, error)
}

// DNSResolver uses the system DNS to resolve host names
type DNSResolver struct{}

func (d DNSResolver) Resolve(ctx context.Context, name string) (context.Context, netip.Addr, error) {
	addr, err := netip.ParseAddr(name)

	if err != nil {
		return ctx, netip.Addr{}, err
	}
	return ctx, addr, err
}
