package main

import (
	"log"
	"regexp"

	"github.com/armon/go-socks5"
	"golang.org/x/net/context"
)

// PermitDestAddrPattern returns a RuleSet which selectively allows addresses
func PermitDestAddrPattern(pattern string) socks5.RuleSet {
	// Валидируем regex при инициализации
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatalf("Некорректный regex паттерн в ALLOWED_DEST_FQDN: %v", err)
	}
	return &PermitDestAddrPatternRuleSet{
		AllowedFqdnPattern: pattern,
		compiledRegex:      compiled,
	}
}

// PermitDestAddrPatternRuleSet is an implementation of the RuleSet which
// enables filtering supported destination address
type PermitDestAddrPatternRuleSet struct {
	AllowedFqdnPattern string
	compiledRegex      *regexp.Regexp // Кэшируем откомпилированный regex для производительности
}

func (p *PermitDestAddrPatternRuleSet) Allow(ctx context.Context, req *socks5.Request) (context.Context, bool) {
	match := p.compiledRegex.MatchString(req.DestAddr.FQDN)
	if !match {
		log.Printf("[FQDN FILTER] Отклонено соединение к %s (не соответствует паттерну)", req.DestAddr.FQDN)
	}
	return ctx, match
}
