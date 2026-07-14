// Package classifier detects disposable domains, role-based addresses,
// and classifies email addresses without network access.
package classifier

import (
	"strings"

	"net/mail"
)

type Category int

const (
	CategoryStandard   Category = iota
	CategoryDisposable          // known temporary/throwaway email domain
	CategoryRoleBased           // group address (admin@, info@, etc.)
)

func (c Category) String() string {
	switch c {
	case CategoryDisposable:
		return "disposable"
	case CategoryRoleBased:
		return "role_based"
	default:
		return "standard"
	}
}

func Classify(email string) Category {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return CategoryStandard
	}

	local, _, _ := strings.Cut(addr.Address, "@")
	domain := extractDomain(addr.Address)

	if IsDisposable(domain) {
		return CategoryDisposable
	}
	if IsRoleBased(local) {
		return CategoryRoleBased
	}
	return CategoryStandard
}

func IsDisposable(domain string) bool {
	return disposableSet[strings.ToLower(domain)]
}

func IsRoleBased(localPart string) bool {
	lower := strings.ToLower(localPart)
	return rolePrefixes[lower]
}

func extractDomain(addr string) string {
	_, domain, found := strings.Cut(addr, "@")
	if !found {
		return ""
	}
	return strings.ToLower(domain)
}
