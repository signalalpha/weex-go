package weexgo

import "fmt"

// AccountBalance represents account balance information
// Response format from /capi/v2/account/assets:
//
//	[
//	  {
//	    "coinName": "USDT",
//	    "available": "5413.06877369",
//	    "equity": "5696.49288823",
//	    "frozen": "81.28240000",
//	    "unrealizePnl": "-34.55300000"
//	  }
//	]
type AccountBalance struct {
	CoinName     string `json:"coinName"`
	Available    string `json:"available"`
	Equity       string `json:"equity"`
	Frozen       string `json:"frozen"`
	UnrealizePnl string `json:"unrealizePnl"`
}

// AccountAssets represents account assets information
// It's a slice of AccountBalance (direct array response from /capi/v2/account/assets)
type AccountAssets []AccountBalance

// Ticker represents market ticker data
type Ticker struct {
	Symbol    string `json:"symbol"`
	LastPrice string `json:"lastPrice"`
	High24h   string `json:"high24h"`
	Low24h    string `json:"low24h"`
	Volume24h string `json:"volume24h"`
	Change24h string `json:"change24h"`
}

// OrderSide represents order side (buy or sell)
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// OrderType represents order type
type OrderType string

const (
	OrderTypeMarket OrderType = "market"
	OrderTypeLimit  OrderType = "limit"
)

// CreateOrderRequest represents a request to create an order
type CreateOrderRequest struct {
	Symbol    string    `json:"symbol"`
	Side      OrderSide `json:"side"`
	OrderType OrderType `json:"orderType"`
	Quantity  string    `json:"quantity"`
	Price     string    `json:"price,omitempty"` // Required for limit orders
}

// Order represents an order
// Supports both API response formats: order_id (snake_case) and orderId (camelCase)
type Order struct {
	OrderID    string    `json:"order_id"` // API returns order_id (snake_case)
	OrderIDAlt string    `json:"orderId"`  // Alternative format (camelCase)
	Symbol     string    `json:"symbol"`
	Side       OrderSide `json:"side"`
	OrderType  OrderType `json:"orderType"`
	Quantity   string    `json:"quantity"`
	Price      string    `json:"price"`
	Status     string    `json:"status"`
	CreateTime int64     `json:"createTime"`
}

// GetOrderID returns the order ID from either field
func (o *Order) GetOrderID() string {
	if o.OrderID != "" {
		return o.OrderID
	}
	return o.OrderIDAlt
}

// APIResponse represents a generic API response
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// TradeFill represents a filled order (trade detail)
// Response format from /capi/v2/order/fills:
// [
//
//	{
//	  "orderId": "123456",
//	  "symbol": "cmt_btcusdt",
//	  "side": "1",  // "1"=buy, "2"=sell
//	  "price": "50000.00",
//	  "size": "0.001",
//	  "fee": "0.05",
//	  "feeCoin": "USDT",
//	  "tradeTime": 1234567890000,
//	  "tradeId": "trade_123"
//	}
//
// ]
type TradeFill struct {
	TradeID              int64  `json:"tradeId"`             // 成交ID
	OrderID              int64  `json:"orderId"`             // 订单ID
	Symbol               string `json:"symbol"`              // 交易对
	MarginMode           string `json:"marginMode"`          // 保证金模式 SHARED/ISOLATED
	SeparatedMode        string `json:"separatedMode"`       // 分离模式
	PositionSide         string `json:"positionSide"`        // 仓位方向 LONG/SHORT
	OrderSide            string `json:"orderSide"`           // 订单方向 BUY/SELL
	FillSize             string `json:"fillSize"`            // 成交数量
	FillValue            string `json:"fillValue"`           // 成交价值
	FillFee              string `json:"fillFee"`             // 手续费
	LiquidateFee         string `json:"liquidateFee"`        // 强平手续费
	RealizePnl           string `json:"realizePnl"`          // 已实现盈亏
	Direction            string `json:"direction"`           // 方向 MAKER/TAKER
	LiquidateType        string `json:"liquidateType"`       // 强平类型
	LegacyOrderDirection string `json:"legacyOrdeDirection"` // 旧订单方向
	CreatedTime          int64  `json:"createdTime"`         // 成交时间（毫秒）
	ContractVal          string `json:"contractVal"`         // 合约面值
}

// GetOrderID returns the order ID as string
func (t *TradeFill) GetOrderID() string {
	return fmt.Sprintf("%d", t.OrderID)
}

// GetTradeID returns the trade ID as string
func (t *TradeFill) GetTradeID() string {
	return fmt.Sprintf("%d", t.TradeID)
}

// TradeFills represents a list of trade fills
type TradeFills []TradeFill
