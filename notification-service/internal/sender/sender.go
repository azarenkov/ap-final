package sender

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

type EmailSender interface {
	Send(to, subject, body string) error
}

type LogSender struct {
	From string
}

func NewLogSender(from string) *LogSender { return &LogSender{From: from} }

func (s *LogSender) Send(to, subject, body string) error {
	log.Printf("EMAIL (log) from=%s to=%s subject=%q body=%q", s.From, to, subject, body)
	return nil
}

type SMTPSender struct {
	host     string
	from     string
	username string
	password string
	useTLS   bool
}

func NewSMTPSender(host, from, username, password string, useTLS bool) *SMTPSender {
	return &SMTPSender{host: host, from: from, username: username, password: password, useTLS: useTLS}
}

func (s *SMTPSender) Send(to, subject, body string) error {
	if to == "" {
		log.Printf("EMAIL (smtp skipped, no recipient) subject=%q", subject)
		return nil
	}
	msg := buildMessage(s.from, to, subject, body)
	hostOnly := s.host
	if i := strings.LastIndex(hostOnly, ":"); i >= 0 {
		hostOnly = hostOnly[:i]
	}

	if s.useTLS {
		return s.sendTLS(hostOnly, to, msg)
	}

	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, hostOnly)
	}
	if err := smtp.SendMail(s.host, auth, s.from, []string{to}, msg); err != nil {
		log.Printf("EMAIL (smtp error) host=%s to=%s err=%v", s.host, to, err)
		return err
	}
	log.Printf("EMAIL (smtp sent) host=%s to=%s subject=%q", s.host, to, subject)
	return nil
}

func (s *SMTPSender) sendTLS(hostOnly, to string, msg []byte) error {
	tlsCfg := &tls.Config{ServerName: hostOnly}
	conn, err := tls.Dial("tcp", s.host, tlsCfg)
	if err != nil {
		log.Printf("EMAIL (tls dial failed) host=%s err=%v", s.host, err)
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, hostOnly)
	if err != nil {
		return err
	}
	defer c.Quit()

	if s.username != "" {
		if err := c.Auth(smtp.PlainAuth("", s.username, s.password, hostOnly)); err != nil {
			log.Printf("EMAIL (auth failed) user=%s err=%v", s.username, err)
			return err
		}
	}
	if err := c.Mail(s.from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	log.Printf("EMAIL (smtp+tls sent) host=%s to=%s", s.host, to)
	return nil
}

func buildMessage(from, to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("From: %s\r\n", from))
	b.WriteString(fmt.Sprintf("To: %s\r\n", to))
	b.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	b.WriteString("\r\n")
	return []byte(b.String())
}
