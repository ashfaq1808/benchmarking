package core

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type BenchmarkResult struct {
	Type      string `json:"type"`
	DelayMs   int64  `json:"delay_ms"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	Timestamp string `json:"timestamp"`
}

var logMutex sync.Mutex

func logResult(testType string, delay int64, success bool, errMsg string) {
	entry := BenchmarkResult{
		Type:      testType,
		DelayMs:   delay,
		Success:   success,
		Error:     errMsg,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	logMutex.Lock()
	defer logMutex.Unlock()
	f, _ := os.OpenFile("results.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	json.NewEncoder(f).Encode(entry)
}
