package output

import (
	"os"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/yamillanz/email-hunt/internal/classifier"
	"github.com/yamillanz/email-hunt/internal/verifier"
)

var (
	clrValid    = color.New(color.FgGreen).SprintFunc()
	clrInvalid  = color.New(color.FgRed).SprintFunc()
	clrCatchAll = color.New(color.FgYellow).SprintFunc()
	clrRole     = color.New(color.FgCyan).SprintFunc()
	clrDisposable = color.New(color.FgHiBlack).SprintFunc()
	clrUnknown  = color.New(color.FgWhite).SprintFunc()
)

func PrintTable(results []verifier.Result) {
	if len(results) == 0 {
		return
	}

	tbl := table.New("Status", "Email", "Category", "Detail")
	tbl.WithWriter(os.Stdout)

	for _, r := range results {
		tbl.AddRow(
			statusWithColor(r.Status.String(), r.Status),
			r.Email,
			categoryWithColor(r.Category.String(), r.Category),
			r.Detail,
		)
	}

	tbl.Print()
}

func statusWithColor(label string, s verifier.Status) string {
	switch s {
	case verifier.StatusValid:
		return clrValid("✓ " + label)
	case verifier.StatusInvalid:
		return clrInvalid("✗ " + label)
	case verifier.StatusCatchAll:
		return clrCatchAll("~ " + label)
	default:
		return clrUnknown("? " + label)
	}
}

func categoryWithColor(label string, c classifier.Category) string {
	switch c {
	case classifier.CategoryDisposable:
		return clrDisposable(label)
	case classifier.CategoryRoleBased:
		return clrRole(label)
	default:
		return label
	}
}
