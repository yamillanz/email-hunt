package classifier

import "testing"

func TestIsDisposable(t *testing.T) {
	tests := []struct {
		domain string
		want   bool
	}{
		{"mailinator.com", true},
		{"MAILINATOR.COM", true},
		{"guerrillamail.com", true},
		{"yopmail.com", true},
		{"10minutemail.com", true},
		{"gmail.com", false},
		{"outlook.com", false},
		{"acmecorp.com", false},
		{"example.com", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got := IsDisposable(tt.domain)
			if got != tt.want {
				t.Errorf("IsDisposable(%q) = %v, want %v", tt.domain, got, tt.want)
			}
		})
	}
}

func TestIsRoleBased(t *testing.T) {
	tests := []struct {
		local string
		want  bool
	}{
		{"admin", true},
		{"Admin", true},
		{"ADMIN", true},
		{"info", true},
		{"support", true},
		{"sales", true},
		{"no-reply", true},
		{"noreply", true},
		{"john.doe", false},
		{"jdoe", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.local, func(t *testing.T) {
			got := IsRoleBased(tt.local)
			if got != tt.want {
				t.Errorf("IsRoleBased(%q) = %v, want %v", tt.local, got, tt.want)
			}
		})
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  Category
	}{
		{"standard corporate", "john.doe@acmecorp.com", CategoryStandard},
		{"disposable", "test@mailinator.com", CategoryDisposable},
		{"role-based", "admin@acmecorp.com", CategoryRoleBased},
		{"disposable role", "admin@mailinator.com", CategoryDisposable},
		{"invalid email", "not-an-email", CategoryStandard},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Classify(tt.email)
			if got != tt.want {
				t.Errorf("Classify(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestCategoryString(t *testing.T) {
	tests := []struct {
		cat  Category
		want string
	}{
		{CategoryStandard, "standard"},
		{CategoryDisposable, "disposable"},
		{CategoryRoleBased, "role_based"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.cat.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDisposableListNotEmpty(t *testing.T) {
	if len(disposableSet) < 10 {
		t.Errorf("disposableSet has only %d entries, expected at least 10", len(disposableSet))
	}
}
