package ai

import (
	"errors"
	"fmt"
	"strings"
)

type ErrorKind string

const (
	ErrNoActiveConfig  ErrorKind = "no_active_config"
	ErrArticleFetch    ErrorKind = "article_fetch"
	ErrRequestBuild    ErrorKind = "request_build"
	ErrRequestFailed   ErrorKind = "request_failed"
	ErrUpstreamStatus  ErrorKind = "upstream_status"
	ErrInvalidResponse ErrorKind = "invalid_response"
	ErrResponseParse   ErrorKind = "response_parse"
)

// Error is a classified error used across AI manager/provider paths.
type Error struct {
	Kind         ErrorKind
	Op           string
	StatusCode   int
	PublicReason string
	Err          error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	reason := strings.TrimSpace(e.PublicReason)
	if reason == "" {
		reason = "ai operation failed"
	}

	switch {
	case e.Op != "" && e.Err != nil:
		return fmt.Sprintf("%s: %s: %v", e.Op, reason, e.Err)
	case e.Op != "":
		return fmt.Sprintf("%s: %s", e.Op, reason)
	case e.Err != nil:
		return fmt.Sprintf("%s: %v", reason, e.Err)
	default:
		return reason
	}
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func AsError(err error) (*Error, bool) {
	var aiErr *Error
	if !errors.As(err, &aiErr) {
		return nil, false
	}
	return aiErr, true
}
