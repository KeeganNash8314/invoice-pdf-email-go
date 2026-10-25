package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInvoiceHTMLDecision(t *testing.T) {
	h := invoiceHTML(Invoice{Number: "INV-7", Customer: "Ada", Amount: "42 USD"})
	if h != "<h1>Invoice INV-7</h1><p>Customer: Ada</p><p>Total: 42 USD</p>" {
		t.Fatalf("unexpected invoice HTML: %s", h)
	}
}

func TestIssueInvoiceUsesPDFURLAndValidEmailFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/pdf/generate":
			w.Write([]byte(`{"ok":true,"data":{"url":"data:application/pdf;base64,JVBERi0="}}`))
		case "/v1/email/send":
			var fields map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
				t.Errorf("decode email: %v", err)
			}
			if len(fields) != 3 || fields["text"] != nil || !strings.Contains(string(fields["html"]), "data:application/pdf;base64,JVBERi0=") {
				t.Errorf("unexpected email request: %v", fields)
			}
			w.Write([]byte(`{"ok":true,"data":{"message_id":"msg-1"}}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()
	id, err := IssueInvoice(context.Background(), &Client{BaseURL: server.URL, HTTP: server.Client()}, Invoice{Number: "INV-1", Customer: "Ada", Amount: "42 USD", Email: "chenhua@changba.com"})
	if err != nil || id != "msg-1" {
		t.Fatalf("IssueInvoice = %q, %v", id, err)
	}
}
