# tokportal-go

[![Go Reference](https://pkg.go.dev/badge/github.com/tokportal/tokportal-go.svg)](https://pkg.go.dev/github.com/tokportal/tokportal-go)
[![license](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

TokPortal is the managed social infrastructure API: real TikTok, Instagram and YouTube accounts created, warmed and operated by human account managers in 16+ countries — exposed as a REST API and an MCP server. No OAuth per account, no 25-posts/day cap, no app review.

Docs https://developers.tokportal.com · API base https://app.tokportal.com/api/ext · OpenAPI https://developers.tokportal.com/openapi.json · MCP remote https://app.tokportal.com/api/ext/mcp · Get an API key https://app.tokportal.com/developer/api-keys · llms.txt https://developers.tokportal.com/llms.txt

---

`github.com/tokportal/tokportal-go` is the official Go SDK for the TokPortal API (Go 1.22+, standard library only). Every public operation is available as a typed resource method.

## Install

```bash
go get github.com/tokportal/tokportal-go
```

## 30-second quickstart

```go
package main

import (
	"context"
	"errors"
	"log"
	"os"

	tokportal "github.com/tokportal/tokportal-go"
)

func main() {
	client, err := tokportal.NewClient(os.Getenv("TOKPORTAL_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	bundle, err := client.Bundles.Create(context.Background(), tokportal.CreateBundleRequest{
		BundleType:     tokportal.BundleTypeAccountAndVideos,
		Platform:       tokportal.PlatformTiktok,
		Country:        "USA",
		Title:          "US launch",
		VideosQuantity: 5,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", bundle)

	csv, err := client.Analytics.ExportVideos(context.Background(), tokportal.Query{
		"account": []string{"saved-account-id"},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s", csv)
}
```

Manage TokPortal Coverage from the latest atomic quote. A zero-credit quote is
valid and still requires an explicit reactivation call:

```go
coverage, err := client.Accounts.Coverage(context.Background(), "saved-account-id")
// Read data.reactivation_quote from coverage, then confirm the exact snapshot.
ctx := tokportal.WithIdempotencyKey(context.Background(), "coverage-reactivate-saved-account-id-v4")
reactivated, err := client.Accounts.ReactivateCoverage(ctx, "saved-account-id", tokportal.ManagedAccountSubscriptionReactivationRequest{
	ExpectedCredits:          50,
	ExpectedCurrentPeriodEnd: "2026-09-09T11:00:00.000Z",
	ExpectedLockVersion:      4,
})
_, _ = coverage, reactivated
```

Credential reveal and verification-code access use the same irreversible
two-step policy flow. First call `RevealCredentials` or `VerificationCode`
without acceptance to receive HTTP 428 and `error.details.policy_version`.
After showing those terms to the account owner, retry with that exact version.
The accepted request may debit credits and permanently detach the account, so
these helpers reject a context created with `WithIdempotencyKey`. Secret-bearing
responses are never stored for replay. After an uncertain transport result,
reconcile the safe account state before deciding whether to call the endpoint
again without a key:

If an accepted call returns HTTP 409 with
`CREDENTIAL_REVEAL_QUOTE_CHANGED`, no charge or reveal occurred. Read the
current policy and `expected_credit_cost` from `APIError.Details`, show the new
terms to the owner, obtain fresh consent, and retry with the new version. Never
retry a 409 automatically.

```go
var apiError tokportal.APIError
_, err = client.Accounts.RevealCredentials(context.Background(), "saved-account-id")
if !errors.As(err, &apiError) || apiError.StatusCode != 428 {
	log.Fatal(err)
}
policyVersion, ok := apiError.Details["policy_version"].(string)
if !ok || policyVersion == "" {
	log.Fatal("TokPortal did not return a policy version")
}
// Show the returned policy terms to the account owner and obtain consent here.
credentials, err := client.Accounts.RevealCredentialsWithAcceptance(context.Background(), "saved-account-id", tokportal.CredentialRevealAcceptance{
	AcknowledgeSupportForfeit: true,
	PolicyVersion:            policyVersion,
})
_ = credentials
```

The same no-replay rule applies to `Webhooks.Create`, `Uploads.Image`,
`Uploads.Video`, and `Analytics.CreateReport` because they return a signing
secret, signed upload capability, or report access token. These helpers and
`DoOperation` reject a context carrying `WithIdempotencyKey` for all six
sensitive operation IDs.

Discover and operate webhooks without dropping to raw HTTP:

```go
catalog, err := client.Webhooks.Events(context.Background())
endpoints, err := client.Webhooks.List(context.Background(), tokportal.Query{
	"event": "bundle.published",
})
retry, err := client.Webhooks.RetryDelivery(
	context.Background(),
	endpoints["data"].([]any)[0].(map[string]any)["id"].(string),
	"delivery-id",
)
_, _, _ = catalog, endpoints, retry
```

Every OpenAPI operation is also reachable through the generated operation map:

```go
sameRetry, err := client.DoOperation(context.Background(), "retryWebhookDelivery", tokportal.OperationRequest{
	Path: map[string]string{
		"id":          endpoints["data"].([]any)[0].(map[string]any)["id"].(string),
		"delivery_id": "delivery-id",
	},
})
_ = sameRetry

csvAgain, err := client.DoTextOperation(context.Background(), "exportAnalyticsVideos", tokportal.OperationRequest{
	Query: tokportal.Query{"account": []string{"saved-account-id"}},
})
_ = csvAgain
```

Use `WithIdempotencyKey` for safe retries on mutating requests:

```go
ctx := tokportal.WithIdempotencyKey(context.Background(), "bundle-create-123")
bundle, err := client.Bundles.Create(ctx, tokportal.CreateBundleRequest{
	BundleType:     tokportal.BundleTypeAccountAndVideos,
	Country:        "USA",
	VideosQuantity: 5,
})
```

```go
bundle, err := client.Bundles.Create(context.Background(), tokportal.CreateBundleRequest{
	BundleType:     tokportal.BundleTypeAccountAndVideos,
	Country:        "USA",
	VideosQuantity: 5,
})
if err != nil {
	var apiErr tokportal.APIError
	if errors.As(err, &apiErr) {
		log.Printf("status=%d code=%s request_id=%s", apiErr.StatusCode, apiErr.Code, apiErr.RequestID)
		if apiErr.Retryable() {
			waitSeconds := apiErr.RetryAfterSeconds
			if waitSeconds == 0 {
				waitSeconds = 1
			}
			// Retry with backoff.
		}
		if apiErr.RateLimit != nil {
			log.Printf("rate_limit_remaining=%d reset=%d", apiErr.RateLimit.Remaining, apiErr.RateLimit.Reset)
		}
	}
	log.Fatal(err)
}
_ = bundle
```

The SDK is generated from the TokPortal public OpenAPI/schema layer and uses `X-API-Key` authentication.

It sends `X-TokPortal-Client: tokportal-go/0.1.0` on API requests for observability and support diagnostics.

Verify signed webhook deliveries with the exact raw request body:

```go
valid := tokportal.VerifyWebhookSignature(
	rawBody,
	r.Header.Get("TokPortal-Signature"),
	os.Getenv("TOKPORTAL_WEBHOOK_SECRET"),
	5*time.Minute,
)
```

## Source of truth

This package is generated from the TokPortal public OpenAPI schema
(https://developers.tokportal.com/openapi.json) in the private TokPortal
monorepo. Generated files (`generated.go`) are overwritten on every release — do not edit
them by hand. See [CONTRIBUTING.md](./CONTRIBUTING.md) for what we accept as PRs
and [SECURITY.md](./SECURITY.md) for vulnerability reporting.

## Links

- Documentation: https://developers.tokportal.com
- SDKs & CLI guide: https://developers.tokportal.com/sdks-cli
- MCP server: https://developers.tokportal.com/mcp · [`tokportal-mcp`](https://www.npmjs.com/package/tokportal-mcp)
- API reference (OpenAPI): https://developers.tokportal.com/openapi.json
- Other packages: [`@tokportal/node`](https://www.npmjs.com/package/@tokportal/node) · [`@tokportal/cli`](https://www.npmjs.com/package/@tokportal/cli) · [`tokportal` (PyPI)](https://pypi.org/project/tokportal/) · [`github.com/tokportal/tokportal-go`](https://github.com/tokportal/tokportal-go)

MIT © TokPortal
