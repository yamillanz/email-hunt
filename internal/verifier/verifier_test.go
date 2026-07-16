package verifier

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

type mockSMTPServer struct {
	listener net.Listener
	addr     string
	handler  func(conn net.Conn)
	done     chan struct{}
}

func newMockSMTPServer(t *testing.T, handler func(conn net.Conn)) *mockSMTPServer {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := &mockSMTPServer{
		listener: listener,
		handler:  handler,
		done:     make(chan struct{}),
	}
	s.addr = listener.Addr().String()

	go func() {
		defer close(s.done)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		s.handler(conn)
	}()

	return s
}

func (s *mockSMTPServer) close() {
	s.listener.Close()
	<-s.done
}

func mockValidHandler() func(conn net.Conn) {
	return func(conn net.Conn) {
		defer conn.Close()
		r := bufio.NewReader(conn)

		fmt.Fprintln(conn, "220 mail.example.com ESMTP")
		readLine(r)
		fmt.Fprintln(conn, "250 Hello")
		readLine(r)
		fmt.Fprintln(conn, "250 Sender OK")
		readLine(r)
		fmt.Fprintln(conn, "250 Recipient OK")
		readLine(r)
		fmt.Fprintln(conn, "221 Bye")
	}
}

func mockInvalidHandler() func(conn net.Conn) {
	return func(conn net.Conn) {
		defer conn.Close()
		r := bufio.NewReader(conn)

		fmt.Fprintln(conn, "220 mail.example.com ESMTP")
		readLine(r)
		fmt.Fprintln(conn, "250 Hello")
		readLine(r)
		fmt.Fprintln(conn, "250 Sender OK")
		readLine(r)
		fmt.Fprintln(conn, "550 5.1.1 User unknown")
		readLine(r)
		fmt.Fprintln(conn, "221 Bye")
	}
}

func mockGreetingFailHandler() func(conn net.Conn) {
	return func(conn net.Conn) {
		defer conn.Close()
		fmt.Fprintln(conn, "421 Service not available")
	}
}

func mockCatchAllHandler() func(conn net.Conn) {
	return func(conn net.Conn) {
		defer conn.Close()
		r := bufio.NewReader(conn)

		fmt.Fprintln(conn, "220 mail.example.com ESMTP")
		readLine(r)
		fmt.Fprintln(conn, "250 Hello")
		readLine(r)
		fmt.Fprintln(conn, "250 Sender OK")
		readLine(r)
		fmt.Fprintln(conn, "250 Recipient OK")
		readLine(r)
		fmt.Fprintln(conn, "221 Bye")
	}
}

func readLine(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

func TestDialAndVerifyValid(t *testing.T) {
	srv := newMockSMTPServer(t, mockValidHandler())
	defer srv.close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, detail, err := dialAndVerify(ctx, srv.addr, "test@example.com")

	if err != nil {
		t.Fatalf("dialAndVerify failed: %v", err)
	}
	if status != StatusValid {
		t.Errorf("expected StatusValid, got %v (detail: %s)", status, detail)
	}
}

func TestDialAndVerifyInvalid(t *testing.T) {
	srv := newMockSMTPServer(t, mockInvalidHandler())
	defer srv.close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, _, err := dialAndVerify(ctx, srv.addr, "nobody@example.com")

	if err != nil {
		t.Fatalf("dialAndVerify failed: %v", err)
	}
	if status != StatusInvalid {
		t.Errorf("expected StatusInvalid, got %v", status)
	}
}

func TestDialAndVerifyGreetingFail(t *testing.T) {
	srv := newMockSMTPServer(t, mockGreetingFailHandler())
	defer srv.close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := dialAndVerify(ctx, srv.addr, "test@example.com")

	if err == nil {
		t.Error("expected error on greeting failure")
	}
}

func TestDialAndVerifyCatchAll(t *testing.T) {
	srv := newMockSMTPServer(t, mockCatchAllHandler())
	defer srv.close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, _, err := dialAndVerify(ctx, srv.addr, "random-xyz-123@example.com")
	if err != nil {
		t.Fatalf("dialAndVerify failed: %v", err)
	}
	if status != StatusValid {
		t.Errorf("expected StatusValid (server accepts all recipients), got %v", status)
	}
}

func TestDialAndVerifyConnectionTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond)

	_, _, err := dialAndVerify(ctx, "255.255.255.255:25", "test@example.com")
	if err == nil {
		t.Error("expected error on connection timeout")
	}
}

func TestVerifyDisposableSkipsSMTP(t *testing.T) {
	ctx := context.Background()
	result := Verify(ctx, "test@mailinator.com")

	if result.Status != StatusUnknown {
		t.Errorf("expected StatusUnknown for disposable, got %v", result.Status)
	}
	if !strings.Contains(result.Detail, "disposable") {
		t.Errorf("expected disposable detail, got: %s", result.Detail)
	}
}

func TestVerifyInvalidFormat(t *testing.T) {
	ctx := context.Background()
	result := Verify(ctx, "not-an-email")

	if result.Status != StatusInvalid {
		t.Errorf("expected StatusInvalid for bad format, got %v", result.Status)
	}
}

func TestResultStruct(t *testing.T) {
	r := Result{
		Email:    "john@example.com",
		Status:   StatusValid,
		Detail:   "OK",
	}
	if r.Email != "john@example.com" {
		t.Errorf("email mismatch")
	}
	if r.Status != StatusValid {
		t.Errorf("status mismatch")
	}
}

func TestSMTPResponseCodes(t *testing.T) {
	if smtpOK != 250 {
		t.Errorf("smtpOK = %d, want 250", smtpOK)
	}
	if smtpUserNotFound != 550 {
		t.Errorf("smtpUserNotFound = %d, want 550", smtpUserNotFound)
	}
}

func TestVerifyAllCount(t *testing.T) {
	emails := []string{
		"test@mailinator.com",
		"not-an-email",
		"user@this-domain-does-not-exist-xyz123.com",
	}
	ctx := context.Background()
	results := VerifyAll(ctx, emails, 3, 0)
	if len(results) != len(emails) {
		t.Errorf("expected %d results, got %d", len(emails), len(results))
	}
	seen := make(map[string]bool)
	for _, r := range results {
		seen[r.Email] = true
	}
	for _, email := range emails {
		if !seen[email] {
			t.Errorf("missing result for %s", email)
		}
	}
}

func TestVerifyAllWithDelay(t *testing.T) {
	emails := []string{"test@mailinator.com", "not-an-email"}
	ctx := context.Background()
	start := time.Now()
	results := VerifyAll(ctx, emails, 2, 10*time.Millisecond)
	elapsed := time.Since(start)
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
	if elapsed < 10*time.Millisecond {
		t.Errorf("delay not applied: elapsed=%v", elapsed)
	}
}

func TestVerifyAllWorkersOne(t *testing.T) {
	emails := []string{"a@mailinator.com", "b@mailinator.com", "c@mailinator.com"}
	ctx := context.Background()
	results := VerifyAll(ctx, emails, 1, 0)
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestVerifyAllEmpty(t *testing.T) {
	ctx := context.Background()
	results := VerifyAll(ctx, nil, 5, 0)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestVerifyAllContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	emails := []string{"test@mailinator.com"}
	results := VerifyAll(ctx, emails, 1, time.Second)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestRandomString(t *testing.T) {
	s1, err := randomString(10)
	if err != nil {
		t.Fatalf("randomString failed: %v", err)
	}
	if len(s1) != 10 {
		t.Errorf("expected length 10, got %d", len(s1))
	}

	s2, err := randomString(10)
	if err != nil {
		t.Fatalf("second randomString failed: %v", err)
	}
	if s1 == s2 {
		t.Error("two random strings should not be equal")
	}
}

func TestRandomStringZero(t *testing.T) {
	s, err := randomString(0)
	if err != nil {
		t.Fatalf("randomString(0) failed: %v", err)
	}
	if s != "" {
		t.Errorf("expected empty string, got %q", s)
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		email string
		want  string
	}{
		{"john@example.com", "example.com"},
		{"user@subdomain.domain.co.uk", "subdomain.domain.co.uk"},
		{"not-an-email", ""},
		{"@nouser.com", "nouser.com"},
		{"nodomain@", ""},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got := extractDomain(tt.email)
			if got != tt.want {
				t.Errorf("extractDomain(%q) = %q, want %q", tt.email, got, tt.want)
			}
		})
	}
}

func TestStatusValues(t *testing.T) {
	if int(StatusUnknown) != 0 {
		t.Errorf("StatusUnknown should be 0 (iota base)")
	}
	if int(StatusValid) != 1 {
		t.Errorf("StatusValid should be 1")
	}
	if int(StatusInvalid) != 2 {
		t.Errorf("StatusInvalid should be 2")
	}
	if int(StatusCatchAll) != 3 {
		t.Errorf("StatusCatchAll should be 3")
	}
}

func TestVerifyWithMockServer(t *testing.T) {
	srv := newMockSMTPServer(t, mockValidHandler())
	defer srv.close()

	ctx := context.Background()

	status, _, err := dialAndVerify(ctx, srv.addr, "real.user@example.com")
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if status != StatusValid {
		t.Errorf("expected StatusValid, got %v", status)
	}
}
