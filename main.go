package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"
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

	fmt.Fprintf(os.Stderr, "Looking for %s at %s...\n", fullName, domain)

	// TODO: Phase 2 — generate email combinations
	// TODO: Phase 3 — classify domain/emails
	// TODO: Phase 4-5 — MX lookup + SMTP verification
	// TODO: Phase 6 — concurrent verification
	// TODO: Phase 7 — format and display results
	// TODO: Phase 8 — AI provider integration

	_ = concurrency
	_ = delay
	_ = aiProvider

	return 0
}

var domainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)+$`)

func isValidDomain(domain string) bool {
	return domainRegex.MatchString(domain)
}
