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
// Converts Go SDK format to WEEX API format
func (c *Client) CreateOrder(req *CreateOrderRequest) (*Order, error) {
	path := "/capi/v2/order/placeOrder"

	// Generate client_oid (timestamp in milliseconds)
	clientOID := fmt.Sprintf("%d", time.Now().UnixMilli())

	// Convert Side to API format: "buy" -> "1" (开多), "sell" -> "2" (开空)
	var sideType string
	if req.Side == OrderSideBuy {
		sideType = "1"
	} else if req.Side == OrderSideSell {
		sideType = "2"
	} else {
		return nil, fmt.Errorf("invalid order side: %s", req.Side)
	}

	// Convert OrderType to match_price: "market" -> "1", "limit" -> "0"
	var matchPrice string
	if req.OrderType == OrderTypeMarket {
		matchPrice = "1"
	} else if req.OrderType == OrderTypeLimit {
		matchPrice = "0"
	} else {
		return nil, fmt.Errorf("invalid order type: %s", req.OrderType)
	}

	// Build API request body in WEEX format
	apiBody := map[string]interface{}{
		"symbol":      req.Symbol,
		"client_oid":  clientOID,
		"size":        req.Quantity,
		"type":        sideType,
		"order_type":  "0", // Default: 普通订单
		"match_price": matchPrice,
	}

	// Add price for limit orders
	if req.OrderType == OrderTypeLimit {
		if req.Price == "" {
			return nil, fmt.Errorf("price is required for limit orders")
		}
		apiBody["price"] = req.Price
	} else {
		// For market orders, price is still required by API but may not be used
		// Use a placeholder if not provided
		if req.Price == "" {
			apiBody["price"] = "0"
		} else {
			apiBody["price"] = req.Price
		}
	}

	resp, err := c.doRequest("POST", path, "", apiBody)
	if err != nil {
		return nil, err
	}

	// Read response body to parse manually (API may return order_id directly)
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
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

	// Try to parse response - API may return order_id directly or wrapped
	var order Order

	// First try to parse as APIResponse wrapper
	var apiResp APIResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err == nil && apiResp.Code == 0 {
		// Success response with data field
		if apiResp.Data != nil {
			dataBytes, err := json.Marshal(apiResp.Data)
			if err == nil {
				json.Unmarshal(dataBytes, &order)
			}
		}
	} else {
		// Try direct unmarshal (response is the order object directly)
		json.Unmarshal(bodyBytes, &order)
	}

	// If OrderID is empty but OrderIDAlt is set, copy it
	if order.OrderID == "" && order.OrderIDAlt != "" {
		order.OrderID = order.OrderIDAlt
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

// TradeFillsOptions contains optional parameters for GetTradeFills
type TradeFillsOptions struct {
	OrderId   string // optional order ID to filter by
	StartTime int64  // optional start time in milliseconds
	EndTime   int64  // optional end time in milliseconds
	Limit     int    // optional limit (default 100, max 100)
}

// GetTradeFillsDebug retrieves filled orders with debug info (returns query string and raw response)
func (c *Client) GetTradeFillsDebug(symbol string, opts *TradeFillsOptions) (TradeFillsResult, string, string, error) {
	path := "/capi/v2/order/fills"

	limit := 100
	if opts != nil && opts.Limit > 0 {
		limit = opts.Limit
		if limit > 100 {
			limit = 100
		}
	}

	queryString := fmt.Sprintf("?symbol=%s&limit=%d", symbol, limit)
	if opts != nil {
		if opts.OrderId != "" {
			queryString += fmt.Sprintf("&orderId=%s", opts.OrderId)
		}
		if opts.StartTime > 0 {
			queryString += fmt.Sprintf("&startTime=%d", opts.StartTime)
		}
		if opts.EndTime > 0 {
			queryString += fmt.Sprintf("&endTime=%d", opts.EndTime)
		}
	}

	url := c.baseURL + path + queryString
	resp, err := c.doRequest("GET", path, queryString, nil)
	if err != nil {
		return TradeFillsResult{}, url, "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return TradeFillsResult{}, url, "", fmt.Errorf("failed to read response body: %w", err)
	}

	rawResponse := string(bodyBytes)

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errResp); err == nil {
			return TradeFillsResult{}, url, rawResponse, &APIError{
				Code:    errResp.Code,
				Message: errResp.Message,
				Status:  resp.StatusCode,
			}
		}
		return TradeFillsResult{}, url, rawResponse, &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       rawResponse,
		}
	}

	// Parse response with totals and nextFlag
	var result TradeFillsResult
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return TradeFillsResult{}, url, rawResponse, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, url, rawResponse, nil
}

// GetTradeFills retrieves filled orders (trade details) for a symbol
// Returns TradeFillsResult containing fills, totals, and nextFlag.
// Note: API does not support traditional pagination. If nextFlag is true,
// use time range splitting (startTime/endTime) to fetch all data.
func (c *Client) GetTradeFills(symbol string, opts *TradeFillsOptions) (TradeFillsResult, error) {
	path := "/capi/v2/order/fills"

	// Default limit to 100 for maximum results
	limit := 100
	if opts != nil && opts.Limit > 0 {
		limit = opts.Limit
		if limit > 100 {
			limit = 100
		}
	}

	queryString := fmt.Sprintf("?symbol=%s&limit=%d", symbol, limit)
	if opts != nil {
		if opts.OrderId != "" {
			queryString += fmt.Sprintf("&orderId=%s", opts.OrderId)
		}
		if opts.StartTime > 0 {
			queryString += fmt.Sprintf("&startTime=%d", opts.StartTime)
		}
		if opts.EndTime > 0 {
			queryString += fmt.Sprintf("&endTime=%d", opts.EndTime)
		}
	}
	resp, err := c.doRequest("GET", path, queryString, nil)
	if err != nil {
		return TradeFillsResult{}, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return TradeFillsResult{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(bodyBytes, &errResp); err == nil {
			return TradeFillsResult{}, &APIError{
				Code:    errResp.Code,
				Message: errResp.Message,
				Status:  resp.StatusCode,
			}
		}
		return TradeFillsResult{}, &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(bodyBytes),
		}
	}

	// Parse response with totals and nextFlag
	var result TradeFillsResult
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return TradeFillsResult{}, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}
