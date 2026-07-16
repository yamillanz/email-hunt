package output

import (
	"testing"

	"github.com/yamillanz/email-hunt/internal/classifier"
	"github.com/yamillanz/email-hunt/internal/verifier"
)

func TestPrintTableEmpty(t *testing.T) {
	PrintTable(nil)
	PrintTable([]verifier.Result{})
}

func TestPrintTableWithResults(t *testing.T) {
	results := []verifier.Result{
		{Email: "john.doe@example.com", Status: verifier.StatusValid, Detail: "OK"},
		{Email: "nobody@example.com", Status: verifier.StatusInvalid, Detail: "550 User unknown"},
		{Email: "jane@catchall.io", Status: verifier.StatusCatchAll, Detail: "domain accepts all"},
		{Email: "unknown@timeout.com", Status: verifier.StatusUnknown, Detail: "connection timeout"},
	}
	PrintTable(results)
}

func TestStatusWithColor(t *testing.T) {
	tests := []verifier.Status{
		verifier.StatusValid,
		verifier.StatusInvalid,
		verifier.StatusCatchAll,
		verifier.StatusUnknown,
	}
	for _, s := range tests {
		result := statusWithColor(s.String(), s)
		if result == "" {
			t.Errorf("statusWithColor returned empty for %v", s)
		}
	}
}

func TestCategoryWithColor(t *testing.T) {
	tests := []classifier.Category{
		classifier.CategoryStandard,
		classifier.CategoryDisposable,
		classifier.CategoryRoleBased,
	}
	for _, c := range tests {
		result := categoryWithColor(c.String(), c)
		if result == "" {
			t.Errorf("categoryWithColor returned empty for %v", c)
		}
	}
}
