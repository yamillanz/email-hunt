package verifier

import (
	"fmt"
	"net"
)

func LookupMX(domain string) ([]*net.MX, error) {
	records, err := net.LookupMX(domain)
	if err != nil {
		return nil, fmt.Errorf("no mail servers found for %s: %w", domain, err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no MX records for %s", domain)
	}
	return records, nil
}
