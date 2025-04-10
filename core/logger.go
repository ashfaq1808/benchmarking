package core

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
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

var logMutex sync.Mutex

// func logResult(testType string, delay int64, success bool, errMsg string) {
// 	entry := BenchmarkResult{
// 		Type:      testType,
// 		DelayMs:   delay,
// 		Success:   success,
// 		Error:     errMsg,
// 		Timestamp: time.Now().Format(time.RFC3339),
// 	}

// 	logMutex.Lock()
// 	defer logMutex.Unlock()
// 	f, _ := os.OpenFile("results.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	defer f.Close()
// 	json.NewEncoder(f).Encode(entry)
// }

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
