# email-hunt

Find and verify corporate email addresses from the terminal.

Given a person's name and a company domain, `email-hunt` generates 20+ possible email combinations and verifies each one via SMTP handshake — **without ever sending an email**. No APIs, no data shared with third parties. Fully offline by default.

## Install

```bash
go install github.com/yamillanz/email-hunt@latest
```

Requires Go 1.21+.

## Usage

```bash
email-hunt "John Doe" example.com
```

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `--ai` | `""` | AI provider for additional combinations (`openai`, `anthropic`) |
| `--concurrency` | `5` | Number of concurrent SMTP verifications |
| `--delay` | `500ms` | Delay between verifications per worker (`1s`, `200ms`) |
| `--version` | | Show version |

## Examples

```bash
# Basic lookup
email-hunt "Jane Smith" acmecorp.com

# With AI-generated combinations (requires API key)
EMAIL_HUNT_OPENAI_KEY=sk-... email-hunt --ai openai "Jane Smith" acmecorp.com

# Slower verification to avoid rate limits
email-hunt --concurrency 2 --delay 1s "John Doe" example.com
```

## AI Providers

Set the environment variable for your provider:

| Provider | Env variable |
|----------|-------------|
| OpenAI | `EMAIL_HUNT_OPENAI_KEY` |
| Anthropic | `EMAIL_HUNT_ANTHROPIC_KEY` |

## How it works

1. **Generate** — 25 deterministic email patterns from the name (first.last, f.last, jdoe, etc.)
2. **Classify** — detect disposable domains (Mailinator, etc.) and role-based addresses (admin@, info@)
3. **DNS** — MX record lookup to find the mail server
4. **Verify** — SMTP handshake: `HELO → MAIL FROM → RCPT TO → QUIT`. Mailbox exists if RCPT TO returns 250. Never sends DATA — no email is transmitted.
5. **Output** — colored terminal table with status per email

## Output

```
email-hunt "John Doe" example.com
```

```
Status  Email                       Category    Detail
------  --------------------------  ----------  -------------------
✓ valid john.doe@example.com        standard    2.1.5 Ok
✗ invalid nobody@example.com        standard    5.1.1 User unknown
~ catch_all jane@catchall.io        standard    domain accepts all email
✓ valid j.doe@example.com           role_based  2.1.5 Ok
```

## Disclaimer

This tool performs SMTP verification without sending email. Use responsibly and only for legitimate B2B/recruiting purposes. Ensure compliance with GDPR, CAN-SPAM, and local regulations. Do not use for spam or harassment.

## Requirements

- Outbound access to port 25 (SMTP). Some ISPs and cloud providers block this. SMTP verification will fail if port 25 is unreachable.
- No external API keys required (offline mode). AI providers are optional and require their respective API keys.
