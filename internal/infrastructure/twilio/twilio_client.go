package twilio

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)


type Config struct {
	AccountSID string
	AuthToken  string
	FromNumber string
	TemplateSID string
}


type Client struct {
	cfg        Config
	httpClient *http.Client
}


func NewTwilioClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}


func (c *Client) SendWhatsApp(ctx context.Context, toPhone, message string) error {
	twilioURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", c.cfg.AccountSID)

	
	toPhone = strings.TrimSpace(toPhone)
	if !strings.HasPrefix(toPhone, "whatsapp:") {
		toPhone = "whatsapp:" + toPhone
	}

	fromNumber := strings.TrimSpace(c.cfg.FromNumber)
	if !strings.HasPrefix(fromNumber, "whatsapp:") {
		fromNumber = "whatsapp:" + fromNumber
	}

	data := url.Values{}
	data.Set("From", fromNumber)
	data.Set("To", toPhone)
	data.Set("Body", message)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, twilioURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to build twilio message request: %w", err)
	}

	
	req.SetBasicAuth(c.cfg.AccountSID, c.cfg.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to communicate with twilio edge networks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("twilio validation failed with status code: %d", resp.StatusCode)
	}

	return nil
}


func (w *Client) SendWhatsAppTemplate(_ context.Context, toPhone string, variables map[string]string) error {
    twilioURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", w.cfg.AccountSID)

    varsJSON, err := json.Marshal(variables)
    if err != nil {
        return fmt.Errorf("failed to marshal template variables: %w", err)
    }

    data := url.Values{}
    data.Set("From", w.cfg.FromNumber)
    data.Set("To", "whatsapp:"+toPhone)
    data.Set("ContentSid", w.cfg.TemplateSID)
    data.Set("ContentVariables", string(varsJSON))

    req, err := http.NewRequest(http.MethodPost, twilioURL, strings.NewReader(data.Encode()))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    req.SetBasicAuth(w.cfg.AccountSID, w.cfg.AuthToken)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return fmt.Errorf("failed to execute request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return fmt.Errorf("twilio returned status %d", resp.StatusCode)
    }
    return nil
}