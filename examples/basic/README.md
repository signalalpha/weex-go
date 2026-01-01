# Basic Example

这是一个基本的 weex-go SDK 使用示例，演示了如何：

1. 创建 API 客户端
2. 查询账户资产
3. 获取交易对行情信息
4. 设置杠杆（可选）
5. 下单（可选，已注释）

## 运行示例

### 1. 设置环境变量

```bash
export WEEX_API_KEY="your_api_key"
export WEEX_SECRET_KEY="your_secret_key"
export WEEX_PASSPHRASE="your_passphrase"
```

### 2. 运行示例

```bash
go run main.go
```

## 注意事项

- 示例中的下单功能已注释，如需测试下单，请先取消注释相关代码
- 下单功能会创建真实订单，请谨慎使用
- 建议在测试网环境中先进行测试

