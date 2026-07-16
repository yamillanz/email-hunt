package main

import "testing"

func TestIsValidDomain(t *testing.T) {
	tests := []struct {
		domain string
		want   bool
	}{
		{"example.com", true},
		{"subdomain.example.com", true},
		{"example.co.uk", true},
		{"my-domain.com", true},
		{"xn--fsq.com", true},
		{"https://example.com", false},
		{"example.com/path", false},
		{"", false},
		{"-example.com", false},
		{"example-.com", false},
		{".example.com", false},
		{"example", false},
	}
	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got := isValidDomain(tt.domain)
			if got != tt.want {
				t.Errorf("isValidDomain(%q) = %v, want %v", tt.domain, got, tt.want)
			}
		})
	}
}
