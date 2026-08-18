package tokportal

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func testClient(t *testing.T, handler http.HandlerFunc) (*Client, func()) {
	t.Helper()

	server := httptest.NewServer(handler)
	client, err := NewClient("sk_test", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	return client, server.Close
}

func TestBundleCreate(t *testing.T) {
	client, closeServer := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/bundles" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("X-API-Key"); got != "sk_test" {
			t.Fatalf("api key = %s", got)
		}
		if got := r.Header.Get("X-TokPortal-Client"); got != ClientHeader {
			t.Fatalf("client header = %s", got)
		}

		var body CreateBundleRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.BundleType != BundleTypeAccountAndVideos || body.Platform != PlatformTiktok || body.Country != "USA" {
			t.Fatalf("unexpected body: %+v", body)
		}

		_ = json.NewEncoder(w).Encode(Response{"data": Response{"id": "b1"}})
	})
	defer closeServer()

	resp, err := client.Bundles.Create(context.Background(), CreateBundleRequest{
		BundleType:     BundleTypeAccountAndVideos,
		Platform:       PlatformTiktok,
		Country:        "USA",
		VideosQuantity: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp["data"] == nil {
		t.Fatalf("missing data: %+v", resp)
	}
}

func TestBatchConfigureVideos(t *testing.T) {
	client, closeServer := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/bundles/b1/videos/batch" {
			t.Fatalf("path = %s", r.URL.Path)
		}

		var body BatchConfigureVideosRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if len(body.Videos) != 1 || body.Videos[0].Position != 1 || body.Videos[0].VideoType != "video" {
			t.Fatalf("unexpected body: %+v", body)
		}

		_ = json.NewEncoder(w).Encode(Response{"data": "ok"})
	})
	defer closeServer()

	_, err := client.Bundles.BatchConfigureVideos(context.Background(), "b1", BatchConfigureVideosRequest{
		Videos: []BatchConfigureVideoItem{{
			Position: 1,
			ConfigureVideoRequest: ConfigureVideoRequest{
				VideoType:         "video",
				Description:       "caption",
				TargetPublishDate: "2026-06-01",
				VideoUrl:          "https://example.com/video.mp4",
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestStructuredAPIError(t *testing.T) {
	client, closeServer := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-TokPortal-Request-ID", "req_test")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    "AUTH_INVALID_KEY",
				"message": "Invalid API key.",
				"details": map[string]any{
					"hint": "Generate a key from the Developer Portal.",
				},
			},
		})
	})
	defer closeServer()

	_, err := client.Me(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	var apiErr APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized || apiErr.Code != "AUTH_INVALID_KEY" {
		t.Fatalf("unexpected api error: %+v", apiErr)
	}
	if apiErr.RequestID != "req_test" || apiErr.RawBody == "" {
		t.Fatalf("missing api error metadata: %+v", apiErr)
	}
	if apiErr.Retryable() {
		t.Fatalf("401 should not be retryable")
	}
}

func TestRetryableAPIError(t *testing.T) {
	client, closeServer := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "2")
		w.Header().Set("X-RateLimit-Limit", "120")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "1779724800")
		w.WriteHeader(http.StatusTooManyRequests)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    "RATE_LIMITED",
				"message": "Too many requests.",
			},
		})
	})
	defer closeServer()

	_, err := client.Me(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsRetryable(err) {
		t.Fatalf("expected retryable APIError, got %T", err)
	}
	var apiErr APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.RetryAfterSeconds != 2 || apiErr.RateLimit == nil || apiErr.RateLimit.Remaining != 0 || apiErr.RateLimit.Limit != 120 {
		t.Fatalf("missing rate limit metadata: %+v", apiErr)
	}
}

func TestIdempotencyKeyFromContext(t *testing.T) {
	client, closeServer := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Idempotency-Key"); got != "bundle-create-123" {
			t.Fatalf("idempotency key = %s", got)
		}
		_ = json.NewEncoder(w).Encode(Response{"data": Response{"id": "b1"}})
	})
	defer closeServer()

	ctx := WithIdempotencyKey(context.Background(), "bundle-create-123")
	_, err := client.Bundles.Create(ctx, CreateBundleRequest{
		BundleType:     BundleTypeAccountAndVideos,
		Country:        "USA",
		VideosQuantity: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAccountsCoverageAndCredentialReveal(t *testing.T) {
	client, closeServer := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-TokPortal-Client"); got != ClientHeader {
			t.Fatalf("client header = %s", got)
		}

		switch r.URL.Path {
		case "/accounts/account-1/managed-subscription":
			if r.Method != http.MethodGet {
				t.Fatalf("coverage method = %s", r.Method)
			}
		case "/accounts/account-1/managed-subscription/cancel":
			if r.Method != http.MethodPost || r.Header.Get("Idempotency-Key") != "coverage-pause-1" {
				t.Fatalf("pause request method=%s key=%s", r.Method, r.Header.Get("Idempotency-Key"))
			}
		case "/accounts/account-1/managed-subscription/reactivate":
			var body ManagedAccountSubscriptionReactivationRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.ExpectedCredits != 50 || body.ExpectedLockVersion != 4 || body.ExpectedCurrentPeriodEnd != "2026-08-10T11:00:00.000Z" {
				t.Fatalf("reactivation body = %+v", body)
			}
			if r.Header.Get("Idempotency-Key") != "coverage-reactivate-1" {
				t.Fatalf("reactivation key = %s", r.Header.Get("Idempotency-Key"))
			}
		case "/accounts/account-1/reveal-credentials", "/accounts/account-1/verification-code":
			payload, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			if r.Header.Get("Idempotency-Key") != "" {
				t.Fatalf("secret request sent idempotency key = %s", r.Header.Get("Idempotency-Key"))
			}
			if len(payload) == 0 {
				break
			}
			var acceptance CredentialRevealAcceptance
			if err := json.Unmarshal(payload, &acceptance); err != nil {
				t.Fatal(err)
			}
			if !acceptance.AcknowledgeSupportForfeit || acceptance.PolicyVersion != "managed-credential-reveal-v1" {
				t.Fatalf("acceptance = %+v", acceptance)
			}
		default:
			t.Fatalf("unexpected path = %s", r.URL.Path)
		}

		_ = json.NewEncoder(w).Encode(Response{"data": Response{"ok": true}})
	})
	defer closeServer()

	ctx := context.Background()
	if _, err := client.Accounts.Coverage(ctx, "account-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Accounts.PauseCoverage(WithIdempotencyKey(ctx, "coverage-pause-1"), "account-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Accounts.ReactivateCoverage(
		WithIdempotencyKey(ctx, "coverage-reactivate-1"),
		"account-1",
		ManagedAccountSubscriptionReactivationRequest{
			ExpectedCredits:          50,
			ExpectedCurrentPeriodEnd: "2026-08-10T11:00:00.000Z",
			ExpectedLockVersion:      4,
		},
	); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Accounts.RevealCredentials(ctx, "account-1"); err != nil {
		t.Fatal(err)
	}
	acceptance := CredentialRevealAcceptance{
		AcknowledgeSupportForfeit: true,
		PolicyVersion:             "managed-credential-reveal-v1",
	}
	if _, err := client.Accounts.RevealCredentialsWithAcceptance(
		ctx,
		"account-1",
		acceptance,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Accounts.VerificationCodeWithAcceptance(
		ctx,
		"account-1",
		acceptance,
	); err != nil {
		t.Fatal(err)
	}
}

func TestSecretEndpointsRejectIdempotencyKeyLocally(t *testing.T) {
	client, closeServer := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("secret request with an idempotency key must not reach the API")
	})
	defer closeServer()

	ctx := WithIdempotencyKey(context.Background(), "must-not-be-sent")
	acceptance := CredentialRevealAcceptance{
		AcknowledgeSupportForfeit: true,
		PolicyVersion:             "managed-credential-reveal-v1",
	}
	calls := []func() (Response, error){
		func() (Response, error) { return client.Accounts.RevealCredentials(ctx, "account-1") },
		func() (Response, error) {
			return client.Accounts.RevealCredentialsWithAcceptance(ctx, "account-1", acceptance)
		},
		func() (Response, error) { return client.Accounts.VerificationCode(ctx, "account-1") },
		func() (Response, error) {
			return client.Accounts.VerificationCodeWithAcceptance(ctx, "account-1", acceptance)
		},
		func() (Response, error) {
			return client.Analytics.CreateReport(ctx, CreateAnalyticsReportRequest{})
		},
		func() (Response, error) {
			return client.Uploads.Video(ctx, UploadVideoRequest{
				Filename: "video.mp4", BundleId: "bundle-1",
			})
		},
		func() (Response, error) {
			return client.Uploads.Image(ctx, UploadImageRequest{
				Filename: "image.jpg", ContentType: "image/jpeg", BundleId: "bundle-1",
			})
		},
		func() (Response, error) {
			return client.Webhooks.Create(ctx, CreateWebhookEndpointRequest{
				Url: "https://example.com/webhook", Events: []string{"bundle.published"},
			})
		},
		func() (Response, error) {
			return client.DoOperation(ctx, "createAnalyticsReport", OperationRequest{
				Body: CreateAnalyticsReportRequest{},
			})
		},
	}
	for _, call := range calls {
		if _, err := call(); err == nil || !strings.Contains(err.Error(), "do not accept Idempotency-Key") {
			t.Fatalf("expected local sensitive-response idempotency error, got %v", err)
		}
	}
}

func TestSecretAcceptanceRejectsInvalidPolicyLocally(t *testing.T) {
	client, closeServer := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("invalid acceptance must not reach the API")
	})
	defer closeServer()

	invalid := []CredentialRevealAcceptance{
		{AcknowledgeSupportForfeit: false, PolicyVersion: "managed-credential-reveal-v1"},
		{AcknowledgeSupportForfeit: true, PolicyVersion: "   "},
	}
	for _, acceptance := range invalid {
		if _, err := client.Accounts.RevealCredentialsWithAcceptance(context.Background(), "account-1", acceptance); err == nil {
			t.Fatal("expected invalid reveal acceptance to fail")
		}
		if _, err := client.Accounts.VerificationCodeWithAcceptance(context.Background(), "account-1", acceptance); err == nil {
			t.Fatal("expected invalid verification-code acceptance to fail")
		}
	}
}

func TestDoTextOperationSendsJSONBody(t *testing.T) {
	client, closeServer := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/analytics/export/reports/html" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Accept"); !strings.Contains(got, "text/html") {
			t.Fatalf("accept = %s", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("content-type = %s", got)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), `"title":"Launch"`) {
			t.Fatalf("body = %s", string(body))
		}

		_, _ = w.Write([]byte("<html>ok</html>"))
	})
	defer closeServer()

	resp, err := client.DoTextOperation(context.Background(), "exportAnalyticsReportHtml", OperationRequest{
		Body: map[string]any{"title": "Launch"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp != "<html>ok</html>" {
		t.Fatalf("response = %s", resp)
	}
}

func TestDoOperationMultipart(t *testing.T) {
	tempFile := t.TempDir() + "/image.png"
	if err := os.WriteFile(tempFile, []byte("image-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}

	client, closeServer := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/upload/image/direct" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "multipart/form-data") {
			t.Fatalf("content-type = %s", got)
		}

		if err := r.ParseMultipartForm(1024 * 1024); err != nil {
			t.Fatal(err)
		}
		if got := r.FormValue("bundle_id"); got != "bundle_123" {
			t.Fatalf("bundle_id = %s", got)
		}
		if got := r.FormValue("purpose"); got != "carousel" {
			t.Fatalf("purpose = %s", got)
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		if header.Filename != "image.png" {
			t.Fatalf("filename = %s", header.Filename)
		}

		_ = json.NewEncoder(w).Encode(Response{"data": Response{"ok": true}})
	})
	defer closeServer()

	_, err := client.DoOperation(context.Background(), "uploadImageDirect", OperationRequest{
		Form:        map[string]string{"bundle_id": "bundle_123", "purpose": "carousel"},
		FilePath:    tempFile,
		ContentType: "image/png",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	body := []byte(`{"id":"evt_test"}`)
	timestamp := time.Now().Unix()
	secret := "whsec_test"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d.", timestamp)))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))
	header := fmt.Sprintf("t=%d,v1=%s", timestamp, signature)

	if !VerifyWebhookSignature(body, header, secret, time.Minute) {
		t.Fatal("expected valid webhook signature")
	}

	if VerifyWebhookSignature(body, header, "wrong_secret", time.Minute) {
		t.Fatal("expected invalid webhook signature with wrong secret")
	}

	oldHeader := fmt.Sprintf("t=%d,v1=%s", timestamp-600, signature)
	if VerifyWebhookSignature(body, oldHeader, secret, time.Minute) {
		t.Fatal("expected expired webhook signature")
	}
}
