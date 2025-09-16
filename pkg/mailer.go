package pkg

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"time"

	"github.com/IAmRiteshKoushik/pulse/cmd"
	"github.com/wneessen/go-mail"
	"gopkg.in/gomail.v2"
)

type OTPEmailData struct {
	FirstDigit  string
	SecondDigit string
	ThirdDigit  string
	FourthDigit string
	FifthDigit  string
	SixthDigit  string
}

func LoadAndRenderTemplate(data any) (string, error) {
	templatePath := filepath.Join("pkg", "mail.htm")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	var rendered bytes.Buffer
	err = tmpl.Execute(&rendered, data)
	if err != nil {
		return "", err
	}
	return rendered.String(), nil
}

func NewSendEmail(ctx context.Context, to, otp string) error {
	data := OTPEmailData{
		FirstDigit:  string(otp[0]),
		SecondDigit: string(otp[1]),
		ThirdDigit:  string(otp[2]),
		FourthDigit: string(otp[3]),
		FifthDigit:  string(otp[4]),
		SixthDigit:  string(otp[5]),
	}

	body, err := LoadAndRenderTemplate(data)
	if err != nil {
		return err
	}

	msg := mail.NewMsg()
	if err := msg.From(cmd.AppConfig.SMTP.Username); err != nil {
		return fmt.Errorf("failed to set From address: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("failed to set To address: %w", err)
	}
	msg.Subject("Amrita Summer of Code 2025 Welcomes You!")
	msg.SetBodyString(mail.TypeTextHTML, body)

	// retry logic if failed to send email
	const retries = 3
	for attempt := 1; attempt <= retries; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		errChan := make(chan error, 1)
		go func() {
			errChan <- cmd.Mailer.Client.DialAndSendWithContext(attemptCtx, msg)
		}()

		select {
		case err := <-errChan:
			if err != nil {
				return nil
			}
			cmd.Log.Warn(fmt.Sprintf("Attempt %d failed: %v", attempt, err))
			if attempt == retries {
				return fmt.Errorf("failed to send email after %d attempts: %w", retries, err)
			}
		case <-attemptCtx.Done():
			cmd.Log.Warn(fmt.Sprintf("Attempt %d failed: timeout after %v", attempt, 30*time.Second))
			if attempt == retries {
				return fmt.Errorf("failed to send email after %d attempts: timeout", retries)
			}
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		case <-time.After(2 * time.Second):
			// Reattempt
		}
	}

	cmd.Log.Info("[SUCCESS]: Email sent successfully.")
	return nil
}

func SendMail(to []string, otp string) error {
	subject := "Amrita Summer of Code 2025 Welcomes You!"

	data := OTPEmailData{
		FirstDigit:  string(otp[0]),
		SecondDigit: string(otp[1]),
		ThirdDigit:  string(otp[2]),
		FourthDigit: string(otp[3]),
		FifthDigit:  string(otp[4]),
		SixthDigit:  string(otp[5]),
	}

	body, err := LoadAndRenderTemplate(data)
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", cmd.AppConfig.SMTP.Username)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(
		cmd.AppConfig.SMTP.Host,
		cmd.AppConfig.SMTP.Port,
		cmd.AppConfig.SMTP.Username,
		cmd.AppConfig.SMTP.AppPassword,
	)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	cmd.Log.Info("[SUCCESS]: Email sent successfully.")
	return nil
}
