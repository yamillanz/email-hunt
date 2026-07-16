# email-hunt

Find and verify corporate email addresses from the terminal.

Given a person's name and a company domain, `email-hunt` generates 20+ possible email combinations and verifies each one via SMTP handshake — **without ever sending an email**. No APIs needed, no data shared with third parties. Fully offline by default.

## Prerequisites

- **Go 1.21+** installed on your machine.
- **Outbound access to port 25** (SMTP). Some ISPs, cloud providers, and corporate networks block this port. If blocked, SMTP verification will fail — the tool will report the error clearly.
- No API keys required for basic usage. AI providers are optional.

## Installation

```bash
go install github.com/yamillanz/email-hunt@latest
```

This downloads the source, compiles it, and places the `email-hunt` binary in your Go binary directory (typically `~/go/bin` on Linux/macOS or `%USERPROFILE%\go\bin` on Windows).

To install a specific version:

```bash
go install github.com/yamillanz/email-hunt@v1.0.0
```

Make sure `$GOPATH/bin` (or `~/go/bin`) is in your `$PATH`. Add this to your shell profile if needed:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Verify the installation:

```bash
email-hunt --version
```

## Quick start

```bash
# Basic lookup — generate and verify emails for John Doe at example.com
email-hunt "John Doe" example.com
```

This will:
1. Generate 18-25 deterministic email combinations (john@, jdoe@, john.doe@, etc.)
2. Classify each email (disposable? role-based?)
3. Look up the domain's MX mail server
4. Verify each email via SMTP handshake without sending any mail
5. Display a colored table with results

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `--ai` | `""` | AI provider for extra combinations: `openai` or `anthropic` |
| `--concurrency` | `5` | Number of parallel SMTP verifications |
| `--delay` | `500ms` | Delay between verifications per worker (e.g. `200ms`, `1s`) |
| `--help` | | Show usage, options, examples, and disclaimer |
| `--version` | | Show version number |

## Examples

```bash
# Show help and available options
email-hunt --help

# Basic lookup
email-hunt "Jane Smith" acmecorp.com

# Slower verification to avoid rate limits (2 workers, 1 second delay)
email-hunt --concurrency 2 --delay 1s "John Doe" example.com

# Fast verification with more workers and minimal delay
email-hunt --concurrency 10 --delay 200ms "Maria Garcia" empresa.es

# Generate extra combinations using OpenAI
EMAIL_HUNT_OPENAI_KEY=sk-your-key email-hunt --ai openai "Jane Smith" acmecorp.com

# Generate extra combinations using Anthropic
EMAIL_HUNT_ANTHROPIC_KEY=sk-ant-your-key email-hunt --ai anthropic "Jane Smith" acmecorp.com
```

## Understanding the output

The tool prints a table with four columns:

| Column | Description |
|--------|-------------|
| Status | Verification result with icon |
| Email | The email address tested |
| Category | Classification label |
| Detail | Server response or explanation |

### Status values

| Icon | Status | Color | Meaning |
|------|--------|-------|---------|
| ✓ | `valid` | Green | Mailbox exists — good to use |
| ✗ | `invalid` | Red | Mailbox does not exist |
| ~ | `catch_all` | Yellow | Domain accepts all emails — uncertain |
| ? | `unknown` | White | Could not determine (timeout, DNS error, blocked port) |

### Category values

| Label | Meaning |
|-------|---------|
| `standard` | Normal personal or corporate email |
| `disposable` | Temporary/throwaway address (Mailinator, GuerrillaMail, etc.) |
| `role_based` | Group or department address (admin@, info@, support@) |

### Example output

```
$ email-hunt "John Doe" acmecorp.com

Generating email combinations for "John Doe" at acmecorp.com...
Generated 18 deterministic combinations
Verifying with 5 workers, 500ms delay...

Status  Email                       Category    Detail
------  --------------------------  ----------  -----------------------------
✓ valid john.doe@acmecorp.com       standard    2.1.5 Ok
✗ invalid johndoe@acmecorp.com      standard    5.1.1 User unknown
✓ valid j.doe@acmecorp.com          standard    2.1.5 Ok
✗ invalid john@acmecorp.com         standard    5.1.1 User unknown
✗ invalid doe@acmecorp.com          standard    5.1.1 User unknown
~ catch_all jane@catchall.io        standard    domain accepts all email

Disclaimer: Use this information responsibly and only for legitimate purposes.
```

## AI Providers (optional)

When `--ai` is specified, the tool calls an LLM API to generate additional email combinations beyond the 18-25 deterministic patterns. Useful for unusual name formats or non-standard corporate email conventions.

### Setup

Export the API key for your chosen provider:

```bash
# OpenAI
export EMAIL_HUNT_OPENAI_KEY="sk-..."

# Anthropic
export EMAIL_HUNT_ANTHROPIC_KEY="sk-ant-..."
```

Then use the `--ai` flag:

```bash
email-hunt --ai openai "John Doe" example.com
```

The AI-generated emails are merged with the deterministic ones (duplicates are removed) and verified together.

### Supported providers

| Provider | Flag | Model | Env Variable |
|----------|------|-------|-------------|
| OpenAI | `--ai openai` | gpt-4o-mini | `EMAIL_HUNT_OPENAI_KEY` |
| Anthropic | `--ai anthropic` | claude-3-5-haiku-latest | `EMAIL_HUNT_ANTHROPIC_KEY` |

## How it works

```
email-hunt "John Doe" example.com
  │
  ├── 1. Generate
  │   Parse "John Doe" → first="john", last="doe"
  │   Generate 25 patterns: john@, doe@, johndoe@, john.doe@, jdoe@, j.doe@, ...
  │   Unicode normalization: "María" → "maria"
  │   Remove duplicates and malformed local parts
  │
  ├── 2. Classify (offline, no network)
  │   Detect disposable domains: mailinator.com, 10minutemail.com, ...
  │   Detect role-based addresses: admin@, info@, support@, ...
  │   Skip SMTP verification for disposable emails
  │
  ├── 3. DNS
  │   MX record lookup for the domain
  │   Find the primary mail server hostname
  │   If no MX records → report error and exit
  │
  ├── 4. Catch-all detection
  │   Send RCPT TO with a random email (tg7xk9p2@domain.com)
  │   If server accepts → domain is catch-all → mark all results
  │
  ├── 5. SMTP verification (5 workers, 500ms delay)
  │   For each email:
  │   │  Connect to mail server on port 25
  │   │  HELO email-hunt.local
  │   │  MAIL FROM: <verify@email-hunt.local>
  │   │  RCPT TO: <target@domain.com>
  │   │  Read response: 250 = valid, 550 = invalid
  │   │  QUIT (NEVER send DATA — no email transmitted)
  │
  └── 6. Output
      Colored terminal table with status, email, category, and detail
```

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Runtime error (generation failed, AI API error, etc.) |
| `2` | Invalid usage (wrong arguments, bad domain format) |

## Troubleshooting

**"connect to mail.example.com:25: i/o timeout"**
Port 25 is blocked by your ISP, VPN, or firewall. Try from a different network or use a VPS/dedicated server where port 25 is open. Some cloud providers (AWS, GCP, Azure) block port 25 by default.

**"no MX records for example.com"**
The domain has no mail server configured. This is common for parked domains or domains that use external email providers without proper MX setup. The tool cannot verify emails on domains without MX records.

**"EMAIL_HUNT_OPENAI_KEY environment variable not set"**
You used `--ai openai` without setting the API key. Export the key or remove the `--ai` flag to use offline-only mode.

## Limitations

- SMTP verification requires port 25 outbound access.
- Catch-all domains cannot be verified per-mailbox (the server accepts everything). Results for catch-all domains are marked with `~` and should be treated as uncertain.
- Some mail servers greylist unknown senders — the first connection may be rejected. Retrying after a few minutes can help.
- The deterministic algorithm covers ~90% of corporate email formats. Unusual patterns may not be generated. Use `--ai` for edge cases.

## Disclaimer

This tool performs SMTP verification without sending email. Use responsibly and only for legitimate B2B/recruiting purposes. Ensure compliance with GDPR, CAN-SPAM, and local regulations. Do not use for spam or harassment.
