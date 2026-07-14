package generator

import (
	"fmt"
	"strings"
)

func Generate(fullName, domain string) []string {
	parts := parseName(fullName)
	if parts.first == "" {
		return nil
	}

	patterns := buildPatterns()
	seen := make(map[string]bool)
	var emails []string

	for _, p := range patterns {
		local := p.fn(parts)
		if !validLocal(local) {
			continue
		}
		email := fmt.Sprintf("%s@%s", local, domain)
		if seen[email] {
			continue
		}
		seen[email] = true
		emails = append(emails, email)
	}

	return emails
}

func validLocal(local string) bool {
	if local == "" {
		return false
	}
	if len(local) > 64 {
		return false
	}
	first := local[0]
	last := local[len(local)-1]
	if first == '.' || first == '-' || first == '_' {
		return false
	}
	if last == '.' || last == '-' || last == '_' {
		return false
	}
	if strings.Contains(local, "..") || strings.Contains(local, "--") || strings.Contains(local, "__") {
		return false
	}
	return true
}
