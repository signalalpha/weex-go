package weexgo

import "fmt"

// APIError represents an error returned by the WEEX API
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  int    // HTTP status code
}

func (e *APIError) Error() string {
	if e.Status > 0 {
		return fmt.Sprintf("API error [%d] (HTTP %d): %s", e.Code, e.Status, e.Message)
	}
	return fmt.Sprintf("API error [%d]: %s", e.Code, e.Message)
}

// IsAPIError checks if an error is an APIError
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// HTTPError represents an HTTP-level error
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP error [%d]: %s", e.StatusCode, e.Body)
}

// IsHTTPError checks if an error is an HTTPError
func IsHTTPError(err error) bool {
	_, ok := err.(*HTTPError)
	return ok
}
