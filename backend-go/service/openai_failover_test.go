package service

import (
	"errors"
	"testing"
)

func TestIsRetryableError_Network(t *testing.T) {
	cases := []string{
		"dial tcp: lookup api.openai.com: no such host",
		"connection refused",
		"connection reset by peer",
		"i/o timeout",
		"unexpected EOF",
	}
	for _, msg := range cases {
		if !isRetryableError(errors.New(msg)) {
			t.Errorf("expected retryable for: %s", msg)
		}
	}
}

func TestIsRetryableError_429(t *testing.T) {
	if !isRetryableError(errors.New("openai error: 429 Too Many Requests")) {
		t.Error("429 should be retryable")
	}
}

func TestIsRetryableError_5xx(t *testing.T) {
	cases := []string{"500", "502", "503", "504"}
	for _, code := range cases {
		if !isRetryableError(errors.New("openai error: " + code)) {
			t.Errorf("expected retryable for: %s", code)
		}
	}
}

func TestIsRetryableError_NonRetryable(t *testing.T) {
	cases := []string{
		"openai error: 400 Bad Request",
		"openai error: 401 Unauthorized",
		"openai error: 403 Forbidden",
		"openai error: 404 Not Found",
	}
	for _, msg := range cases {
		if isRetryableError(errors.New(msg)) {
			t.Errorf("expected NOT retryable for: %s", msg)
		}
	}
}

func TestIsRetryableError_Nil(t *testing.T) {
	if isRetryableError(nil) {
		t.Error("nil error should not be retryable")
	}
}
