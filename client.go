package weexgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client represents the WEEX API client
type Client struct {
	apiKey     string
	secretKey  string
	passphrase string
	baseURL    string
	httpClient *http.Client
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client) error

// WithAPIKey sets the API key
func WithAPIKey(apiKey string) ClientOption {
	return func(c *Client) error {
		c.apiKey = apiKey
		return nil
	}
}

// WithSecretKey sets the secret key
func WithSecretKey(secretKey string) ClientOption {
	return func(c *Client) error {
		c.secretKey = secretKey
		return nil
	}
}

// WithPassphrase sets the passphrase
func WithPassphrase(passphrase string) ClientOption {
	return func(c *Client) error {
		c.passphrase = passphrase
		return nil
	}
}

// WithBaseURL sets the base URL
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		c.baseURL = baseURL
		return nil
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) error {
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Timeout = timeout
		return nil
	}
}

// WithProxy sets the HTTP proxy URL
// Proxy format examples:
//   - Without auth: http://proxy.example.com:3128
//   - With auth: http://username:password@proxy.example.com:3128
func WithProxy(proxyURL string) ClientOption {
	return func(c *Client) error {
		if proxyURL == "" {
			return nil
		}
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		// Parse proxy URL
		proxyURLParsed, err := url.Parse(proxyURL)
		if err != nil {
			return fmt.Errorf("invalid proxy URL: %w", err)
		}
		// Set proxy transport
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURLParsed),
		}
		c.httpClient.Transport = transport
		return nil
	}
}

// NewClient creates a new WEEX API client with the given options
func NewClient(opts ...ClientOption) (*Client, error) {
	cfg := DefaultConfig()

	client := &Client{
		apiKey:     cfg.APIKey,
		secretKey:  cfg.SecretKey,
		passphrase: cfg.Passphrase,
		baseURL:    cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Validate configuration
	if err := client.validate(); err != nil {
		return nil, err
	}

	return client, nil
}

// validate validates the client configuration
func (c *Client) validate() error {
	if c.apiKey == "" {
		return &ConfigError{Field: "APIKey", Message: "API key is required"}
	}
	if c.secretKey == "" {
		return &ConfigError{Field: "SecretKey", Message: "Secret key is required"}
	}
	if c.passphrase == "" {
		return &ConfigError{Field: "Passphrase", Message: "Passphrase is required"}
	}
	return nil
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, path string, queryString string, body interface{}) (*http.Response, error) {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	url := c.baseURL + path
	if queryString != "" {
		// If queryString already starts with "?", use it directly; otherwise add "?"
		if queryString[0] == '?' {
			url += queryString
		} else {
			url += "?" + queryString
		}
	}

	var reqBody string
	if bodyBytes != nil {
		reqBody = string(bodyBytes)
	}

	// Get authentication headers
	headers, err := GetAuthHeaders(c.apiKey, c.secretKey, c.passphrase, method, path, queryString, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth headers: %w", err)
	}

	// Create request
	var req *http.Request
	if bodyBytes != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Add User-Agent header
	req.Header.Set("User-Agent", "weex-go/1.0")

	// Perform request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	return resp, nil
}

// parseResponse parses the API response
func (c *Client) parseResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// For 521 errors, provide more helpful error message
		if resp.StatusCode == 521 {
			return &HTTPError{
				StatusCode: resp.StatusCode,
				Body:       fmt.Sprintf("Web Server Is Down. This usually means: 1) Your IP is not whitelisted in WEEX API settings, 2) API endpoint is unreachable. Response body: %s", string(bodyBytes)),
			}
		}

		var errResp ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errResp); err == nil {
			return &APIError{
				Code:    errResp.Code,
				Message: errResp.Message,
				Status:  resp.StatusCode,
			}
		}

		return &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(bodyBytes),
		}
	}

	// Try to parse as APIResponse first
	var apiResp APIResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err == nil {
		if apiResp.Code != 0 {
			return &APIError{
				Code:    apiResp.Code,
				Message: apiResp.Message,
				Status:  resp.StatusCode,
			}
		}
		// Unmarshal data field
		if apiResp.Data != nil && result != nil {
			dataBytes, err := json.Marshal(apiResp.Data)
			if err != nil {
				return fmt.Errorf("failed to marshal data: %w", err)
			}
			return json.Unmarshal(dataBytes, result)
		}
		return nil
	}

	// Try direct unmarshal
	if result != nil {
		return json.Unmarshal(bodyBytes, result)
	}

	return nil
}

// GetAccountAssets retrieves account assets information
// Returns a slice of AccountBalance (direct array response from /capi/v2/account/assets)
func (c *Client) GetAccountAssets() (AccountAssets, error) {
	path := "/capi/v2/account/assets"
	resp, err := c.doRequest("GET", path, "", nil)
	if err != nil {
		return nil, err
	}

	var assets AccountAssets
	if err := c.parseResponse(resp, &assets); err != nil {
		return nil, err
	}

	return assets, nil
}

// GetTicker retrieves ticker information for a symbol
func (c *Client) GetTicker(symbol string) (*Ticker, error) {
	path := "/capi/v2/market/ticker"
	// Query string must include "?" for signature generation (matching Python implementation)
	queryString := fmt.Sprintf("?symbol=%s", symbol)
	resp, err := c.doRequest("GET", path, queryString, nil)
	if err != nil {
		return nil, err
	}

	var ticker Ticker
	if err := c.parseResponse(resp, &ticker); err != nil {
		return nil, err
	}

	return &ticker, nil
}

// CreateOrder creates a new order
func (c *Client) CreateOrder(req *CreateOrderRequest) (*Order, error) {
	path := "/capi/v2/order/placeOrder"
	resp, err := c.doRequest("POST", path, "", req)
	if err != nil {
		return nil, err
	}

	var order Order
	if err := c.parseResponse(resp, &order); err != nil {
		return nil, err
	}

	return &order, nil
}

// GetOrder retrieves order information by order ID
func (c *Client) GetOrder(orderID string) (*Order, error) {
	path := "/capi/v2/order"
	queryString := fmt.Sprintf("orderId=%s", orderID)
	resp, err := c.doRequest("GET", path, queryString, nil)
	if err != nil {
		return nil, err
	}

	var order Order
	if err := c.parseResponse(resp, &order); err != nil {
		return nil, err
	}

	return &order, nil
}

// CancelOrder cancels an order
func (c *Client) CancelOrder(orderID string) error {
	path := "/capi/v2/order/cancel_order"
	body := map[string]string{
		"orderId": orderID,
	}
	resp, err := c.doRequest("POST", path, "", body)
	if err != nil {
		return err
	}

	return c.parseResponse(resp, nil)
}

// GetCurrentOrders retrieves current active orders for a symbol
// Response format from /capi/v2/order/current:
// - Direct array: [Order, ...]
// - Or wrapped: {"data": [Order, ...]} or {"list": [Order, ...]}
func (c *Client) GetCurrentOrders(symbol string) ([]Order, error) {
	path := "/capi/v2/order/current"
	// Query string must include "?" for signature generation (matching Python implementation)
	queryString := fmt.Sprintf("?symbol=%s", symbol)
	resp, err := c.doRequest("GET", path, queryString, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// For 521 errors, provide more helpful error message
		if resp.StatusCode == 521 {
			return nil, &HTTPError{
				StatusCode: resp.StatusCode,
				Body:       fmt.Sprintf("Web Server Is Down. This usually means: 1) Your IP is not whitelisted in WEEX API settings, 2) API endpoint is unreachable. Response body: %s", string(bodyBytes)),
			}
		}

		var errResp ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errResp); err == nil {
			return nil, &APIError{
				Code:    errResp.Code,
				Message: errResp.Message,
				Status:  resp.StatusCode,
			}
		}

		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(bodyBytes),
		}
	}

	// Try to parse as direct array first
	var orders []Order
	if err := json.Unmarshal(bodyBytes, &orders); err == nil {
		return orders, nil
	}

	// Try to parse as wrapped response
	type OrdersResponse struct {
		Data []Order `json:"data,omitempty"`
		List []Order `json:"list,omitempty"`
	}

	var wrappedResp OrdersResponse
	if err := json.Unmarshal(bodyBytes, &wrappedResp); err == nil {
		if len(wrappedResp.Data) > 0 {
			return wrappedResp.Data, nil
		}
		if len(wrappedResp.List) > 0 {
			return wrappedResp.List, nil
		}
		// Empty wrapped response, return empty array
		return []Order{}, nil
	}

	// If neither format works, return empty array (no orders)
	return []Order{}, nil
}

// SetLeverage sets leverage for a symbol
func (c *Client) SetLeverage(symbol string, marginMode int, longLeverage, shortLeverage string) error {
	path := "/capi/v2/account/leverage"
	body := map[string]interface{}{
		"symbol":        symbol,
		"marginMode":    marginMode,
		"longLeverage":  longLeverage,
		"shortLeverage": shortLeverage,
	}
	resp, err := c.doRequest("POST", path, "", body)
	if err != nil {
		return err
	}

	return c.parseResponse(resp, nil)
}
