package verifier

import (
	"context"
	"fmt"
	"net"
	"net/textproto"
)

const (
	smtpOK              = 250
	smtpUserNotFound    = 550
	smtpMailboxNotFound = 551
	smtpRelayDenied     = 554
)

func dialAndVerify(ctx context.Context, addr string, recipient string) (Status, string, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return StatusUnknown, "", fmt.Errorf("connect to %s: %w", addr, err)
	}
	defer conn.Close()

	tc := textproto.NewConn(conn)
	defer tc.Close()

	if _, _, err := tc.ReadResponse(220); err != nil {
		return StatusUnknown, "", fmt.Errorf("read greeting: %w", err)
	}

	if err := tc.PrintfLine("HELO email-hunt.local"); err != nil {
		return StatusUnknown, "", fmt.Errorf("send HELO: %w", err)
	}
	if _, _, err := tc.ReadResponse(250); err != nil {
		return StatusUnknown, "", fmt.Errorf("read HELO response: %w", err)
	}

	if err := tc.PrintfLine("MAIL FROM:<verify@email-hunt.local>"); err != nil {
		return StatusUnknown, "", fmt.Errorf("send MAIL FROM: %w", err)
	}
	if _, _, err := tc.ReadResponse(250); err != nil {
		return StatusUnknown, "", fmt.Errorf("read MAIL FROM response: %w", err)
	}

	if err := tc.PrintfLine("RCPT TO:<%s>", recipient); err != nil {
		return StatusUnknown, "", fmt.Errorf("send RCPT TO: %w", err)
	}
	code, msg, err := tc.ReadResponse(0)
	if err != nil {
		return StatusUnknown, "", fmt.Errorf("read RCPT TO response: %w", err)
	}

	tc.PrintfLine("QUIT")

	switch {
	case code == smtpOK:
		return StatusValid, msg, nil
	case code >= 500 && code < 600:
		return StatusInvalid, msg, nil
	default:
		return StatusUnknown, msg, nil
	}
}
