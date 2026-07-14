# Implementation Plan: email-hunt

## Phase Order & Dependencies

```
Phase 0: Init  ─────────────────────────────────────────┐
Phase 1: CLI  ──────────────────────────────────────────┤   No dependencies
Phase 2: Generator  ────────────────────────────────────┘

Phase 3: Classifier ───────────┐
Phase 4: MX Lookup ────────────┤   No dependencies on
Phase 5: SMTP Engine ──────────┘   each other

Phase 6: Concurrency ── depends on Phase 5 (needs Verify)
Phase 7: Output ─────── depends on Phase 5 (needs Result type)
Phase 8: AI ─────────── depends on Phase 2 (needs pattern concept)

Phase 9: Testing ───── runs against all phases
Phase 10: CI & Dist ── final step
```

Phases 3-5 can be built in parallel if multiple people work on this. For a solo developer, sequential is fine — each phase builds on the last.

---

## Final Project Tree

```
email-hunt/
├── go.mod
├── go.sum
├── main.go
├── .goreleaser.yaml
├── .github/
│   └── workflows/
│       └── test.yml
├── README.md
└── internal/
    ├── generator/
    │   ├── generator.go
    │   ├── generator_test.go
    │   ├── patterns.go
    │   └── name.go
    ├── verifier/
    │   ├── verifier.go
    │   ├── verifier_test.go
    │   ├── dns.go
    │   ├── smtp.go
    │   └── result.go
    ├── classifier/
    │   ├── classifier.go
    │   ├── classifier_test.go
    │   ├── disposable_domains.txt
    │   └── roles.go
    ├── ai/
    │   ├── provider.go
    │   ├── openai.go
    │   ├── anthropic.go
    │   └── provider_test.go
    └── output/
        ├── table.go
        └── colors.go
```

---

## Phase 0: Project Initialization & Skeleton

### What we build

A Go module, the directory tree, and a `main.go` that compiles but does nothing yet.

### Structure created

```
email-hunt/
├── go.mod
├── main.go
└── internal/
    ├── generator/
    │   └── doc.go
    ├── verifier/
    │   └── doc.go
    ├── classifier/
    │   └── doc.go
    ├── ai/
    │   └── doc.go
    └── output/
        └── doc.go
```

### Why each decision

#### `go.mod` with module path `github.com/yamillanz/email-hunt`

Go needs a module path to identify your project. This path doubles as the import path for `go install`. It must match where the code lives (GitHub repo URL). Without this, `go install github.com/yamillanz/email-hunt@latest` would fail — Go resolves the module path to a real Git repository.

#### `main.go` at root, not inside `cmd/`

The `cmd/` pattern (`cmd/email-hunt/main.go`) is for repos that ship **multiple binaries** (e.g., a server + a CLI + a worker). We have exactly one binary, so a root `main.go` is idiomatic and simpler. Less nesting = less cognitive load. If we ever add a second binary (like a daemon), we'd refactor to `cmd/` at that point.

#### Everything else inside `internal/`

`internal/` is a Go compiler-enforced boundary. Any package inside `internal/` can only be imported by code within the same module. This means:
- No external project can accidentally depend on our internal logic.
- We can refactor the internals freely without breaking anyone.
- It communicates intent: "this is implementation detail, not a public library."

We specifically do **not** use `pkg/`. The Go team [explicitly advises against `pkg/`](https://go.dev/doc/modules/layout) unless you are building a library others import. For a CLI tool, `pkg/` is a smell — it signals "this might be a library" when it isn't.

#### `doc.go` files in each package

A `doc.go` file is a Go convention for package-level documentation. It gives each package a place to exist before we write real code, and `go build` won't complain about empty packages. It also lets us run `go doc ./internal/generator` immediately.

```go
// Package generator creates email address permutations from a person's name.
package generator
```

#### `CGO_ENABLED=0` from day one

We'll verify it compiles with `CGO_ENABLED=0 go build`. CGo links against C libraries, which makes binaries dynamically linked and platform-dependent. Disabling it produces pure Go static binaries — the holy grail of distribution: a single file that runs on any Linux/macOS/Windows machine of the same architecture, no dependencies. This is a constraint from the PRD.

---

## Phase 1: CLI Argument Parsing & Help

### What we build

The `main.go` that reads two positional arguments (name, domain) plus optional flags (`--ai`, `--concurrency`, `--delay`, `--help`, `--version`), validates them, and prints help.

### Why each decision

#### We use `flag` from stdlib, not Cobra

For a CLI with zero subcommands and ~5 flags, `flag` is the right choice:
- Zero external dependencies. Binaries stay small (critical for `go install` UX).
- `flag` is battle-tested; it ships with Go itself.
- We have exactly two positional args (name and domain). `flag` handles this with `flag.Args()` after parsing flags.

Cobra would add ~17 transitive dependencies for something `flag` handles in 30 lines. The PRD says: "if subcommands are added, migrate to Cobra." That's the right trigger — complexity, not anticipation.

#### Positional args via `flag.Args()`, not `flag.String("name", ...)`

Go's `flag` package doesn't support required positional arguments natively (it's flag-oriented, not arg-oriented). We parse flags first, then grab the remaining arguments with `flag.Args()`. If `len(flag.Args()) < 2`, we print usage and exit with code 2 (the Unix convention for "you used it wrong"). This is idiomatic Go CLI behavior.

#### `--version` via a build-time variable, not hardcoded

```go
var version = "dev"
```

In `main.go` we declare a package-level `var version` with a default of `"dev"`. At build time, `go build -ldflags "-X main.version=v1.0.0"` overwrites it. This means:
- Development builds automatically show `dev`.
- Tagged releases embed the real version.
- No manual version bumps in source files.

#### `--help` is free with `flag`

`flag` auto-generates `-h` and `--help` output from flag definitions. We customize `flag.Usage` to add the disclaimer text required by the PRD and to show the positional argument syntax (which `flag` doesn't do by default).

#### Error output goes to stderr, data goes to stdout

This is Unix convention: you can pipe `email-hunt "John Doe" example.com | grep valid` and the table goes through the pipe, while errors and progress don't contaminate the output. We enforce this from phase 1 so every subsequent phase respects it.

---

## Phase 2: Name Parser & Email Pattern Generator

### What we build

`internal/generator/` — takes a full name string, normalizes it, splits it into first/last/middle, and produces 22+ email combinations for a given domain.

### Structure

```
internal/generator/
├── generator.go      # public API: Generate(name, domain string) []string
├── generator_test.go # table-driven tests
├── patterns.go       # pattern templates (first.last, f.last, etc.)
└── name.go           # name parsing & Unicode normalization
```

### Why each decision

#### A dedicated `generator` package, not a function in `main.go`

Go encourages small, focused packages. The generator has a single responsibility: "given a name and domain, return email permutations." By isolating it in `internal/generator`, we can:
- Test it in isolation without any CLI or network concerns.
- Replace the algorithm later without touching anything else.
- Understand it by reading ~100 lines, not 1000.

#### Package API is one exported function: `Generate(name, domain string) []string`

In Go, you export by capitalizing the first letter. `Generate` is exported; `splitName` is not. This is the simplest possible API — callers don't need to know about patterns, separators, or name parsing internals. The `[]string` return is a slice (Go's dynamic array). No pointers, no channels — just data in, data out. This makes it trivial to test and compose.

#### Name parsing: first word = firstName, last word = lastName, middle = everything else

This handles "John Doe", "María José García López", and "John Michael Doe" with the same logic. It's a heuristic, not perfect, but it covers the vast majority of real-world inputs. We document the assumption in the package doc.

#### Unicode normalization with `golang.org/x/text`

Names like "María José García" contain accented characters. Email addresses are ASCII. We use `golang.org/x/text/unicode/norm` to normalize to NFKD (compatibility decomposition) and then strip diacritics — "María" becomes "Maria", "José" becomes "Jose". This is a Go extended library (`x/` = experimental/extra, maintained by the Go team but not in stdlib). It's the standard way to handle Unicode in Go.

#### Patterns as a data structure, not if/else chains

Instead of:

```go
if pattern == "first.last" { ... }
else if pattern == "f.last" { ... }
```

We define patterns as a slice of structs:

```go
type pattern struct {
    name string
    fn   func(first, middle, last string) string
}

var patterns = []pattern{
    {"{first}.{last}", func(f, m, l string) string { return f + "." + l }},
    {"{f}{last}",      func(f, m, l string) string { return string(f[0]) + l }},
    // ... 22+ patterns
}
```

This is data-driven design — adding a new pattern is one line, not a code change. It's also testable: we can iterate patterns and verify each one produces the expected output for a known input. This is a Go idiom: prefer data structures over control flow.

#### Table-driven tests

Go tests use a convention called "table-driven tests":

```go
func TestGenerate(t *testing.T) {
    tests := []struct {
        name     string
        fullName string
        domain   string
        want     []string
    }{
        {name: "simple", fullName: "John Doe", domain: "example.com", want: []string{...}},
        {name: "with middle", fullName: "John Michael Doe", domain: "example.com", want: []string{...}},
        {name: "unicode", fullName: "María José García", domain: "example.com", want: []string{...}},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Generate(tt.fullName, tt.domain)
            // assert
        })
    }
}
```

This is the standard Go testing pattern because:
- Adding a new test case is one struct literal.
- Each case runs as a subtest (`t.Run`), so you can run just one: `go test -run TestGenerate/simple`.
- The test reads like documentation: input → expected output.

---

## Phase 3: Email & Domain Classifiers (Offline)

### What we build

`internal/classifier/` — classifies emails and domains without network access. Detects:
- **Disposable domains** (Mailinator, 10MinuteMail, GuerrillaMail, etc.)
- **Role-based addresses** (admin@, info@, support@, etc.)

### Structure

```
internal/classifier/
├── classifier.go              # public API
├── classifier_test.go
├── disposable_domains.txt     # embedded with go:embed
└── roles.go                   # role-based prefix list
```

### Why each decision

#### We separate classification from verification

Classification is purely local and deterministic. SMTP verification requires network access, has timeouts, and can fail. By keeping them in separate packages, we can:
- Classify everything instantly before any network call.
- Skip SMTP verification for emails we already know are disposable (saving time and avoiding hitting mail servers needlessly).
- Test classification in milliseconds, no network needed.

#### `//go:embed` for the disposable domains list

```go
package classifier

import _ "embed"

//go:embed disposable_domains.txt
var disposableDomainsData []byte
```

`go:embed` is a compiler directive (introduced in Go 1.16) that bundles files into the binary at compile time. The file literally becomes part of the executable. This means:
- No external file dependency at runtime — the binary is self-contained.
- No file path resolution issues ("where is my data file?")
- The list can be a plain text file, one domain per line, easy to maintain and PR.

This is a classic Go pattern for embedding static data — config templates, SQL migrations, HTML assets, etc.

#### Classification returns an enum type, not a string

```go
type Category int

const (
    CategoryUnknown Category = iota
    CategoryDisposable
    CategoryRoleBased
    CategoryStandard
)
```

`iota` is Go's enum generator — each const gets an auto-incrementing integer. Using a typed constant instead of a string (`"disposable"`) means:
- Compile-time safety: you can't typo `"disposble"` without a compiler error.
- Switch statements are exhaustive-checkable.
- String representation is added via `String()` method, not used internally.

#### `net/mail` for basic parsing

We don't need regex to extract the local part of an email. Go's `net/mail.ParseAddress` parses RFC 5322 addresses and gives us the local part and domain. Use the standard library when it exists — less code to write, test, and maintain.

---

## Phase 4: MX Lookup & Catch-All Detection

### What we build

`internal/verifier/dns.go` — DNS MX record lookup using `net.LookupMX` from stdlib. Plus catch-all detection by testing a random email.

### Why each decision

#### We use `net.LookupMX` from stdlib, not `miekg/dns`

`net.LookupMX` is a high-level function that queries the system resolver for MX records, sorts them by priority, and returns a clean `[]*net.MX` slice. `miekg/dns` is a full DNS library — powerful but overkill for "find the mail server." Using stdlib means:
- No additional dependency.
- Uses the OS resolver (handles /etc/resolv.conf, caching, IPv4/IPv6).
- The API is exactly what we need.

The tradeoff: `net.LookupMX` doesn't let you specify a custom DNS server. For v1, that's fine. If we need custom resolvers later, we migrate.

#### MX records sorted by priority

`net.LookupMX` returns records sorted by preference (lower number = higher priority). We try the lowest-priority server first. If it's unreachable, we fall back to the next. This mirrors how real MTAs deliver mail.

#### No MX = no verification, not a crash

If a domain has no MX records, we return a clear error: `"domain example.com has no mail servers"`. We don't fall back to A records (the RFC technically allows it, but in practice email to domains without MX is almost certainly undeliverable). This keeps the logic simple and matches real-world mail delivery behavior.

#### Catch-all detection: MX lookup is the prerequisite

Catch-all detection needs SMTP, so it belongs in Phase 5 (SMTP Engine). But the MX lookup step is the prerequisite — without MX records, there's nothing to connect to. Phase 4 does the DNS work; Phase 5 does the SMTP work.

---

## Phase 5: SMTP Verification Engine

### What we build

`internal/verifier/smtp.go` — connects to the resolved MX server, performs MAIL FROM + RCPT TO, reads the response, and QUITs without sending DATA. **Never sends an actual email.**

### Structure

```
internal/verifier/
├── verifier.go      # public API: Verify(email string) Result
├── verifier_test.go # tests with mock SMTP server
├── dns.go           # MX lookup (from Phase 4)
├── smtp.go          # SMTP handshake
└── result.go        # Result type definition
```

### Why each decision

#### Raw TCP connection + SMTP commands, not `net/smtp`

Go's `net/smtp` package is designed for **sending** email — it does `AUTH`, `DATA`, and `SendMail`. We need the opposite: connect, ask if a mailbox exists, and leave. Using raw `net.Dial` + `textproto.Conn` gives us:
- Precise control over the SMTP conversation (HELO → MAIL FROM → RCPT TO → QUIT).
- The ability to read and interpret specific SMTP response codes.
- No risk of accidentally triggering `DATA` and sending an email.

This is the right call because we're using SMTP for a non-standard purpose (verification, not delivery).

#### The SMTP conversation step by step

```
Client: HELO email-hunt.local               (introduce ourselves)
Server: 250 mail.example.com                (OK)
Client: MAIL FROM: <verify@email-hunt.local> (we are the sender)
Server: 250 OK                              (sender accepted)
Client: RCPT TO: <target@example.com>        (does this mailbox exist?)
Server: 250 OK                              → VALID
Server: 550 No such user                    → INVALID
Server: 550 5.1.1 User unknown              → INVALID
Client: QUIT                                (disconnect, never send DATA)
```

The email is **never transmitted**. We hang up after RCPT TO. This is how every email verification service works.

#### SMTP response codes as constants

```go
const (
    smtpOK            = 250
    smtpUserUnknown   = 550
    smtpMailboxNotFound = 551
    smtpRelayDenied   = 554
)
```

Not magic numbers scattered in the code. Named constants make the code self-documenting and let us add new response codes in one place.

#### `context.Context` for timeouts and cancellation

```go
func Verify(ctx context.Context, email string) (Result, error) {
    dialer := net.Dialer{Timeout: 5 * time.Second}
    conn, err := dialer.DialContext(ctx, "tcp", mxHost+":25")
    // ...
}
```

`context.Context` is Go's standard mechanism for deadlines, cancellation, and request-scoped values. We pass a context with a 5-second timeout. If the mail server hangs, the context cancels and the connection closes. Without context, a hung server would block the goroutine forever. With context, the user gets a timeout error in 5 seconds. This is idiomatic Go for any I/O operation.

#### `defer conn.Close()` is not just cleanup — it's safety

```go
conn, err := dialer.DialContext(ctx, "tcp", addr)
if err != nil {
    return Result{}, err
}
defer conn.Close()
```

`defer` schedules a function to run when the enclosing function returns. Placing `defer conn.Close()` immediately after a successful dial guarantees the connection closes even if we `return` early due to an error, a timeout, or a panic. Without `defer`, every error path must manually close — and someone will forget.

#### Result as a struct, not multiple return values

```go
type Result struct {
    Email    string
    Status   Status    // valid, invalid, catch_all
    Category Category  // standard, disposable, role_based
    Detail   string    // human-readable explanation
}
```

Go functions can return multiple values (`(Result, error)`), but the result itself is a struct, not a tuple. A struct gives names to each field, documents what each piece means, and lets us add fields later without changing every call site.

---

## Phase 6: Concurrency & Rate Limiting

### What we build

A worker pool pattern that verifies multiple emails in parallel with configurable concurrency and delays.

### Why each decision

#### Worker pool pattern: goroutines + buffered channel, not a library

```go
func VerifyAll(emails []string, workers int, delay time.Duration) []Result {
    jobs := make(chan string, len(emails))
    results := make(chan Result, len(emails))

    var wg sync.WaitGroup
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for email := range jobs {
                results <- Verify(ctx, email)
                time.Sleep(delay)
            }
        }()
    }

    for _, email := range emails {
        jobs <- email
    }
    close(jobs)

    wg.Wait()
    close(results)
    // collect results
}
```

This is the standard Go concurrency pattern. We explicitly choose not to use a third-party worker pool library because:
- The pattern is ~20 lines of code — a dependency would be larger than the implementation.
- It teaches goroutines, channels, and `sync.WaitGroup` — foundational Go concepts.
- We control exactly how delays and concurrency interact.

#### Goroutines are not threads

A goroutine is a lightweight, user-space thread managed by the Go runtime. Starting 20 goroutines costs microseconds and kilobytes — unlike OS threads, which cost megabytes each. This is why we can spawn a goroutine per email without thinking about resource limits. The runtime multiplexes goroutines onto OS threads automatically.

#### Buffered channels prevent blocking

`make(chan string, len(emails))` creates a channel with enough capacity for all emails. Without the buffer, sending to the channel would block until a worker reads — which could deadlock if workers haven't started yet. Buffered channels decouple producers (main goroutine) from consumers (workers).

#### `sync.WaitGroup` to synchronize completion

```go
var wg sync.WaitGroup
wg.Add(workers)      // "I'm waiting for N things"
go func() {
    defer wg.Done()  // "this thing is done"
}()
wg.Wait()            // "block until all N things are done"
```

Without `WaitGroup`, we'd have to count completed jobs manually or use another channel. `WaitGroup` is the canonical Go way to wait for a group of goroutines to finish.

#### Delay between batches, not between every email

The PRD says `--delay` controls time between requests. We interpret this as a delay after each email verification within a worker. With 5 workers and 500ms delay, each worker waits 500ms after completing one email before taking the next. This means approximately 10 verifications per second — fast enough for UX, slow enough to not trigger rate limits.

---

## Phase 7: Output Formatting

### What we build

`internal/output/` — takes a `[]Result` and renders a colored terminal table.

### Structure

```
internal/output/
├── table.go      # table rendering
└── colors.go     # color helpers
```

### Why each decision

#### `tablewriter` for the table

[`olekukoneko/tablewriter`](https://github.com/olekukoneko/tablewriter) is the standard Go library for ASCII tables. It handles column alignment, padding, borders, and auto-sizing. We could write our own with `fmt.Printf` and manual padding math, but:
- Edge cases (Unicode width, terminal width detection) are already solved.
- It's one dependency with no transitive dependencies — ~250 lines of Go.
- It matches what users expect from CLI tools (it's used by Docker, Kubernetes tools, etc.).

#### `fatih/color` for colors

[`fatih/color`](https://github.com/fatih/color) is the simplest Go color library:

```go
color.GreenString("valid")    // green text
color.RedString("invalid")    // red text
color.YellowString("catch_all") // yellow text
```

We could use raw ANSI escape codes (`\033[32m`), but `color` handles Windows (where ANSI codes don't work natively), respects `NO_COLOR` env var, and detects non-TTY output (pipes, redirects). These are real-world concerns that raw ANSI doesn't address.

#### Color coding by status

| Status | Color | Meaning |
|--------|-------|---------|
| `valid` | Green | Email exists, go ahead |
| `invalid` | Red | Doesn't exist, skip |
| `catch_all` | Yellow | Server accepts all, uncertain |
| `disposable` | Gray | Temporary/throwaway |
| `role_based` | Cyan | Group address, not personal |
| `unknown` | White | Couldn't determine |

This makes the table scannable at a glance — green rows are actionable, red rows can be ignored.

#### Data to stdout, progress to stderr

```go
fmt.Fprintln(os.Stderr, "Verifying 22 emails with 5 workers...")  // progress
table.Render()  // writes to stdout
```

This lets users pipe the output: `email-hunt "John Doe" example.com | grep valid`. The grep matches only data rows, not progress messages.

---

## Phase 8: AI Provider Integration (Optional)

### What we build

`internal/ai/` — an interface for AI providers, plus OpenAI and Anthropic implementations.

### Structure

```
internal/ai/
├── provider.go          # Provider interface
├── openai.go            # OpenAI implementation
├── anthropic.go         # Anthropic implementation
└── provider_test.go     # mock-based tests
```

### Why each decision

#### An interface, not a concrete type

```go
type Provider interface {
    Generate(ctx context.Context, name, domain string) ([]string, error)
}
```

Go interfaces are **implicitly satisfied** — any type with a `Generate(ctx, name, domain) ([]string, error)` method implements `Provider` automatically. There's no `implements` keyword. This means:
- `OpenAIProvider` and `AnthropicProvider` don't declare "I implement Provider" — they just have the right method, and the compiler verifies it at the call site.
- Adding a new provider (e.g., Ollama for local LLMs) is one new file — no changes to any other file.
- The main code only knows about `Provider`, not OpenAI or Anthropic specifically.

This is Go's approach to polymorphism: small interfaces, satisfied implicitly, composed at the call site.

#### API key from environment variables, not config files or flags

```go
func NewOpenAIProvider() (*OpenAIProvider, error) {
    key := os.Getenv("EMAIL_HUNT_OPENAI_KEY")
    if key == "" {
        return nil, fmt.Errorf("EMAIL_HUNT_OPENAI_KEY not set")
    }
    return &OpenAIProvider{apiKey: key}, nil
}
```

Env vars are the [12-factor app](https://12factor.net/config) standard for secrets. They're never committed to git, never in shell history (if set via a secrets manager), and never visible in `ps aux`. The PRD explicitly forbids flags for API keys (shell history exposure).

#### Provider selection via a simple factory

```go
func NewProvider(name string) (Provider, error) {
    switch name {
    case "openai":
        return NewOpenAIProvider()
    case "anthropic":
        return NewAnthropicProvider()
    default:
        return nil, fmt.Errorf("unknown AI provider: %s", name)
    }
}
```

A switch statement is simpler than a registry map or plugin system for 2-3 providers. If we reach 10 providers, we'd refactor to a map-based registry. But YAGNI (You Ain't Gonna Need It) applies here.

---

## Phase 9: Testing & Coverage

### What we test

| Package | Test type | What |
|---------|-----------|------|
| `generator` | Unit | Every pattern produces expected output |
| `classifier` | Unit | Correctly identifies disposable, role-based, standard |
| `verifier` | Unit with mock | SMTP conversation parsing, MX lookup error handling |
| `ai` | Unit with mock | Provider selection, env var handling |
| Whole binary | Integration | `email-hunt "John Doe" example.com` (requires network) |

### Why each decision

#### Mock SMTP server for verifier tests

We start a real TCP server in the test:

```go
func TestVerify(t *testing.T) {
    listener, _ := net.Listen("tcp", "127.0.0.1:0")
    defer listener.Close()

    go func() {
        conn, _ := listener.Accept()
        defer conn.Close()
        // Simulate SMTP conversation
        fmt.Fprintln(conn, "220 mail.example.com ESMTP")
        // read HELO
        // respond 250
        // read MAIL FROM
        // respond 250
        // read RCPT TO
        fmt.Fprintln(conn, "250 OK")  // pretend mailbox exists
        // read QUIT
        fmt.Fprintln(conn, "221 Bye")
    }()

    result, err := Verify(ctx, "test@"+listener.Addr().String())
    // assert result.Status == Valid
}
```

This tests real network I/O without depending on an external mail server. It's fast, deterministic, and works offline. The pattern of starting a local server in tests is common in Go for testing HTTP clients, gRPC services, and TCP protocols.

#### Table-driven tests everywhere

Every test file uses the table-driven pattern from Phase 2. Consistency matters — when every test file looks the same, adding new tests is muscle memory.

#### Race detector: `go test -race`

Go has a built-in race detector. `-race` instruments the binary to detect concurrent access to shared memory without synchronization. We run it on the verifier tests (which use goroutines). If two goroutines write to the same variable without a mutex or channel, the race detector catches it. Critical for Phase 6's concurrency.

#### Coverage target: 90% on generator, 80% elsewhere

`go test -cover ./...` shows per-package coverage. The generator is pure logic — easy to cover exhaustively. The verifier has network code — harder to cover every error path. 90% on business logic is achievable; 80% on I/O code is pragmatic.

---

## Phase 10: Polish, CI & Distribution

### What we build

- `.goreleaser.yaml` for future binary releases
- GitHub Actions for CI (test on push, test on PR)
- README with usage examples
- Git tag `v1.0.0` for `go install`

### Why each decision

#### Goreleaser config now, use later (v2)

We write `.goreleaser.yaml` in v1 even though binary releases are v2. It costs nothing to write, documents the build targets, and makes v2 faster.

#### GitHub Actions for CI

```yaml
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: go test -race -cover ./...
      - run: CGO_ENABLED=0 go build -o /dev/null .
```

Every push runs tests and verifies the build compiles without CGo. CI is table stakes for any project others might use — it proves the code works on a clean machine.

#### Tag triggers `go install` availability

```bash
git tag v1.0.0
git push origin v1.0.0
```

After this, `go install github.com/yamillanz/email-hunt@v1.0.0` works. The `@v1.0.0` tells Go's module proxy (proxy.golang.org) to serve that specific tagged version. Without tags, `@latest` works but returns a pseudo-version like `v0.0.0-20250706150432-14c0d48ead0c` — tagged releases are cleaner and communicate stability.
