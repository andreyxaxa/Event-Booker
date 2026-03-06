package mailsender

import (
	"fmt"

	"github.com/andreyxaxa/Event-Booker/internal/dto"
	"github.com/andreyxaxa/Event-Booker/pkg/smtpsender"
)

const (
	subject = "Cancellation of booking."
)

type MailSender struct {
	*smtpsender.SmtpSender
}

func New(sender *smtpsender.SmtpSender) *MailSender {
	return &MailSender{sender}
}

func (ms *MailSender) Send(cb dto.CancelledBooking) error {
	to := []string{cb.Email}

	text := fmt.Sprintf("Hello!\nYour booking for event %d was cancelled after the booking period expired.", cb.EventID)

	msg := []byte("To: " + cb.Email + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		text + "\r\n",
	)

	err := ms.SendMail(to, msg)
	if err != nil {
		return fmt.Errorf("MailSender - Send - ms.SendMail: %w", err)
	}

	return nil
}
