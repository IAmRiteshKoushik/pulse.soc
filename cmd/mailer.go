package cmd

import (
	"fmt"
	"sync"
	"time"

	"github.com/wneessen/go-mail"
)

var Mailer *MailClient

type MailClient struct {
	Client *mail.Client
	once   sync.Once
}

func NewMailClient() error {
	if Mailer == nil {
		Mailer = &MailClient{}
	}

	var initErr error
	Mailer.once.Do(func() {
		client, err := mail.NewClient(
			AppConfig.SMTP.Host,
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(AppConfig.SMTP.Username),
			mail.WithPassword(AppConfig.SMTP.AppPassword),
			mail.WithTimeout(20*time.Second),
			mail.WithTLSPortPolicy(mail.TLSOpportunistic),
			mail.WithPort(AppConfig.SMTP.Port),
		)
		if err != nil {
			initErr = fmt.Errorf("failed to create mail client: %w", err)
			return
		}
		Mailer.Client = client
	})
	if initErr != nil {
		return initErr
	}
	return nil
}
