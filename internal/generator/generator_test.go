package generator

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name     string
		fullName string
		domain   string
		wantLen  int
		want     []string
	}{
		{
			name:     "simple first+last",
			fullName: "John Doe",
			domain:   "example.com",
			wantLen:  18,
			want: []string{
				"john@example.com",
				"doe@example.com",
				"johndoe@example.com",
				"john.doe@example.com",
				"john_doe@example.com",
				"john-doe@example.com",
				"jdoe@example.com",
				"j.doe@example.com",
			},
		},
		{
			name:     "with middle name",
			fullName: "John Michael Doe",
			domain:   "example.com",
			wantLen:  25,
			want: []string{
				"john.m.doe@example.com",
				"jmdoe@example.com",
				"johnmdoe@example.com",
				"johnmichaeldoe@example.com",
			},
		},
		{
			name:     "single name only",
			fullName: "John",
			domain:   "example.com",
			wantLen:  2,
			want: []string{
				"john@example.com",
			},
		},
		{
			name:     "empty name",
			fullName: "",
			domain:   "example.com",
			wantLen:  0,
		},
		{
			name:     "unicode normalization",
			fullName: "Maria Jose Garcia",
			domain:   "empresa.es",
			want: []string{
				"maria@empresa.es",
				"garcia@empresa.es",
				"mariagarcia@empresa.es",
				"maria.garcia@empresa.es",
				"m.garcia@empresa.es",
			},
		},
		{
			name:     "no duplicates",
			fullName: "A B",
			domain:   "example.com",
			wantLen:  8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Generate(tt.fullName, tt.domain)

			if tt.wantLen > 0 && len(got) != tt.wantLen {
				t.Errorf("Generate() returned %d emails, want %d: %v", len(got), tt.wantLen, got)
			}

			if len(tt.want) > 0 {
				gotSet := make(map[string]bool)
				for _, email := range got {
					gotSet[email] = true
				}
				for _, expected := range tt.want {
					if !gotSet[expected] {
						t.Errorf("Generate() missing expected email: %s (got: %v)", expected, got)
					}
				}
			}
		})
	}
}

func TestParseName(t *testing.T) {
	tests := []struct {
		name      string
		fullName  string
		wantFirst string
		wantLast  string
	}{
		{"simple", "John Doe", "john", "doe"},
		{"three words", "John Michael Doe", "john", "doe"},
		{"unicode", "Maria Jose Garcia", "maria", "garcia"},
		{"single word", "John", "john", ""},
		{"empty", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := parseName(tt.fullName)
			if parts.first != tt.wantFirst {
				t.Errorf("first = %q, want %q", parts.first, tt.wantFirst)
			}
			if parts.last != tt.wantLast {
				t.Errorf("last = %q, want %q", parts.last, tt.wantLast)
			}
		})
	}
}

func TestPatternsCoverage(t *testing.T) {
	parts := parseName("John Michael Doe")

	for _, p := range buildPatterns() {
		result := p.fn(parts)
		if result == "" {
			t.Errorf("pattern %q produced empty string", p.name)
		}
	}
}

func TestGenerateNoDuplicates(t *testing.T) {
	emails := Generate("John Doe", "example.com")
	seen := make(map[string]bool)
	for _, email := range emails {
		if seen[email] {
			t.Errorf("duplicate email generated: %s", email)
		}
		seen[email] = true
	}
}

func TestGenerateResultOrder(t *testing.T) {
	got := Generate("John Doe", "example.com")
	want := []string{
		"john@example.com",
		"doe@example.com",
		"johndoe@example.com",
		"john.doe@example.com",
		"john_doe@example.com",
		"john-doe@example.com",
		"jdoe@example.com",
	}
	for i, expected := range want {
		if i >= len(got) || got[i] != expected {
			t.Errorf("position %d: got %q, want %q", i, got[i], expected)
		}
	}
}

func TestValidLocal(t *testing.T) {
	tests := []struct {
		local string
		want  bool
	}{
		{"john", true},
		{"j.doe", true},
		{"j.d", true},
		{"j..doe", false},
		{"john.", false},
		{".john", false},
		{"john-", false},
		{"-john", false},
		{"john_", false},
		{"_john", false},
		{"john--doe", false},
		{"john__doe", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.local, func(t *testing.T) {
			got := validLocal(tt.local)
			if got != tt.want {
				t.Errorf("validLocal(%q) = %v, want %v", tt.local, got, tt.want)
			}
		})
	}
}
