package api

import "errors"

var errCardNotFound = errors.New("card not found")

// RejectedError is a safe, machine-readable error for a request not accepted
// into the durable operation journal.
type RejectedError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *RejectedError) Error() string { return e.Message }

func reject(code, message string) error { return &RejectedError{Code: code, Message: message} }

func classifyCardLoadError(err error) error {
	if errors.Is(err, errCardNotFound) {
		return reject("not_found", "card was not found")
	}
	return reject("registry_unavailable", "card registry is temporarily unavailable")
}

// SafeRejection maps pre-acceptance failures to a stable, diagnostic-free contract.
func SafeRejection(err error) *RejectedError {
	if rejected, ok := errors.AsType[*RejectedError](err); ok {
		return rejected
	}
	return &RejectedError{Code: "request_rejected", Message: "request was not accepted; verify authorization, inputs, and operation state"}
}
