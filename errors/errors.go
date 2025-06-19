package errors

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidURL      = errors.New("invalid URL")
	ErrPageLoad        = errors.New("failed to load page")
	ErrTimeout         = errors.New("operation timed out")
	ErrProxyFailure    = errors.New("proxy connection failed")
	ErrCloudflareBlock = errors.New("blocked by Cloudflare protection")
	ErrLLMAPIFailure   = errors.New("LLM API call failed")
)

func WithCause(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

func IsType(err, target error) bool {
	return errors.Is(err, target)
}