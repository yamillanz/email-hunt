package verifier

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/yamillanz/email-hunt/internal/classifier"
)

type Result struct {
	Email    string
	Status   Status
	Category classifier.Category
	Detail   string
}

func Verify(ctx context.Context, email string) Result {
	result := Result{
		Email:    email,
		Status:   StatusUnknown,
		Category: classifier.Classify(email),
		Detail:   "",
	}

	if result.Category == classifier.CategoryDisposable {
		result.Status = StatusUnknown
		result.Detail = "disposable email domain — skipping SMTP verification"
		return result
	}

	domain := extractDomain(email)
	if domain == "" {
		result.Status = StatusInvalid
		result.Detail = "invalid email format"
		return result
	}

	records, err := LookupMX(domain)
	if err != nil {
		result.Status = StatusUnknown
		result.Detail = err.Error()
		return result
	}

	catchAll, err := isCatchAll(ctx, records, domain)
	if err != nil {
		result.Status = StatusUnknown
		result.Detail = fmt.Sprintf("catch-all check failed: %v", err)
		return result
	}
	if catchAll {
		result.Status = StatusCatchAll
		result.Detail = "domain accepts all email (catch-all)"
		return result
	}

	addr := net.JoinHostPort(strings.TrimSuffix(records[0].Host, "."), "25")
	status, detail, err := dialAndVerify(ctx, addr, email)
	if err != nil {
		result.Status = StatusUnknown
		result.Detail = err.Error()
		return result
	}

	result.Status = status
	result.Detail = detail
	return result
}

func isCatchAll(ctx context.Context, records []*net.MX, domain string) (bool, error) {
	randomLocal, err := randomString(20)
	if err != nil {
		return false, err
	}
	randomEmail := fmt.Sprintf("%s@%s", randomLocal, domain)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	addr := net.JoinHostPort(strings.TrimSuffix(records[0].Host, "."), "25")
	status, _, err := dialAndVerify(ctx, addr, randomEmail)
	if err != nil {
		return false, err
	}

	return status == StatusValid, nil
}

func extractDomain(addr string) string {
	_, domain, found := strings.Cut(addr, "@")
	if !found {
		return ""
	}
	return domain
}

func randomString(n int) (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, n)
	for i := range result {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		result[i] = letters[idx.Int64()]
	}
	return string(result), nil
}

func VerifyAll(ctx context.Context, emails []string, workers int, delay time.Duration) []Result {
	jobs := make(chan string, len(emails))
	results := make(chan Result, len(emails))

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for email := range jobs {
				results <- Verify(ctx, email)
				if delay > 0 {
					select {
					case <-ctx.Done():
						return
					case <-time.After(delay):
					}
				}
			}
		}()
	}

	for _, email := range emails {
		jobs <- email
	}
	close(jobs)

	wg.Wait()
	close(results)

	var out []Result
	for r := range results {
		out = append(out, r)
	}
	return out
}
