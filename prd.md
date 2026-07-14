# PRD: email-hunt — Email Finder & Verifier CLI

## Problem Statement

Finding a person's corporate email to send a CV or network requires manually testing combinations in Gmail or using paid SaaS services (Hunter.io, Tomba, Snov.io) that share your data with third parties. There is no open-source, offline-first tool that generates email combinations and verifies them via SMTP in a single step from the terminal.

## Solution

A Go CLI, installable via `go install`, that takes a name and a domain, generates 20+ possible email combinations using a deterministic algorithm (with optional AI fallback), verifies each one via SMTP handshake (without sending a real email) using automatic MX lookup, and presents the results in a terminal table showing which are valid, catch-all, disposable, or role-based.

The user runs:

```
email-hunt "John Doe" example.com
```

And gets a table with the status of each verified email. No data shared with third parties. No per-use cost.

## User Stories

1. **As a recruiter/candidate**, when I run `email-hunt "Jane Smith" acmecorp.com`, the system generates at least 20 email combinations (jane.smith@, jsmith@, j.smith@, etc.) and verifies each via SMTP, showing a table with the status of each email (valid, invalid, catch-all, disposable, role-based).

2. **As an advanced user**, when I want to improve the accuracy of the combinations, I can enable AI mode with `--ai openai` and the system uses an LLM to generate additional or more creative combinations, requiring the API key via environment variable (`EMAIL_HUNT_OPENAI_KEY`).

3. **As a user concerned about being blocked**, when I verify emails against a domain, the system applies a configurable delay between requests and limits concurrency (`--concurrency 3 --delay 500ms`) to avoid being rate-limited by the target SMTP server.

4. **As a user who wants to audit results**, when verification finishes, the system distinguishes between "valid" (SMTP 250), "invalid" (SMTP 550), "catch-all" (domain accepts everything), "disposable" (temporary domain), and "role-based" (admin@, info@) emails, and shows a disclaimer about ethical use.

## Constraints

- **Zero dependency on paid external APIs** in the default flow. The deterministic algorithm must work 100% offline.
- SMTP verification **must never send a real email** (MAIL FROM + RCPT TO only, never DATA).
- AI API keys must only be read from **environment variables**, never from command-line flags or config files.
- The binary must be installable with a single command: `go install github.com/yamillanz/email-hunt@latest`.
- Primary output to stdout must be a readable table. Any logging/progress goes to stderr.
- The project must compile with `CGO_ENABLED=0` to produce static binaries.

## Acceptance Criteria

- [ ] `email-hunt "John Doe" example.com` generates >=20 email combinations and shows a table with SMTP verification results.
- [ ] `email-hunt --help` shows usage, available flags, and legal/ethical disclaimer.
- [ ] `email-hunt --version` shows the binary version.
- [ ] `email-hunt "Maria Jose Garcia" empresa.es` correctly handles names with accents, n~, and spaces (Unicode normalization).
- [ ] `email-hunt --ai openai "John Doe" example.com` generates additional combinations using the OpenAI API if `EMAIL_HUNT_OPENAI_KEY` is set.
- [ ] `email-hunt --ai anthropic "John Doe" example.com` works analogously with `EMAIL_HUNT_ANTHROPIC_KEY`.
- [ ] `email-hunt --concurrency 5 --delay 500ms "John Doe" example.com` verifies with a maximum of 5 concurrent connections and 500ms between each batch.
- [ ] If the domain has no MX records, the system reports a clear error and does not attempt verification.
- [ ] Results classify each email as: `valid`, `invalid`, `catch_all`, `disposable`, `role_based`, `unknown`.
- [ ] Verification detects catch-all domains (by testing a random non-existent email on the domain).
- [ ] Verification detects disposable email services (Mailinator, GuerrillaMail, 10MinuteMail, etc.) using an embedded list.
- [ ] Verification detects role-based emails (admin@, info@, support@, etc.) and flags them as such.
- [ ] Unit tests cover pattern generation (>=90% coverage in the generation package).
- [ ] Unit tests cover SMTP verification with a mock server.
- [ ] `go install github.com/yamillanz/email-hunt@latest` correctly installs the binary in `$GOPATH/bin`.

## Out of Scope

- **Batch/CSV processing** (multiple people in a file) — phase 2.
- **Homebrew Tap / pre-compiled binaries** — only `go install` in v1. GoReleaser in v2.
- **Verification via third-party APIs** (Hunter.io, Tomba, ZeroBounce) — direct SMTP only.
- **Actual email sending** (SMTP DATA) — never in any phase.
- **Advanced per-domain rate limiting** (persisting state across runs) — only configurable delay per execution.
- **Auto-detection of domain email pattern** (inferring the corporate format like Hunter.io does) — v2 if there is demand.
- **Data enrichment** (company name, job title, LinkedIn) — permanently out of scope.
- **Interactive mode / TUI** — single-execution CLI only.

## Assumptions

- The user has Go 1.21+ installed for `go install`.
- The user has access to port 25 (or 2525/587 as fallback) for SMTP. In environments where it is blocked (cloud, some ISPs), SMTP verification will fail and the user will be notified. In v1 we assume the user knows they need outbound SMTP connectivity.
- The name is passed as a single string `"First Last"`. If it has more than two words, the first is `first` and the last is `last`, with intermediate words treated as middle names.
- The domain is passed without a protocol (only `example.com`, not `https://example.com`).
- AI API keys are optional. If not configured, the `--ai` flag shows a descriptive error with instructions.
- The project is hosted on GitHub under `github.com/yamillanz/email-hunt`.
- The disposable domain list will be embedded in the binary (go:embed) and kept updated in the repository.
- SMTP verification will use aggressive timeouts (5s per attempt) to avoid hanging on slow servers.
- The stdlib `flag` package will be used for the CLI (2 main positional args + optional flags). If subcommands are added in the future, it will be migrated to Cobra.

## Implementation Notes

- Consider separating the code into `internal/generator` (pattern generation), `internal/verifier` (SMTP + MX lookup), `internal/ai` (AI providers), `internal/output` (table formatting), and `internal/classifier` (disposable/role-based/catch-all detection).
- For deterministic generation, Mailfoguess (Python) and guessmail (Ruby) are good pattern references. The initial set of ~22 patterns should cover >90% of corporate formats.
- For SMTP verification, `truemail-go` and `mailsherpa` (Go) are excellent references. Evaluate whether to use `truemail-go` as a dependency or implement the SMTP handshake using `net/smtp` from stdlib.
- For the spinner/progress indicator, evaluate `briandowns/spinner` vs a custom implementation using `\r` to keep dependencies minimal.
- For table output, `olekukoneko/tablewriter` is lightweight and popular. Alternative: `rodaine/table`.
- For terminal colors, `fatih/color` is the de facto standard in Go.
- AI providers should implement a common interface (`ai.Provider`) to make adding new providers easy.
- Catch-all detection requires sending an RCPT TO with a random email (UUID)@domain. If the server responds 250, the domain is catch-all.
- Include a disclaimer message in the output and in `--help`: "This tool performs SMTP verification without sending email. Use responsibly and only for legitimate B2B/recruiting purposes. Ensure compliance with GDPR, CAN-SPAM, and local regulations. Do not use for spam or harassment."
- The disposable domain list can be maintained in `internal/classifier/disposable_domains.txt` and embedded with `//go:embed`.
