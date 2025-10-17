package main

import (
	"net"
	"regexp"
	"strings"

	"github.com/armon/go-socks5"
	"golang.org/x/net/context"
)

// PermitDestAddrPattern returns a RuleSet which selectively allows addresses
// PermitDestAddrPatternCompiled compiles the provided regex and returns a RuleSet
// which matches request FQDNs against that compiled regex.
func PermitDestAddrPatternCompiled(pattern string) (socks5.RuleSet, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &PermitDestAddrPatternRuleSet{r}, nil
}

// PermitDestAddrPatternRuleSet is an implementation of the RuleSet which
// enables filtering supported destination address by FQDN using a compiled regexp
type PermitDestAddrPatternRuleSet struct {
	re *regexp.Regexp
}

func (p *PermitDestAddrPatternRuleSet) Allow(ctx context.Context, req *socks5.Request) (context.Context, bool) {
	if req == nil || req.DestAddr == nil {
		return ctx, false
	}
	return ctx, p.re.MatchString(req.DestAddr.FQDN)
}

// PermitDestAddrIP returns a RuleSet which allows requests whose resolved
// destination IP (or direct IP) matches any of the provided ipOrCIDR entries.
// entries can be individual IPs (e.g. "192.0.2.1") or CIDR blocks (e.g. "10.0.0.0/8").
func PermitDestAddrIP(entries []string) (socks5.RuleSet, error) {
	var cidrs []*net.IPNet
	var ips []net.IP
	for _, e := range entries {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		if strings.Contains(e, "/") {
			_, network, err := net.ParseCIDR(e)
			if err != nil {
				return nil, err
			}
			cidrs = append(cidrs, network)
		} else {
			ip := net.ParseIP(e)
			if ip == nil {
				return nil, &net.ParseError{Type: "IP address", Text: e}
			}
			ips = append(ips, ip)
		}
	}
	return &PermitDestAddrIPRuleSet{cidrs: cidrs, ips: ips}, nil
}

// PermitDestAddrIPRuleSet checks destination IP against configured IPs/CIDRs
type PermitDestAddrIPRuleSet struct {
	cidrs []*net.IPNet
	ips   []net.IP
}

func (p *PermitDestAddrIPRuleSet) Allow(ctx context.Context, req *socks5.Request) (context.Context, bool) {
	if req == nil || req.DestAddr == nil {
		return ctx, false
	}
	// prefer realDestAddr if set (but in RuleSet we only have req.DestAddr).
	// Use DestAddr.IP if available; otherwise attempt to match on FQDN -> not performed here.
	if len(req.DestAddr.IP) != 0 {
		for _, ip := range p.ips {
			if ip.Equal(req.DestAddr.IP) {
				return ctx, true
			}
		}
		for _, netw := range p.cidrs {
			if netw.Contains(req.DestAddr.IP) {
				return ctx, true
			}
		}
	}
	// If DestAddr.IP is empty (client provided FQDN), do not resolve here; let server's resolver fill IP
	// Note: go-socks5 resolves FQDN earlier and fills DestAddr.IP before dialing, but the RuleSet
	// is called before resolution in some flows; so this rule will only match if IP is present.
	return ctx, false
}

// CombineRuleSets returns a RuleSet that permits a request if any of the provided
// rules permits it (logical OR).
func CombineRuleSets(a, b socks5.RuleSet) socks5.RuleSet {
	return &combinedRuleSet{a: a, b: b}
}

type combinedRuleSet struct {
	a socks5.RuleSet
	b socks5.RuleSet
}

func (c *combinedRuleSet) Allow(ctx context.Context, req *socks5.Request) (context.Context, bool) {
	if ctx_, ok := c.a.Allow(ctx, req); ok {
		return ctx_, true
	}
	return c.b.Allow(ctx, req)
}
