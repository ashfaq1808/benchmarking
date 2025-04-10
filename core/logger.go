package core

import (
	"encoding/json"
	"fmt"
	"os"
)

type BenchmarkResult struct {
	Type             string `json:"type"`
	DelayMs          int64  `json:"delay_ms"`
	Success          bool   `json:"success"`
	Timestamp        string `json:"timestamp"`
	ConcurrencyLevel int    `json:"concurrency_level"`
	Throughput       int    `json:"throughput"`
	ErrorCode        string `json:"error_code,omitempty"`
}

// var logMutex sync.Mutex

func logResult(result BenchmarkResult) {
	file, err := os.OpenFile("results.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening results.json:", err)
		return
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	encoder.Encode(result)
}
