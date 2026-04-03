# USDC Payment API

A REST API for querying USDC data on Ethereum Mainnet, built with Go and Gin.

## Endpoints

- `GET /balance/:address` — Query USDC balance of any address
- `GET /transfers/:address` — Get transfer history (sent + received) for any address
- `GET /tx/:hash` — Get transaction status by hash

## Tech Stack

- Go
- Gin (HTTP framework)
- go-ethereum
- Infura RPC

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

## Example
```bash
# Query balance
curl http://localhost:8080/balance/0x55FE002aefF02F77364de339a1292923A15844B8

# Query transfer history
curl http://localhost:8080/transfers/0x55FE002aefF02F77364de339a1292923A15844B8

# Query transaction status
curl http://localhost:8080/tx/0x686306f2b9ab5d6308b92619dbaa5ae005a259fed05cd3a12c73f99401d4542e
```