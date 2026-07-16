package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/yamillanz/email-hunt/internal/generator"
	"github.com/yamillanz/email-hunt/internal/verifier"
)

var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	aiProvider := flag.String("ai", "", "AI provider for additional combinations (openai, anthropic)")
	concurrency := flag.Int("concurrency", 5, "Number of concurrent verifications")
	delay := flag.Duration("delay", 500*time.Millisecond, "Delay between verifications (e.g. 500ms, 1s)")
	showVersion := flag.Bool("version", false, "Show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "email-hunt — Find and verify corporate email addresses\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  email-hunt [flags] \"Full Name\" domain.com\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  email-hunt \"John Doe\" example.com\n")
		fmt.Fprintf(os.Stderr, "  email-hunt --ai openai \"Jane Smith\" acmecorp.com\n")
		fmt.Fprintf(os.Stderr, "  email-hunt --concurrency 3 --delay 1s \"John Doe\" example.com\n")
		fmt.Fprintf(os.Stderr, "\nDisclaimer:\n")
		fmt.Fprintf(os.Stderr, "  This tool performs SMTP verification without sending email.\n")
		fmt.Fprintf(os.Stderr, "  Use responsibly and only for legitimate B2B/recruiting purposes.\n")
		fmt.Fprintf(os.Stderr, "  Ensure compliance with GDPR, CAN-SPAM, and local regulations.\n")
		fmt.Fprintf(os.Stderr, "  Do not use for spam or harassment.\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Println("email-hunt version", version)
		return 0
	}

	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Error: requires exactly 2 arguments (name and domain)\n\n")
		flag.Usage()
		return 2
	}

	fullName := args[0]
	domain := args[1]

	if !isValidDomain(domain) {
		fmt.Fprintf(os.Stderr, "Error: invalid domain format: %s\n", domain)
		return 2
	}

	if *concurrency < 1 {
		fmt.Fprintf(os.Stderr, "Error: --concurrency must be at least 1\n")
		return 2
	}

	if *delay < 0 {
		fmt.Fprintf(os.Stderr, "Error: --delay cannot be negative\n")
		return 2
	}

	if *aiProvider != "" && *aiProvider != "openai" && *aiProvider != "anthropic" {
		fmt.Fprintf(os.Stderr, "Error: unknown AI provider: %s (valid: openai, anthropic)\n", *aiProvider)
		return 2
	}

	fmt.Fprintf(os.Stderr, "Generating email combinations for %q at %s...\n", fullName, domain)
	emails := generator.Generate(fullName, domain)
	if len(emails) == 0 {
		fmt.Fprintf(os.Stderr, "Error: could not generate any email combinations\n")
		return 1
	}
	fmt.Fprintf(os.Stderr, "Generated %d combinations\n", len(emails))

	// TODO: Phase 8 — AI provider: append additional combinations here

	fmt.Fprintf(os.Stderr, "Verifying with %d workers, %s delay...\n", *concurrency, *delay)
	ctx := context.Background()
	results := verifier.VerifyAll(ctx, emails, *concurrency, *delay)

	printResults(results)
	return 0
}

func printResults(results []verifier.Result) {
	fmt.Println()
	fmt.Println("Results:")
	fmt.Println("-------")

	for _, r := range results {
		statusIcon := "?"
		switch r.Status {
		case verifier.StatusValid:
			statusIcon = "✓"
		case verifier.StatusInvalid:
			statusIcon = "✗"
		case verifier.StatusCatchAll:
			statusIcon = "~"
		}
		category := ""
		if r.Category.String() != "standard" {
			category = " [" + r.Category.String() + "]"
		}
		fmt.Printf("  %s %s — %s%s\n", statusIcon, r.Email, r.Status, category)
	}
	fmt.Println()
	fmt.Println("Disclaimer: Use this information responsibly and only for legitimate purposes.")
}

var domainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)+$`)

func isValidDomain(domain string) bool {
	return domainRegex.MatchString(domain)
}
