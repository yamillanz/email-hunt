package main

import (
	"flag"
	"fmt"
	"os"
)

var version = "dev"

func main() {
	aiProvider := flag.String("ai", "", "AI provider for additional combinations (openai, anthropic)")
	concurrency := flag.Int("concurrency", 5, "Number of concurrent verifications")
	delayMs := flag.Int("delay", 500, "Delay between verifications in milliseconds")
	showVersion := flag.Bool("version", false, "Show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "email-hunt — Find and verify corporate email addresses\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  email-hunt [flags] \"Full Name\" domain.com\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  email-hunt \"John Doe\" example.com\n")
		fmt.Fprintf(os.Stderr, "  email-hunt --ai openai \"Jane Smith\" acmecorp.com\n")
		fmt.Fprintf(os.Stderr, "  email-hunt --concurrency 3 --delay 500ms \"John Doe\" example.com\n")
		fmt.Fprintf(os.Stderr, "\nDisclaimer:\n")
		fmt.Fprintf(os.Stderr, "  This tool performs SMTP verification without sending email.\n")
		fmt.Fprintf(os.Stderr, "  Use responsibly and only for legitimate B2B/recruiting purposes.\n")
		fmt.Fprintf(os.Stderr, "  Ensure compliance with GDPR, CAN-SPAM, and local regulations.\n")
		fmt.Fprintf(os.Stderr, "  Do not use for spam or harassment.\n")
	}

	flag.Parse()

	if *showVersion {
		fmt.Println("email-hunt version", version)
		return
	}

	_ = aiProvider
	_ = concurrency
	_ = delayMs

	fmt.Println("email-hunt", version)
}
