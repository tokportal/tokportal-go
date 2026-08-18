package tokportal

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type BundlesService struct {
	client *Client
}

func (service *BundlesService) List(ctx context.Context, query Query) (Response, error) {
	return service.client.request(ctx, http.MethodGet, "/bundles", query, nil)
}

func (service *BundlesService) Create(ctx context.Context, body CreateBundleRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPost, "/bundles", nil, body)
}

func (service *BundlesService) BulkCreate(ctx context.Context, body CreateBulkBundlesRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPost, "/bundles/bulk", nil, body)
}

func (service *BundlesService) Get(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/bundles/%s", url.PathEscape(id)), nil, nil)
}

func (service *BundlesService) Update(ctx context.Context, id string, body PatchBundleRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPatch, fmt.Sprintf("/bundles/%s", url.PathEscape(id)), nil, body)
}

func (service *BundlesService) Publish(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/bundles/%s/publish", url.PathEscape(id)), nil, nil)
}

func (service *BundlesService) Readiness(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/bundles/%s/publish-readiness", url.PathEscape(id)), nil, nil)
}

func (service *BundlesService) Unpublish(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/bundles/%s/unpublish", url.PathEscape(id)), nil, nil)
}

func (service *BundlesService) AddVideoSlots(ctx context.Context, id string, quantity int) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/bundles/%s/add-video-slots", url.PathEscape(id)), nil, QuantityRequest{Quantity: quantity})
}

func (service *BundlesService) AddEditSlots(ctx context.Context, id string, quantity int) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/bundles/%s/add-edit-slots", url.PathEscape(id)), nil, QuantityRequest{Quantity: quantity})
}

func (service *BundlesService) GetAccount(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/bundles/%s/account", url.PathEscape(id)), nil, nil)
}

func (service *BundlesService) ConfigureAccount(ctx context.Context, id string, body ConfigureAccountRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPut, fmt.Sprintf("/bundles/%s/account", url.PathEscape(id)), nil, body)
}

func (service *BundlesService) RequestAccountCorrections(ctx context.Context, id string, body AccountCorrectionsRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/bundles/%s/account/corrections", url.PathEscape(id)), nil, body)
}

func (service *BundlesService) FinalizeAccount(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/bundles/%s/account/finalize", url.PathEscape(id)), nil, nil)
}

func (service *BundlesService) ListVideos(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/bundles/%s/videos", url.PathEscape(id)), nil, nil)
}

func (service *BundlesService) GetVideo(ctx context.Context, id string, position int) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/bundles/%s/videos/%d", url.PathEscape(id), position), nil, nil)
}

func (service *BundlesService) ConfigureVideo(ctx context.Context, id string, position int, body ConfigureVideoRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPut, fmt.Sprintf("/bundles/%s/videos/%d", url.PathEscape(id), position), nil, body)
}

func (service *BundlesService) PatchVideo(ctx context.Context, id string, position int, body PatchVideoRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPatch, fmt.Sprintf("/bundles/%s/videos/%d", url.PathEscape(id), position), nil, body)
}

func (service *BundlesService) BatchConfigureVideos(ctx context.Context, id string, body BatchConfigureVideosRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPut, fmt.Sprintf("/bundles/%s/videos/batch", url.PathEscape(id)), nil, body)
}

func (service *BundlesService) PublishAllVideos(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/bundles/%s/videos/publish-all", url.PathEscape(id)), nil, nil)
}

func (service *BundlesService) PublishVideo(ctx context.Context, id string, position int) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/bundles/%s/videos/%d/publish", url.PathEscape(id), position), nil, nil)
}

func (service *BundlesService) ResetVideo(ctx context.Context, id string, position int) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/bundles/%s/videos/%d/reset", url.PathEscape(id), position), nil, nil)
}

func (service *BundlesService) UnscheduleVideo(ctx context.Context, id string, position int) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/bundles/%s/videos/%d/unschedule", url.PathEscape(id), position), nil, nil)
}

func (service *BundlesService) FinalizeVideo(ctx context.Context, id string, position int) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/bundles/%s/videos/%d/finalize", url.PathEscape(id), position), nil, nil)
}

func (service *BundlesService) RequestVideoCorrections(ctx context.Context, id string, position int, body VideoCorrectionsRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/bundles/%s/videos/%d/corrections", url.PathEscape(id), position), nil, body)
}

func (service *BundlesService) FixVideoDownload(ctx context.Context, id string, position int, body FixVideoDownloadRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/bundles/%s/videos/%d/fix-download", url.PathEscape(id), position), nil, body)
}

type CreditsService struct {
	client *Client
}

func (service *CreditsService) Balance(ctx context.Context) (Response, error) {
	return service.client.request(ctx, http.MethodGet, "/credits/balance", nil, nil)
}

func (service *CreditsService) History(ctx context.Context, query Query) (Response, error) {
	return service.client.request(ctx, http.MethodGet, "/credits/history", query, nil)
}

type AccountsService struct {
	client *Client
}

type CredentialRevealAcceptance struct {
	AcknowledgeSupportForfeit bool   `json:"acknowledge_support_forfeit"`
	PolicyVersion             string `json:"policy_version"`
}

type ManagedAccountSubscriptionReactivationRequest struct {
	ExpectedCredits          int    `json:"expected_credits"`
	ExpectedCurrentPeriodEnd string `json:"expected_current_period_end"`
	ExpectedLockVersion      int    `json:"expected_lock_version"`
}

func (service *AccountsService) List(ctx context.Context, query Query) (Response, error) {
	return service.client.request(ctx, http.MethodGet, "/accounts", query, nil)
}

func (service *AccountsService) Get(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/accounts/%s", url.PathEscape(id)), nil, nil)
}

func (service *AccountsService) Bundles(ctx context.Context, id string, query Query) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/accounts/%s/bundles", url.PathEscape(id)), query, nil)
}

func (service *AccountsService) VerificationCode(ctx context.Context, id string) (Response, error) {
	if err := rejectSensitiveResponseIdempotency(ctx); err != nil {
		return nil, err
	}
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/accounts/%s/verification-code", url.PathEscape(id)), nil, nil)
}

func (service *AccountsService) VerificationCodeWithAcceptance(ctx context.Context, id string, body CredentialRevealAcceptance) (Response, error) {
	if err := rejectSensitiveResponseIdempotency(ctx); err != nil {
		return nil, err
	}
	if !body.AcknowledgeSupportForfeit || strings.TrimSpace(body.PolicyVersion) == "" {
		return nil, fmt.Errorf("credential reveal acceptance requires acknowledge_support_forfeit=true and a non-empty policy_version returned by the 428 preview")
	}
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/accounts/%s/verification-code", url.PathEscape(id)), nil, body)
}

func (service *AccountsService) RevealCredentials(ctx context.Context, id string) (Response, error) {
	if err := rejectSensitiveResponseIdempotency(ctx); err != nil {
		return nil, err
	}
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/accounts/%s/reveal-credentials", url.PathEscape(id)), nil, nil)
}

func (service *AccountsService) RevealCredentialsWithAcceptance(ctx context.Context, id string, body CredentialRevealAcceptance) (Response, error) {
	if err := rejectSensitiveResponseIdempotency(ctx); err != nil {
		return nil, err
	}
	if !body.AcknowledgeSupportForfeit || strings.TrimSpace(body.PolicyVersion) == "" {
		return nil, fmt.Errorf("credential reveal acceptance requires acknowledge_support_forfeit=true and a non-empty policy_version returned by the 428 preview")
	}
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/accounts/%s/reveal-credentials", url.PathEscape(id)), nil, body)
}

func (service *AccountsService) Coverage(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/accounts/%s/managed-subscription", url.PathEscape(id)), nil, nil)
}

func (service *AccountsService) ReactivateCoverage(ctx context.Context, id string, body ManagedAccountSubscriptionReactivationRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/accounts/%s/managed-subscription/reactivate", url.PathEscape(id)), nil, body)
}

func (service *AccountsService) PauseCoverage(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/accounts/%s/managed-subscription/cancel", url.PathEscape(id)), nil, nil)
}

func (service *AccountsService) CanRefreshAnalytics(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/accounts/%s/analytics/can-refresh", url.PathEscape(id)), nil, nil)
}

func (service *AccountsService) RefreshAnalytics(ctx context.Context, id string, body RefreshAnalyticsRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/accounts/%s/analytics/refresh", url.PathEscape(id)), nil, body)
}

func (service *AccountsService) GetEditRequest(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/accounts/%s/edit-request", url.PathEscape(id)), nil, nil)
}

func (service *AccountsService) CreateEditRequest(ctx context.Context, id string, body AccountEditRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/accounts/%s/edit-request", url.PathEscape(id)), nil, body)
}

type AnalyticsService struct {
	client *Client
}

func (service *AnalyticsService) Dashboard(ctx context.Context, query Query) (Response, error) {
	return service.client.request(ctx, http.MethodGet, "/analytics", query, nil)
}

func (service *AnalyticsService) Contract(ctx context.Context) (Response, error) {
	return service.client.request(ctx, http.MethodGet, "/analytics/contract", nil, nil)
}

func (service *AnalyticsService) Series(ctx context.Context, query Query) (Response, error) {
	return service.client.request(ctx, http.MethodGet, "/analytics/series", query, nil)
}

func (service *AnalyticsService) Account(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/analytics/accounts/%s", url.PathEscape(id)), nil, nil)
}

func (service *AnalyticsService) RefreshAccount(ctx context.Context, id string, body RefreshAnalyticsRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/analytics/accounts/%s/refresh", url.PathEscape(id)), nil, body)
}

func (service *AnalyticsService) AccountRaw(ctx context.Context, id string, query Query) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/analytics/accounts/%s/raw", url.PathEscape(id)), query, nil)
}

func (service *AnalyticsService) AccountCompatibility(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/accounts/%s/analytics", url.PathEscape(id)), nil, nil)
}

func (service *AnalyticsService) AccountVideos(ctx context.Context, id string, query Query) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/accounts/%s/analytics/videos", url.PathEscape(id)), query, nil)
}

func (service *AnalyticsService) Video(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/videos/%s/analytics", url.PathEscape(id)), nil, nil)
}

func (service *AnalyticsService) CommentPulse(ctx context.Context, query Query) (Response, error) {
	return service.client.request(ctx, http.MethodGet, "/analytics/comments", query, nil)
}

func (service *AnalyticsService) AccountComments(ctx context.Context, id string, query Query) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/analytics/accounts/%s/comments", url.PathEscape(id)), query, nil)
}

func (service *AnalyticsService) PostRaw(ctx context.Context, id string, query Query) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/analytics/posts/%s/raw", url.PathEscape(id)), query, nil)
}

func (service *AnalyticsService) ExportVideos(ctx context.Context, query Query) (string, error) {
	return service.client.requestText(ctx, http.MethodGet, "/analytics/export/videos", query)
}

func (service *AnalyticsService) CreateReport(ctx context.Context, body CreateAnalyticsReportRequest) (Response, error) {
	if err := rejectSensitiveResponseIdempotency(ctx); err != nil {
		return nil, err
	}
	return service.client.request(ctx, http.MethodPost, "/analytics/export/reports", nil, body)
}

func (service *AnalyticsService) ExportReportHTML(ctx context.Context, body CreateAnalyticsReportRequest) (string, error) {
	return service.client.requestTextWithBody(ctx, http.MethodPost, "/analytics/export/reports/html", nil, body)
}

type UploadsService struct {
	client *Client
}

func (service *UploadsService) Video(ctx context.Context, body UploadVideoRequest) (Response, error) {
	if err := rejectSensitiveResponseIdempotency(ctx); err != nil {
		return nil, err
	}
	return service.client.request(ctx, http.MethodPost, "/upload/video", nil, body)
}

func (service *UploadsService) Image(ctx context.Context, body UploadImageRequest) (Response, error) {
	if err := rejectSensitiveResponseIdempotency(ctx); err != nil {
		return nil, err
	}
	return service.client.request(ctx, http.MethodPost, "/upload/image", nil, body)
}

func (service *UploadsService) ImageFromURL(ctx context.Context, body UploadImageFromUrlRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPost, "/upload/image/from-url", nil, body)
}

func (service *UploadsService) VideoDirect(ctx context.Context, filePath string, bundleID string, contentType string) (Response, error) {
	return service.client.requestMultipart(ctx, http.MethodPost, "/upload/video/direct", map[string]string{"bundle_id": bundleID}, "file", filePath, contentType)
}

func (service *UploadsService) ImageDirect(ctx context.Context, filePath string, bundleID string, purpose string, contentType string) (Response, error) {
	if purpose == "" {
		purpose = "carousel"
	}
	return service.client.requestMultipart(ctx, http.MethodPost, "/upload/image/direct", map[string]string{"bundle_id": bundleID, "purpose": purpose}, "file", filePath, contentType)
}

type CommentsService struct {
	client *Client
}

func (service *CommentsService) List(ctx context.Context, query Query) (Response, error) {
	return service.client.request(ctx, http.MethodGet, "/comments", query, nil)
}

func (service *CommentsService) Create(ctx context.Context, body CreateCommentTasksRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPost, "/comments", nil, body)
}

func (service *CommentsService) Get(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/comments/%s", url.PathEscape(id)), nil, nil)
}

func (service *CommentsService) Delete(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodDelete, fmt.Sprintf("/comments/%s", url.PathEscape(id)), nil, nil)
}

func (service *CommentsService) Approve(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/comments/%s/approve", url.PathEscape(id)), nil, nil)
}

func (service *CommentsService) Dispute(ctx context.Context, id string, reason string) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/comments/%s/dispute", url.PathEscape(id)), nil, DisputeCommentRequest{Reason: reason})
}

func (service *CommentsService) Verifications(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/comments/%s/verifications", url.PathEscape(id)), nil, nil)
}

type WebhooksService struct {
	client *Client
}

func (service *WebhooksService) Events(ctx context.Context) (Response, error) {
	return service.client.request(ctx, http.MethodGet, "/webhooks/events", nil, nil)
}

func (service *WebhooksService) List(ctx context.Context, query Query) (Response, error) {
	return service.client.request(ctx, http.MethodGet, "/webhooks", query, nil)
}

func (service *WebhooksService) Create(ctx context.Context, body CreateWebhookEndpointRequest) (Response, error) {
	if err := rejectSensitiveResponseIdempotency(ctx); err != nil {
		return nil, err
	}
	return service.client.request(ctx, http.MethodPost, "/webhooks", nil, body)
}

func (service *WebhooksService) Get(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/webhooks/%s", url.PathEscape(id)), nil, nil)
}

func (service *WebhooksService) Update(ctx context.Context, id string, body UpdateWebhookEndpointRequest) (Response, error) {
	return service.client.request(ctx, http.MethodPatch, fmt.Sprintf("/webhooks/%s", url.PathEscape(id)), nil, body)
}

func (service *WebhooksService) Delete(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodDelete, fmt.Sprintf("/webhooks/%s", url.PathEscape(id)), nil, nil)
}

func (service *WebhooksService) Deliveries(ctx context.Context, id string, query Query) (Response, error) {
	return service.client.request(ctx, http.MethodGet, fmt.Sprintf("/webhooks/%s/deliveries", url.PathEscape(id)), query, nil)
}

func (service *WebhooksService) RetryDelivery(ctx context.Context, id string, deliveryID string) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/webhooks/%s/deliveries/%s/retry", url.PathEscape(id), url.PathEscape(deliveryID)), nil, nil)
}

func (service *WebhooksService) Test(ctx context.Context, id string) (Response, error) {
	return service.client.request(ctx, http.MethodPost, fmt.Sprintf("/webhooks/%s/test", url.PathEscape(id)), nil, nil)
}
