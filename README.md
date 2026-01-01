# weex-go

WEEX 交易所官方 Go SDK

## 特性

- ✅ 完整的 REST API 支持
- ✅ HMAC-SHA256 签名认证
- ✅ 类型安全的请求/响应模型
- ✅ 灵活的配置选项（Functional Options 模式）
- ✅ 完善的错误处理

## 安装

```bash
go get github.com/signalalpha/weex-go
```

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "os"
    "github.com/signalalpha/weex-go"
)

func main() {
    // 创建客户端
    client, err := weexgo.NewClient(
        weexgo.WithAPIKey(os.Getenv("WEEX_API_KEY")),
        weexgo.WithSecretKey(os.Getenv("WEEX_SECRET_KEY")),
        weexgo.WithPassphrase(os.Getenv("WEEX_PASSPHRASE")),
    )
    if err != nil {
        panic(err)
    }

    // 查询账户信息
    assets, err := client.GetAccountAssets()
    if err != nil {
        panic(err)
    }

    for _, asset := range assets {
        fmt.Printf("%s: 可用余额 %s, 权益 %s\n", 
            asset.CoinName, asset.Available, asset.Equity)
    }
}
```

### 配置选项

```go
client, err := weexgo.NewClient(
    weexgo.WithAPIKey("your_api_key"),
    weexgo.WithSecretKey("your_secret_key"),
    weexgo.WithPassphrase("your_passphrase"),
    weexgo.WithBaseURL("https://api-contract.weex.com"), // 可选，默认值
    weexgo.WithTimeout(30 * time.Second),                // 可选，默认 30s
)
```

### 环境变量

建议使用环境变量来管理敏感信息：

```bash
export WEEX_API_KEY="your_api_key"
export WEEX_SECRET_KEY="your_secret_key"
export WEEX_PASSPHRASE="your_passphrase"
```

## API 文档

### 账户相关

#### GetAccountAssets

查询账户资产信息。

```go
assets, err := client.GetAccountAssets()
if err != nil {
    // 处理错误
}
```

### 市场数据

#### GetTicker

获取交易对行情信息。

```go
ticker, err := client.GetTicker("cmt_btcusdt")
if err != nil {
    // 处理错误
}
```

### 交易相关

#### CreateOrder

创建订单。

```go
order, err := client.CreateOrder(&weexgo.CreateOrderRequest{
    Symbol:    "cmt_btcusdt",
    Side:      weexgo.OrderSideBuy,
    OrderType: weexgo.OrderTypeMarket,
    Quantity:  "10",
})
```

#### SetLeverage

设置杠杆。

```go
err := client.SetLeverage("cmt_btcusdt", 1, "1", "1")
```

## 错误处理

SDK 提供了完善的错误类型：

```go
assets, err := client.GetAccountAssets()
if err != nil {
    if apiErr, ok := err.(*weexgo.APIError); ok {
        fmt.Printf("API 错误: %d - %s\n", apiErr.Code, apiErr.Message)
    } else if httpErr, ok := err.(*weexgo.HTTPError); ok {
        fmt.Printf("HTTP 错误: %d\n", httpErr.StatusCode)
    } else {
        fmt.Printf("其他错误: %v\n", err)
    }
}
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 相关链接

- [WEEX 官网](https://www.weex.com)
- [WEEX API 文档](https://www.weex.com/zh-CN/news/detail/ai-wars-weex-alpha-awakens-weex-global-hackathon-api-test-process-guide-266016)
