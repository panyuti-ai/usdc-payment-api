# USDC Payment API

A REST API for querying USDC data on Ethereum Mainnet, built with Go and Gin.

Supports querying balance, transfer history (sent + received), and transaction status for any Ethereum address.

## Demo

Open `http://localhost:8080` after running to use the web dashboard.

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/balance/:address` | Query USDC balance of any address |
| GET | `/transfers/:address` | Get transfer history (sent + received) for any address within the last 1000 blocks |
| GET | `/tx/:hash` | Get transaction status by hash |

## Example Requests
```bash
# Query Circle Treasury balance
curl http://localhost:8080/balance/0x55FE002aefF02F77364de339a1292923A15844B8

# Query transfer history
curl http://localhost:8080/transfers/0x55FE002aefF02F77364de339a1292923A15844B8

# Query transaction status
curl http://localhost:8080/tx/0x686306f2b9ab5d6308b92619dbaa5ae005a259fed05cd3a12c73f99401d4542e
```

## Example Response
```json
{
  "address": "0x55FE002aefF02F77364de339a1292923A15844B8",
  "balance": 38926.24,
  "symbol": "USDC"
}
```
```json
{
  "address": "0x55FE002aefF02F77364de339a1292923A15844B8",
  "count": 118,
  "transfers": [
    {
      "block_number": 24799808,
      "tx_hash": "0xb75dd7...",
      "from": "0x55FE002aefF02F77364de339a1292923A15844B8",
      "to": "0x4735880F32cb20E53a37bD04AA5C0EeBa3Ceb637",
      "value_usdc": 354252.30
    }
  ]
}
```

## Tech Stack

- Go
- Gin (HTTP framework)
- go-ethereum
- Infura RPC
- Vanilla HTML/CSS (no framework)

## How to Run

1. Clone the repo
```bash
git clone https://github.com/panyuti-ai/usdc-payment-api
cd usdc-payment-api
```

2. Install dependencies
```bash
go mod tidy
```

3. Get a free Infura API key at https://app.infura.io

4. Run
```bash
RPC_URL="https://mainnet.infura.io/v3/YOUR_API_KEY" go run main.go
```

5. Open http://localhost:8080