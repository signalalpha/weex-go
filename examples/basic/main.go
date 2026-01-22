package main

import (
	"fmt"
	"log"
	"os"

	weexgo "github.com/signalalpha/weex-go"
)

// This is a basic example demonstrating how to use the weex-go SDK
func main() {
	// Read API credentials from environment variables
	apiKey := os.Getenv("WEEX_API_KEY")
	secretKey := os.Getenv("WEEX_SECRET_KEY")
	passphrase := os.Getenv("WEEX_PASSPHRASE")

	// Validate that all required environment variables are set
	if apiKey == "" || secretKey == "" || passphrase == "" {
		log.Fatal("Missing required environment variables. Please set:\n" +
			"  - WEEX_API_KEY\n" +
			"  - WEEX_SECRET_KEY\n" +
			"  - WEEX_PASSPHRASE")
	}

	// Create a new WEEX API client using functional options
	client, err := weexgo.NewClient(
		weexgo.WithAPIKey(apiKey),
		weexgo.WithSecretKey(secretKey),
		weexgo.WithPassphrase(passphrase),
		// Optional: Set a custom base URL (default: "https://api-contract.weex.com")
		// weexgo.WithBaseURL("https://api-contract.weex.com"),
		// Optional: Set a custom timeout (default: 30 seconds)
		// weexgo.WithTimeout(30 * time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== WEEX Go SDK Basic Example ===")

	// Example 1: Get account assets
	fmt.Println("1. Getting account assets...")
	assets, err := client.GetAccountAssets()
	if err != nil {
		log.Printf("Error getting account assets: %v", err)
	} else {
		fmt.Printf("✅ Found %d asset(s):\n", len(assets))
		for _, asset := range assets {
			fmt.Printf("   - %s: Available=%s, Equity=%s, Frozen=%s, Unrealized PnL=%s\n",
				asset.CoinName, asset.Available, asset.Equity, asset.Frozen, asset.UnrealizePnl)
		}
	}

	// Example 2: Get ticker for a symbol
	fmt.Println("\n2. Getting ticker for cmt_btcusdt...")
	ticker, err := client.GetTicker("cmt_btcusdt")
	if err != nil {
		log.Printf("Error getting ticker: %v", err)
	} else {
		fmt.Printf("✅ Ticker information:\n")
		fmt.Printf("   - Symbol: %s\n", ticker.Symbol)
		fmt.Printf("   - Last Price: %s\n", ticker.LastPrice)
		fmt.Printf("   - 24h High: %s\n", ticker.High24h)
		fmt.Printf("   - 24h Low: %s\n", ticker.Low24h)
		fmt.Printf("   - 24h Volume: %s\n", ticker.Volume24h)
		fmt.Printf("   - 24h Change: %s\n", ticker.Change24h)
	}

	// Example 3: Set leverage (optional - uncomment to test)
	/*
		fmt.Println("\n3. Setting leverage to 1x (full margin mode)...")
		err = client.SetLeverage("cmt_btcusdt", 1, "1", "1")
		if err != nil {
			log.Printf("Error setting leverage: %v", err)
		} else {
			fmt.Println("✅ Leverage set successfully")
		}
	*/

	// Example 4: Create an order (optional - uncomment to test)
	// WARNING: This will place a real order! Use with caution.
	/*
		fmt.Println("\n4. Creating a market buy order for 0.0001 BTC...")
		order, err := client.CreateOrder(&weexgo.CreateOrderRequest{
			Symbol:    "cmt_btcusdt",
			Side:      weexgo.OrderSideBuy,
			OrderType: weexgo.OrderTypeMarket,
			Quantity:  "0.0001",
		})
		if err != nil {
			log.Printf("Error creating order: %v", err)
		} else {
			fmt.Printf("✅ Order created successfully:\n")
			fmt.Printf("   - Order ID: %s\n", order.OrderID)
			fmt.Printf("   - Symbol: %s\n", order.Symbol)
			fmt.Printf("   - Side: %s\n", order.Side)
			fmt.Printf("   - Type: %s\n", order.OrderType)
			fmt.Printf("   - Quantity: %s\n", order.Quantity)
			fmt.Printf("   - Status: %s\n", order.Status)
		}
	*/

	fmt.Println("\n=== Example completed ===")
}
