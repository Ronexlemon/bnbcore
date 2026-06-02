package senders

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type Sender struct {
	config Config
}

func NewSender(config Config) *Sender {
	return &Sender{config: config}
}


type EmailPayload struct {
	To       string
	Subject  string
	Title    string
	Greeting string
	Body     string 
	Footer   string
}
func (s *Sender) Send(payload EmailPayload) error {
    htmlBody := BuildHTML(payload.Title, payload.Greeting, payload.Body, payload.Footer)

    msg := fmt.Sprintf(
        "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
        s.config.From,
        payload.To,
        payload.Subject,
        htmlBody,
    )

    addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)

    client, err := smtp.Dial(addr)
    if err != nil {
        return fmt.Errorf("failed to dial SMTP server: %w", err)
    }
    defer client.Close()

    tlsConfig := &tls.Config{
        InsecureSkipVerify: true,
        ServerName:         s.config.Host,
    }

    if err = client.StartTLS(tlsConfig); err != nil {
        return fmt.Errorf("failed to establish TLS: %w", err)
    }
    auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

    if err = client.Auth(auth); err != nil {
        return fmt.Errorf("smtp authentication failed: %w", err)
    }

    if err = client.Mail(s.config.From); err != nil {
        return fmt.Errorf("failed to set MAIL FROM: %w", err)
    }
    if err = client.Rcpt(payload.To); err != nil {
        return fmt.Errorf("failed to set RCPT TO: %w", err)
    }

    w, err := client.Data()
    if err != nil {
        return fmt.Errorf("failed to open data writer: %w", err)
    }

    _, err = w.Write([]byte(msg))
    if err != nil {
        return fmt.Errorf("failed to write message body: %w", err)
    }

    return w.Close()
}

func BuildHTML(title, greeting, body, footer string) string {
	const tmpl = `<!DOCTYPE html>
<html>
<body style="font-family:Arial,sans-serif;max-width:600px;margin:auto;padding:20px;color:#333">
  <h2 style="color:#1a1a2e">{{.Title}}</h2>
  <p>{{.Greeting}}</p>
  <div style="background:#f9f9f9;padding:15px;border-radius:8px;line-height:1.6">{{.Body}}</div>
  <p style="color:#999;font-size:12px;margin-top:20px;border-top:1px solid #eee;padding-top:10px">{{.Footer}}</p>
</body>
</html>`

	t, err := template.New("email").Funcs(template.FuncMap{}).Parse(tmpl)
	if err != nil {
		return fmt.Sprintf("<p>%s</p><p>%s</p><p>%s</p>", greeting, body, footer)
	}

	var buf bytes.Buffer
	_ = t.Execute(&buf, struct {
		Title    string
		Greeting string
		Body     template.HTML
		Footer   string
	}{
		Title:    title,
		Greeting: greeting,
		Body:     template.HTML(body), 
		Footer:   footer,
	})

	return buf.String()
}