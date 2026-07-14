package verifier

import "testing"

func TestLookupMX(t *testing.T) {
	tests := []struct {
		name   string
		domain string
		wantOK bool
	}{
		{"real domain with MX", "gmail.com", true},
		{"real domain with MX", "outlook.com", true},
		{"nonexistent domain", "this-domain-does-not-exist-xyz123.com", false},
		{"invalid domain format", "not a domain", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records, err := LookupMX(tt.domain)
			if tt.wantOK {
				if err != nil {
					t.Errorf("LookupMX(%q) returned error: %v", tt.domain, err)
					return
				}
				if len(records) == 0 {
					t.Errorf("LookupMX(%q) returned zero records", tt.domain)
				}
			} else {
				if err == nil {
					t.Errorf("LookupMX(%q) should have failed but returned: %v", tt.domain, records)
				}
			}
		})
	}
}

func TestMXPrioritySorted(t *testing.T) {
	records, err := LookupMX("gmail.com")
	if err != nil {
		t.Skipf("skipping: DNS lookup failed: %v", err)
	}
	if len(records) < 2 {
		t.Skip("need at least 2 MX records to test priority sort")
	}
	for i := 1; i < len(records); i++ {
		if records[i-1].Pref > records[i].Pref {
			t.Errorf("MX records not sorted by priority: pref %d before %d",
				records[i-1].Pref, records[i].Pref)
		}
	}
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusValid, "valid"},
		{StatusInvalid, "invalid"},
		{StatusCatchAll, "catch_all"},
		{StatusUnknown, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.status.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
