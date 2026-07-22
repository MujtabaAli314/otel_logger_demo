package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/oteldemo/service1-frontend/types"
)

var defaultHTTPClient = &http.Client{Timeout: 10 * time.Second}

// httpError preserves the downstream HTTP status so callers can map it
// to an appropriate response code.
type httpError struct {
	status int
	method string
	url    string
	body   types.ErrorResponse
}

func (e *httpError) Error() string {
	return fmt.Sprintf("downstream %s %s -> %d: %s: %s",
		e.method, e.url, e.status, e.body.Error, e.body.Message)
}

// IsNotFound reports whether the error came from a downstream 404.
func IsNotFound(err error) bool {
	var he *httpError
	if errors.As(err, &he) {
		return he.status == http.StatusNotFound
	}
	return false
}

// doJSON performs an HTTP request against a downstream service, sending
// `in` as the JSON body (when non-nil) and decoding the response into
// `out` (when non-nil).
func doJSON(ctx context.Context, method, fullURL string, in any, out any) error {
	var bodyReader *bytes.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var er types.ErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&er)
		return &httpError{status: resp.StatusCode, method: method, url: fullURL, body: er}
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return err
		}
	}
	return nil
}
