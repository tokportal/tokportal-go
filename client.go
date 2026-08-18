package tokportal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client

	Bundles   *BundlesService
	Credits   *CreditsService
	Accounts  *AccountsService
	Analytics *AnalyticsService
	Uploads   *UploadsService
	Comments  *CommentsService
	Webhooks  *WebhooksService
}

type Option func(*Client)

const SDKVersion = "0.1.0"
const ClientHeader = "tokportal-go/" + SDKVersion

type contextKey string

const idempotencyKeyContextKey contextKey = "tokportal-idempotency-key"

var sensitiveResponseOperationIDs = map[string]struct{}{
	"retrieveAccountVerificationCode": {},
	"revealAccountCredentials":        {},
	"createWebhookEndpoint":           {},
	"uploadImage":                     {},
	"uploadVideo":                     {},
	"createAnalyticsReport":           {},
}

func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyKeyContextKey, key)
}

func rejectSensitiveResponseIdempotency(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	value, _ := ctx.Value(idempotencyKeyContextKey).(string)
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return fmt.Errorf("tokportal: secret-bearing operations do not accept Idempotency-Key because their responses are never stored for replay; remove the key and reconcile safe resource state before retrying")
}

func rejectSensitiveOperationIdempotency(ctx context.Context, operationID string) error {
	if _, sensitive := sensitiveResponseOperationIDs[operationID]; !sensitive {
		return nil
	}
	return rejectSensitiveResponseIdempotency(ctx)
}

func WithBaseURL(baseURL string) Option {
	return func(client *Client) {
		client.baseURL = strings.TrimRight(baseURL, "/")
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) {
		if httpClient != nil {
			client.httpClient = httpClient
		}
	}
}

func NewClient(apiKey string, options ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("tokportal: api key is required")
	}

	client := &Client{
		apiKey:     apiKey,
		baseURL:    BaseURL,
		httpClient: http.DefaultClient,
	}

	for _, option := range options {
		option(client)
	}

	client.Bundles = &BundlesService{client: client}
	client.Credits = &CreditsService{client: client}
	client.Accounts = &AccountsService{client: client}
	client.Analytics = &AnalyticsService{client: client}
	client.Uploads = &UploadsService{client: client}
	client.Comments = &CommentsService{client: client}
	client.Webhooks = &WebhooksService{client: client}

	return client, nil
}

type APIError struct {
	StatusCode        int
	Code              string
	Message           string
	Details           map[string]any
	RequestID         string
	RawBody           string
	RetryAfterSeconds int
	RateLimit         *RateLimit
}

type RateLimit struct {
	Limit     int
	Remaining int
	Reset     int
}

func (error APIError) Error() string {
	if error.Code != "" {
		return fmt.Sprintf("tokportal: %s (%d): %s", error.Code, error.StatusCode, error.Message)
	}
	return fmt.Sprintf("tokportal: request failed (%d): %s", error.StatusCode, error.Message)
}

func (error APIError) Retryable() bool {
	return error.StatusCode == http.StatusRequestTimeout || error.StatusCode == http.StatusTooManyRequests || error.StatusCode >= 500
}

func IsRetryable(err error) bool {
	var apiError APIError
	return errors.As(err, &apiError) && apiError.Retryable()
}

type apiErrorResponse struct {
	Error struct {
		Code    string         `json:"code"`
		Message string         `json:"message"`
		Details map[string]any `json:"details"`
	} `json:"error"`
}

type Query map[string]any
type Response map[string]any
type OperationRequest struct {
	Path        map[string]string
	Query       Query
	Body        any
	Form        map[string]string
	FilePath    string
	FileField   string
	ContentType string
}

func (client *Client) Me(ctx context.Context) (Response, error) {
	return client.request(ctx, http.MethodGet, "/me", nil, nil)
}

func (client *Client) Countries(ctx context.Context) (Response, error) {
	return client.request(ctx, http.MethodGet, "/countries", nil, nil)
}

func (client *Client) Platforms(ctx context.Context) (Response, error) {
	return client.request(ctx, http.MethodGet, "/platforms", nil, nil)
}

func (client *Client) CreditCosts(ctx context.Context) (Response, error) {
	return client.request(ctx, http.MethodGet, "/credit-costs", nil, nil)
}

func (client *Client) DoOperation(ctx context.Context, operationID string, input OperationRequest) (Response, error) {
	operation, ok := OperationDefinitions[operationID]
	if !ok {
		return nil, fmt.Errorf("tokportal: unknown operation id %q", operationID)
	}
	if err := rejectSensitiveOperationIdempotency(ctx, operationID); err != nil {
		return nil, err
	}

	resolvedPath := operation.Path
	for _, name := range operation.PathParams {
		value, ok := input.Path[name]
		if !ok {
			return nil, fmt.Errorf("tokportal: missing path parameter %q for operation %q", name, operationID)
		}
		resolvedPath = strings.ReplaceAll(resolvedPath, "{"+name+"}", url.PathEscape(value))
	}

	if operation.RequestContentType == "multipart/form-data" {
		if input.FilePath == "" {
			return nil, fmt.Errorf("tokportal: missing file path for multipart operation %q", operationID)
		}
		fileField := input.FileField
		if fileField == "" {
			fileField = "file"
		}
		return client.requestMultipart(ctx, operation.Method, resolvedPath, input.Form, fileField, input.FilePath, input.ContentType)
	}
	if operationHasTextSuccess(operation) {
		return nil, fmt.Errorf("tokportal: operation %q returns text; use DoTextOperation", operationID)
	}

	return client.request(ctx, operation.Method, resolvedPath, input.Query, input.Body)
}

func (client *Client) DoTextOperation(ctx context.Context, operationID string, input OperationRequest) (string, error) {
	operation, ok := OperationDefinitions[operationID]
	if !ok {
		return "", fmt.Errorf("tokportal: unknown operation id %q", operationID)
	}
	if err := rejectSensitiveOperationIdempotency(ctx, operationID); err != nil {
		return "", err
	}
	if !operationHasTextSuccess(operation) {
		return "", fmt.Errorf("tokportal: operation %q does not return text", operationID)
	}

	resolvedPath := operation.Path
	for _, name := range operation.PathParams {
		value, ok := input.Path[name]
		if !ok {
			return "", fmt.Errorf("tokportal: missing path parameter %q for operation %q", name, operationID)
		}
		resolvedPath = strings.ReplaceAll(resolvedPath, "{"+name+"}", url.PathEscape(value))
	}

	return client.requestTextWithBody(ctx, operation.Method, resolvedPath, input.Query, input.Body)
}

func operationHasTextSuccess(operation OperationDefinition) bool {
	for _, contentType := range operation.SuccessContentTypes {
		if strings.HasPrefix(contentType, "text/") {
			return true
		}
	}
	return false
}

func (client *Client) request(ctx context.Context, method string, path string, query Query, body any) (Response, error) {
	endpoint, err := url.Parse(client.baseURL + path)
	if err != nil {
		return nil, err
	}

	values := endpoint.Query()
	for key, value := range query {
		if value != nil {
			addQueryValue(values, key, value)
		}
	}
	endpoint.RawQuery = values.Encode()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", client.apiKey)
	req.Header.Set("X-TokPortal-Client", ClientHeader)
	setIdempotencyHeader(req, ctx)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		parsed := apiErrorResponse{}
		_ = json.Unmarshal(raw, &parsed)
		message := parsed.Error.Message
		if message == "" {
			message = string(raw)
		}
		return nil, APIError{
			StatusCode:        resp.StatusCode,
			Code:              parsed.Error.Code,
			Message:           message,
			Details:           parsed.Error.Details,
			RequestID:         requestID(resp),
			RawBody:           string(raw),
			RetryAfterSeconds: headerInt(resp, "Retry-After"),
			RateLimit:         rateLimit(resp),
		}
	}

	if len(raw) == 0 {
		return Response{}, nil
	}

	out := Response{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (client *Client) requestText(ctx context.Context, method string, path string, query Query) (string, error) {
	return client.requestTextWithBody(ctx, method, path, query, nil)
}

func (client *Client) requestTextWithBody(ctx context.Context, method string, path string, query Query, body any) (string, error) {
	endpoint, err := url.Parse(client.baseURL + path)
	if err != nil {
		return "", err
	}

	values := endpoint.Query()
	for key, value := range query {
		if value != nil {
			addQueryValue(values, key, value)
		}
	}
	endpoint.RawQuery = values.Encode()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return "", err
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), reader)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "text/csv, text/html, application/json")
	req.Header.Set("X-API-Key", client.apiKey)
	req.Header.Set("X-TokPortal-Client", ClientHeader)
	setIdempotencyHeader(req, ctx)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		parsed := apiErrorResponse{}
		_ = json.Unmarshal(raw, &parsed)
		message := parsed.Error.Message
		if message == "" {
			message = string(raw)
		}
		return "", APIError{
			StatusCode:        resp.StatusCode,
			Code:              parsed.Error.Code,
			Message:           message,
			Details:           parsed.Error.Details,
			RequestID:         requestID(resp),
			RawBody:           string(raw),
			RetryAfterSeconds: headerInt(resp, "Retry-After"),
			RateLimit:         rateLimit(resp),
		}
	}

	return string(raw), nil
}

func (client *Client) requestMultipart(ctx context.Context, method string, path string, fields map[string]string, fileField string, filePath string, contentType string) (Response, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, err
		}
	}

	partContentType := contentType
	if partContentType == "" {
		partContentType = "application/octet-stream"
	}
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fileField, filepath.Base(filePath)))
	header.Set("Content-Type", partContentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, client.baseURL+path, &body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Key", client.apiKey)
	req.Header.Set("X-TokPortal-Client", ClientHeader)
	setIdempotencyHeader(req, ctx)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		parsed := apiErrorResponse{}
		_ = json.Unmarshal(raw, &parsed)
		message := parsed.Error.Message
		if message == "" {
			message = string(raw)
		}
		return nil, APIError{
			StatusCode:        resp.StatusCode,
			Code:              parsed.Error.Code,
			Message:           message,
			Details:           parsed.Error.Details,
			RequestID:         requestID(resp),
			RawBody:           string(raw),
			RetryAfterSeconds: headerInt(resp, "Retry-After"),
			RateLimit:         rateLimit(resp),
		}
	}

	if len(raw) == 0 {
		return Response{}, nil
	}

	out := Response{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func requestID(resp *http.Response) string {
	if value := resp.Header.Get("x-tokportal-request-id"); value != "" {
		return value
	}
	if value := resp.Header.Get("x-request-id"); value != "" {
		return value
	}
	if value := resp.Header.Get("request-id"); value != "" {
		return value
	}
	return resp.Header.Get("x-vercel-id")
}

func headerInt(resp *http.Response, name string) int {
	value := resp.Header.Get(name)
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}

func rateLimit(resp *http.Response) *RateLimit {
	limit := headerInt(resp, "X-RateLimit-Limit")
	remaining := headerInt(resp, "X-RateLimit-Remaining")
	reset := headerInt(resp, "X-RateLimit-Reset")
	if limit == 0 && remaining == 0 && reset == 0 {
		return nil
	}
	return &RateLimit{
		Limit:     limit,
		Remaining: remaining,
		Reset:     reset,
	}
}

func setIdempotencyHeader(req *http.Request, ctx context.Context) {
	value, ok := ctx.Value(idempotencyKeyContextKey).(string)
	if ok && value != "" {
		req.Header.Set("Idempotency-Key", value)
	}
}

func addQueryValue(values url.Values, key string, value any) {
	switch typed := value.(type) {
	case []string:
		for _, item := range typed {
			values.Add(key, item)
		}
	case []int:
		for _, item := range typed {
			values.Add(key, fmt.Sprint(item))
		}
	case []bool:
		for _, item := range typed {
			values.Add(key, fmt.Sprint(item))
		}
	case []any:
		for _, item := range typed {
			values.Add(key, fmt.Sprint(item))
		}
	default:
		values.Set(key, fmt.Sprint(value))
	}
}
