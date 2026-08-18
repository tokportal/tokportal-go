package tokportal

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

func VerifyWebhookSignature(rawBody []byte, signatureHeader string, signingSecret string, tolerance time.Duration) bool {
	parts := map[string]string{}
	for _, part := range strings.Split(signatureHeader, ",") {
		key, value, found := strings.Cut(strings.TrimSpace(part), "=")
		if found && key != "" && value != "" {
			parts[key] = value
		}
	}

	timestamp := parts["t"]
	signature := parts["v1"]
	if timestamp == "" || signature == "" || signingSecret == "" {
		return false
	}

	timestampSeconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	if tolerance <= 0 {
		tolerance = 5 * time.Minute
	}
	if time.Since(time.Unix(timestampSeconds, 0)) > tolerance || time.Until(time.Unix(timestampSeconds, 0)) > tolerance {
		return false
	}

	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(rawBody)
	expected := mac.Sum(nil)

	received, err := hex.DecodeString(signature)
	if err != nil || len(received) != len(expected) {
		return false
	}

	return hmac.Equal(received, expected)
}
