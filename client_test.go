package weexgo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTradeFills(t *testing.T) {
	// Mock API response
	mockResponse := map[string]interface{}{
		"list": []map[string]interface{}{
			{
				"tradeId":      12345,
				"orderId":      67890,
				"symbol":       "cmt_btcusdt",
				"marginMode":   "SHARED",
				"positionSide": "LONG",
				"orderSide":    "BUY",
				"fillSize":     "0.001",
				"fillValue":    "50.00",
				"fillFee":      "0.05",
				"realizePnl":   "0",
				"direction":    "TAKER",
				"createdTime":  1705900800000,
			},
			{
				"tradeId":      12346,
				"orderId":      67891,
				"symbol":       "cmt_btcusdt",
				"marginMode":   "SHARED",
				"positionSide": "SHORT",
				"orderSide":    "SELL",
				"fillSize":     "0.002",
				"fillValue":    "100.00",
				"fillFee":      "0.10",
				"realizePnl":   "1.50",
				"direction":    "MAKER",
				"createdTime":  1705900900000,
			},
		},
		"totals":   150,
		"nextFlag": true,
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request path and method
		if r.URL.Path != "/capi/v2/order/fills" {
			t.Errorf("Expected path /capi/v2/order/fills, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		// Verify query parameters
		query := r.URL.Query()
		if query.Get("symbol") != "cmt_btcusdt" {
			t.Errorf("Expected symbol=cmt_btcusdt, got %s", query.Get("symbol"))
		}
		if query.Get("limit") != "100" {
			t.Errorf("Expected limit=100, got %s", query.Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client with mock server URL
	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	// Call GetTradeFills
	result, err := client.GetTradeFills("cmt_btcusdt", nil)
	if err != nil {
		t.Fatalf("GetTradeFills failed: %v", err)
	}

	// Verify results
	if len(result.Fills) != 2 {
		t.Errorf("Expected 2 fills, got %d", len(result.Fills))
	}
	if result.Totals != 150 {
		t.Errorf("Expected totals=150, got %d", result.Totals)
	}
	if !result.NextFlag {
		t.Error("Expected nextFlag=true, got false")
	}

	// Verify first fill
	if result.Fills[0].TradeID != 12345 {
		t.Errorf("Expected tradeId=12345, got %d", result.Fills[0].TradeID)
	}
	if result.Fills[0].Symbol != "cmt_btcusdt" {
		t.Errorf("Expected symbol=cmt_btcusdt, got %s", result.Fills[0].Symbol)
	}
}

func TestGetTradeFillsWithOptions(t *testing.T) {
	mockResponse := map[string]interface{}{
		"list":     []map[string]interface{}{},
		"totals":   0,
		"nextFlag": false,
	}

	var capturedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	opts := &TradeFillsOptions{
		StartTime: 1705900000000,
		EndTime:   1705990000000,
		Limit:     50,
	}

	_, err := client.GetTradeFills("cmt_btcusdt", opts)
	if err != nil {
		t.Fatalf("GetTradeFills failed: %v", err)
	}

	// Verify query string contains expected parameters
	expectedParams := []string{"symbol=cmt_btcusdt", "limit=50", "startTime=1705900000000", "endTime=1705990000000"}
	for _, param := range expectedParams {
		if !containsParam(capturedQuery, param) {
			t.Errorf("Expected query to contain %s, got %s", param, capturedQuery)
		}
	}
}

func TestGetTradeFillsLimitCap(t *testing.T) {
	mockResponse := map[string]interface{}{
		"list":     []map[string]interface{}{},
		"totals":   0,
		"nextFlag": false,
	}

	var capturedLimit string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedLimit = r.URL.Query().Get("limit")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := &Client{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	// Try to set limit > 100, should be capped to 100
	opts := &TradeFillsOptions{Limit: 200}
	_, err := client.GetTradeFills("cmt_btcusdt", opts)
	if err != nil {
		t.Fatalf("GetTradeFills failed: %v", err)
	}

	if capturedLimit != "100" {
		t.Errorf("Expected limit to be capped at 100, got %s", capturedLimit)
	}
}

func TestTradeFillsResultNextFlag(t *testing.T) {
	// Test case: nextFlag indicates more data available
	result := TradeFillsResult{
		Fills:    make(TradeFills, 100),
		Totals:   250,
		NextFlag: true,
	}

	if !result.NextFlag {
		t.Error("NextFlag should be true when there are more records")
	}
	if result.Totals <= 100 {
		t.Error("Totals should be greater than limit when NextFlag is true")
	}
}

func containsParam(query, param string) bool {
	return len(query) > 0 && (query == param ||
		len(query) > len(param) && (query[:len(param)+1] == param+"&" ||
			query[len(query)-len(param)-1:] == "&"+param ||
			contains(query, "&"+param+"&")))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
