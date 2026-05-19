package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/krtw00/konbu/internal/config"
)

// Mailer sends a plain-text email via SMTP relay.
type Mailer interface {
	Send(to, subject, body string) error
}

// SMTPMailer talks to an SMTP relay over STARTTLS using the stdlib net/smtp
// PLAIN auth. It's intentionally minimal; we only need outbound notifications.
type SMTPMailer struct {
	host string
	port string
	user string
	pass string
	from string
}

// NewSMTPMailer returns a mailer only if all required SMTP env are set.
// Otherwise it returns nil, signalling the notification feature should be
// disabled (no-op).
func NewSMTPMailer(cfg *config.Config) *SMTPMailer {
	if cfg.SMTPHost == "" || cfg.SMTPPort == "" || cfg.SMTPUsername == "" || cfg.SMTPPassword == "" || cfg.SMTPFrom == "" {
		return nil
	}
	return &SMTPMailer{
		host: cfg.SMTPHost,
		port: cfg.SMTPPort,
		user: cfg.SMTPUsername,
		pass: cfg.SMTPPassword,
		from: cfg.SMTPFrom,
	}
}

// Send delivers a single-recipient text/plain message. Subject is encoded as
// a RFC 2047 MIME word when it contains non-ASCII so it survives most relays.
func (m *SMTPMailer) Send(to, subject, body string) error {
	if to == "" {
		return errors.New("notification: empty recipient")
	}
	addr := net.JoinHostPort(m.host, m.port)
	auth := smtp.PlainAuth("", m.user, m.pass, m.host)

	msg := buildMessage(m.from, to, subject, body)
	return smtp.SendMail(addr, auth, m.from, []string{to}, msg)
}

func buildMessage(from, to, subject, body string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", mimeEncodeHeader(subject))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}

// mimeEncodeHeader returns subject as a RFC 2047 "B" encoded word when it
// contains non-ASCII. Pure ASCII is returned unchanged.
func mimeEncodeHeader(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(s)) + "?="
		}
	}
	return s
}
