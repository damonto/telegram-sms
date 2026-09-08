package license

import (
	"crypto/tls"
	"errors"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func authorizationRetryDelay(err error, backoff time.Duration) (time.Duration, bool) {
	if !retryableAuthorizationError(err) {
		return 0, false
	}
	// Retain half the delay so jitter cannot turn an outage into a busy loop.
	delay := backoff/2 + time.Duration(rand.Int64N(int64(backoff-backoff/2)))
	var remote *serviceError
	if errors.As(err, &remote) {
		delay = max(delay, time.Until(remote.RetryAt))
	}
	return delay, true
}

func retryableAuthorizationError(err error) bool {
	var remote *serviceError
	if errors.As(err, &remote) {
		if isExplicitAuthorizationError(remote.ErrorCode) {
			return false
		}
		switch remote.ErrorCode {
		case "license_challenge_invalid", "license_challenge_expired":
			// A new transport challenge can retry the same pending rotation.
			return true
		}
		switch remote.StatusCode {
		case http.StatusRequestTimeout, http.StatusTooManyRequests,
			http.StatusInternalServerError, http.StatusBadGateway,
			http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return true
		default:
			return false
		}
	}
	if _, ok := errors.AsType[*tls.CertificateVerificationError](err); ok {
		return false
	}
	// url.Error implements net.Error even when its cause is a permanent error.
	if requestErr, ok := errors.AsType[*url.Error](err); ok {
		err = requestErr.Err
	}
	var networkErr net.Error
	return errors.As(err, &networkErr) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF)
}

func parseRetryAfter(value string, now time.Time) time.Time {
	value = strings.TrimSpace(value)
	if seconds, err := strconv.ParseInt(value, 10, 32); err == nil && seconds >= 0 {
		return now.Add(time.Duration(seconds) * time.Second)
	}
	if retryAt, err := http.ParseTime(value); err == nil && retryAt.After(now) {
		return retryAt
	}
	return time.Time{}
}
