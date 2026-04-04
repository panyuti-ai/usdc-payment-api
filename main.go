package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
)

const USDC_ADDRESS = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"

// Transfer event signature hash
const TRANSFER_TOPIC = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

var client *ethclient.Client

// 查詢 USDC 餘額
func getBalance(c *gin.Context) {
	address := c.Param("address")

	if !common.IsHexAddress(address) {
		c.JSON(400, gin.H{"error": "invalid address"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	contractAddress := common.HexToAddress(USDC_ADDRESS)

	// balanceOf(address) 的 function selector + 地址
	data := fmt.Sprintf("0x70a08231000000000000000000000000%s", address[2:])

	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: common.FromHex(data),
	}

	result, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	balance := new(big.Int).SetBytes(result)
	balanceFloat, _ := new(big.Float).Quo(
		new(big.Float).SetInt(balance),
		big.NewFloat(1e6),
	).Float64()

	c.JSON(200, gin.H{
		"address": address,
		"balance": balanceFloat,
		"symbol":  "USDC",
	})
}

// 查詢地址的轉帳歷史
func getTransfers(c *gin.Context) {
	address := c.Param("address")

	if !common.IsHexAddress(address) {
		c.JSON(400, gin.H{"error": "invalid address"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	latestBlock := header.Number.Uint64()
	fromBlock := latestBlock - 1000

	contractAddress := common.HexToAddress(USDC_ADDRESS)
	addr := common.HexToAddress(address)

	// 把地址轉成 32 bytes 的 topic 格式
	addrHash := common.BytesToHash(addr.Bytes())

	// 查詢這個地址作為發送方
	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()

	sentLogs, err := client.FilterLogs(ctx2, ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(latestBlock)),
		Addresses: []common.Address{contractAddress},
		Topics: [][]common.Hash{
			{common.HexToHash(TRANSFER_TOPIC)},
			{addrHash},
			nil,
		},
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 查詢這個地址作為接收方
	ctx3, cancel3 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel3()

	receivedLogs, err := client.FilterLogs(ctx3, ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(latestBlock)),
		Addresses: []common.Address{contractAddress},
		Topics: [][]common.Hash{
			{common.HexToHash(TRANSFER_TOPIC)},
			nil,
			{addrHash},
		},
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 合併發送和接收的結果
	logs := append(sentLogs, receivedLogs...)

	type Transfer struct {
		BlockNumber uint64  `json:"block_number"`
		TxHash      string  `json:"tx_hash"`
		From        string  `json:"from"`
		To          string  `json:"to"`
		ValueUsdc   float64 `json:"value_usdc"`
	}

	var transfers []Transfer
	for _, vLog := range logs {
		if len(vLog.Topics) < 3 {
			continue
		}
		value := new(big.Int).SetBytes(vLog.Data)
		valueFloat, _ := new(big.Float).Quo(
			new(big.Float).SetInt(value),
			big.NewFloat(1e6),
		).Float64()

		transfers = append(transfers, Transfer{
			BlockNumber: vLog.BlockNumber,
			TxHash:      vLog.TxHash.Hex(),
			From:        common.HexToAddress(vLog.Topics[1].Hex()).Hex(),
			To:          common.HexToAddress(vLog.Topics[2].Hex()).Hex(),
			ValueUsdc:   valueFloat,
		})
	}

	c.JSON(200, gin.H{
		"address":   address,
		"transfers": transfers,
		"count":     len(transfers),
	})
}

// 查詢交易狀態
func getTx(c *gin.Context) {
	hash := c.Param("hash")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, isPending, err := client.TransactionByHash(ctx, common.HexToHash(hash))
	if err != nil {
		c.JSON(404, gin.H{"error": "transaction not found"})
		return
	}

	if isPending {
		c.JSON(200, gin.H{
			"hash":   hash,
			"status": "pending",
		})
		return
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()

	receipt, err := client.TransactionReceipt(ctx2, common.HexToHash(hash))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	status := "success"
	if receipt.Status == types.ReceiptStatusFailed {
		status = "failed"
	}

	c.JSON(200, gin.H{
		"hash":     hash,
		"status":   status,
		"block":    receipt.BlockNumber,
		"gas_used": receipt.GasUsed,
		"to":       tx.To(),
	})
}

func main() {
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		log.Fatal("請設定 RPC_URL 環境變數")
	}

	var err error
	client, err = ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatal("連接失敗:", err)
	}
	fmt.Println("✅ 成功連接 Ethereum 節點")

	r := gin.Default()

	// 允許跨域請求，讓其他網站也能呼叫這個 API
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET")
		c.Next()
	})
	r.GET("/balance/:address", getBalance)
	r.GET("/transfers/:address", getTransfers)
	r.GET("/tx/:hash", getTx)
	r.GET("/", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.File("index.html")
	})

	fmt.Println("🚀 API 伺服器啟動：http://localhost:8080")
	r.Run(":8080")
}
