package main

import (
	"regexp"

	"github.com/armon/go-socks5"
	"golang.org/x/net/context"
)

// PermitDestAddrPattern returns a RuleSet which selectively allows addresses
func PermitDestAddrPattern(pattern string) socks5.RuleSet {
	return &PermitDestAddrPatternRuleSet{pattern}
}

// PermitDestAddrPatternRuleSet is an implementation of the RuleSet which
// enables filtering supported destination address
type PermitDestAddrPatternRuleSet struct {
	AllowedFqdnPattern string
}

func (p *PermitDestAddrPatternRuleSet) Allow(ctx context.Context, req *socks5.Request) (context.Context, bool) {
    var match bool
    if req.DestAddr.FQDN != "" {
        match, _ = regexp.MatchString(p.AllowedFqdnPattern, req.DestAddr.FQDN)
    } else if req.DestAddr.IP != nil {
        match, _ = regexp.MatchString(p.AllowedFqdnPattern, req.DestAddr.IP.String())
    } else {
	match = true
    }
    return ctx, match
}
