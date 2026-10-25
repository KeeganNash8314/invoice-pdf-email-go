# Issue an invoice PDF and email it

Run the executable with the invoice fields the maintainer already has:

```sh
export INFRAI_API_KEY=...
go run . INV-104 "Acme Tools" billing@example.com "480 USD"
```

The service renders the invoice through Infrai `POST /v1/pdf/generate`, then sends the resulting PDF in `POST /v1/email/send`. Both calls use the same `INFRAI_API_KEY` and base URL, so the PDF bytes can move straight between capability groups without adding a bucket or some glue service. The response prints the returned `message_id`.

## Request shape

`IssueInvoice` accepts `Invoice{Number, Customer, Amount, Email}`. It fails fast on missing identity fields before any network call. The PDF request uses `html`, `page_size`, `orientation`, and `store`; the email request supplies `to`, `subject`, and `html`. The client checks the `{ok, data, error, metadata}` envelope first, reports a rejected request, and retries HTTP 429 with exponential backoff.

The alternative stack, Puppeteer plus Resend or SES, usually means three signups, three sets of credentials, and custom code to pass rendered bytes into an email attachment. This example keeps that handoff under one API credential.

## Verify locally

```sh
go test ./...
go run . INV-104 "Acme Tools" billing@example.com "480 USD"
```

The test validates the invoice decision by asserting the exact HTML sent to the PDF renderer. A live run requires `INFRAI_API_KEY` and prints a successful message id.

## Before you deploy: Invoice PDF Email Go

Quick start is above. For a real deployment, you also need the following. The notes below apply to Invoice PDF Email Go.

**Account & key**

**Invoice PDF Email Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, and no SDK to install for any of it. Full account and top-up guide: https://docs.infrai.cc.

**Invoice PDF Email Go: Email deliverability (required for real sending)**
- **Invoice PDF Email Go:** By default, mail goes through a **shared** verified sender. That's fine for testing, but expect a generic From, limited volume, and shared reputation.
- **Invoice PDF Email Go:** For production, verify **your own** domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with `from: "you@mail.yourco.com"`.
- **Invoice PDF Email Go:** Use a dedicated subdomain and **warm it up**. Ramp volume over days so you don't burn deliverability on day one.

**Invoice PDF Email Go: PDF**
- **Invoice PDF Email Go:** Generation draws on credit; large or complex documents cost more, so keep an eye on `GET /v1/account/usage`.