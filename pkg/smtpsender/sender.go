package smtpsender

import (
	"fmt"
	"net"
	"net/smtp"
)

const (
	_defaultSMTPHost = "smtp.gmail.com"
	_defaultSMTPPort = "587"
)

type SmtpSender struct {
	username string
	password string
	host     string
	port     string
	Addr     string
	auth     smtp.Auth
}

func New(opts ...Option) *SmtpSender {
	s := &SmtpSender{
		host: _defaultSMTPHost,
		port: _defaultSMTPPort,
	}

	for _, opt := range opts {
		opt(s)
	}

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	s.auth = auth

	s.Addr = net.JoinHostPort(s.host, s.port)

	return s
}

func (ss *SmtpSender) SendMail(to []string, msg []byte) error {
	err := smtp.SendMail(ss.Addr, ss.auth, ss.username, to, msg)
	if err != nil {
		return fmt.Errorf("SMTPSender - SendMail - smtp.SendMail: %w", err)
	}

	return nil
}
