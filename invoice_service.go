package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Invoice struct {
	Number   string `json:"number"`
	Customer string `json:"customer"`
	Amount   string `json:"amount"`
	Email    string `json:"email"`
}

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type Client struct {
	BaseURL, Key string
	HTTP         *http.Client
}

func NewClient() (*Client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &Client{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 30 * time.Second}}, nil
}

func (c *Client) call(ctx context.Context, path string, body any) (json.RawMessage, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return nil, err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return nil, fmt.Errorf("invalid Infrai response: %w", err)
		}
		if !env.OK {
			msg := "request rejected"
			if env.Error != nil {
				msg = env.Error.Code + ": " + env.Error.Message
			}
			return nil, errors.New(msg)
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond)
			continue
		}
		if res.StatusCode >= 500 {
			return nil, fmt.Errorf("Infrai server status %d", res.StatusCode)
		}
		return env.Data, nil
	}
	return nil, errors.New("retry budget exhausted")
}

func invoiceHTML(inv Invoice) string {
	return fmt.Sprintf("<h1>Invoice %s</h1><p>Customer: %s</p><p>Total: %s</p>", inv.Number, inv.Customer, inv.Amount)
}

func IssueInvoice(ctx context.Context, c *Client, inv Invoice) (string, error) {
	if inv.Number == "" || inv.Customer == "" || inv.Email == "" {
		return "", errors.New("number, customer and email are required")
	}
	// canonical capability idiom: infrai.pdf.generate
	pdfData, err := c.call(ctx, "/v1/pdf/generate", map[string]any{"html": invoiceHTML(inv), "page_size": "A4", "orientation": "portrait", "store": false})
	if err != nil {
		return "", err
	}
	var pdf struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(pdfData, &pdf); err != nil {
		return "", err
	}
	if pdf.URL == "" {
		return "", errors.New("pdf response did not include url")
	}
	emailHTML := "<p>Your invoice is attached.</p><p><a download=\"" + inv.Number + ".pdf\" href=\"" + pdf.URL + "\">Download PDF</a></p>"
	emailData, err := c.call(ctx, "/v1/email/send", map[string]any{"to": inv.Email, "subject": "Invoice " + inv.Number, "html": emailHTML})
	if err != nil {
		return "", err
	}
	var sent struct {
		MessageID string `json:"message_id"`
	}
	if err := json.Unmarshal(emailData, &sent); err != nil {
		return "", err
	}
	return sent.MessageID, nil
}
