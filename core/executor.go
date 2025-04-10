package core

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var writtenKeys []string
var keyMutex sync.Mutex

func RunOpenLoopBenchmark(e *CassandraEngine, cfg *Config) {
	totalOps := cfg.TotalReads + cfg.TotalWrites
	opsPerWorker := totalOps / cfg.Concurrency
	var wg sync.WaitGroup

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reads, writes := 0, 0
			for j := 0; j < opsPerWorker; j++ {
				start := time.Now()
				if writes < cfg.TotalWrites && (reads >= cfg.TotalReads || rand.Intn(2) == 0) {
					// Perform write
					success := performWrite(e)
					latency := time.Since(start).Milliseconds()
					logResult(BenchmarkResult{
						Type:             "write",
						DelayMs:          latency,
						Success:          success,
						Timestamp:        time.Now().Format(time.RFC3339),
						ConcurrencyLevel: cfg.Concurrency,
						Throughput:       1,
					})
					if success {
						writes++
					}
				} else if reads < cfg.TotalReads {
					// Perform read
					success := performRead(e)
					latency := time.Since(start).Milliseconds()
					logResult(BenchmarkResult{
						Type:             "read",
						DelayMs:          latency,
						Success:          success,
						Timestamp:        time.Now().Format(time.RFC3339),
						ConcurrencyLevel: cfg.Concurrency,
						Throughput:       1,
					})
					if success {
						reads++
					}
				}
			}
		}()
	}
	wg.Wait()
}
func RunClosedLoopBenchmark(e *CassandraEngine, cfg *Config) {
	targetRate := 100.0 // Placeholder for target throughput
	lastTime := time.Now()
	totalRequests := 0

	for {
		currentTime := time.Now()
		elapsedTime := currentTime.Sub(lastTime).Seconds()

		if elapsedTime >= 1 {
			throughput := float64(totalRequests) / elapsedTime
			adjustRate := calculateAdjustment(throughput, targetRate)
			performRequests(e, adjustRate)
			lastTime = currentTime
			totalRequests = 0
		}
	}
}

func RunLoadTest(e *CassandraEngine, cfg *Config) {
	fmt.Println("Starting load test...")
	maxConcurrency := cfg.MaxConcurrency
	failureThreshold := cfg.FailureThreshold

	for concurrency := 1; concurrency <= maxConcurrency; concurrency++ {
		fmt.Printf("Running with %d concurrent requests...\n", concurrency)
		var wg sync.WaitGroup
		failureCount := 0

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if !performRandomReadWrite(e) {
					failureCount++
				}
			}()
		}
		wg.Wait()

		failureRate := float64(failureCount) / float64(concurrency)
		if failureRate > failureThreshold {
			fmt.Printf("Server started failing at concurrency %d with failure rate %.2f%%\n", concurrency, failureRate*100)
			break
		}
	}
}

func performWrite(e *CassandraEngine) bool {
	id := fmt.Sprintf("w_%d", time.Now().UnixNano())
	payload := "payload"
	session := e.GetRandomSession()
	var err error
	for retries := 0; retries < 3; retries++ {
		err = session.Query(fmt.Sprintf("INSERT INTO %s (id, data) VALUES (?, ?)", e.Config.Table), id, payload).Exec()
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second) // Retry after a short delay
	}

	if err != nil {
		logResult(BenchmarkResult{
			Type:      "write",
			Success:   false,
			DelayMs:   0,
			ErrorCode: err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return false
	}
	keyMutex.Lock()
	writtenKeys = append(writtenKeys, id)
	keyMutex.Unlock()
	return true
}

func performRead(e *CassandraEngine) bool {
	keyMutex.Lock()
	if len(writtenKeys) == 0 {
		keyMutex.Unlock()
		return false
	}
	id := writtenKeys[rand.Intn(len(writtenKeys))]
	keyMutex.Unlock()

	var data string
	session := e.GetRandomSession()
	err := session.Query(fmt.Sprintf("SELECT data FROM %s WHERE id = ?", e.Config.Table), id).Scan(&data)
	if err != nil {
		logResult(BenchmarkResult{
			Type:      "read",
			Success:   false,
			DelayMs:   0,
			ErrorCode: err.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return false
	}
	return true
}

func performRandomReadWrite(e *CassandraEngine) bool {
	id := fmt.Sprintf("w_%d", time.Now().UnixNano())
	payload := "payload"
	session := e.GetRandomSession()
	err := session.Query(fmt.Sprintf("INSERT INTO %s (id, data) VALUES (?, ?)", e.Config.Table), id, payload).Exec()
	return err == nil
}

func calculateAdjustment(currentThroughput, targetRate float64) float64 {
	adjustmentFactor := targetRate / currentThroughput
	if adjustmentFactor > 1 {
		return 1 + adjustmentFactor*0.1
	} else {
		return 1 - adjustmentFactor*0.1
	}
}

// New function to perform requests dynamically based on the adjusted rate
func performRequests(e *CassandraEngine, rateAdjustment float64) {
	numRequests := int(rateAdjustment * float64(100))

	for i := 0; i < numRequests; i++ {
		if rand.Intn(2) == 0 {
			performWrite(e)
		} else {
			performRead(e)
		}
	}
}

func ValidateReads(e *CassandraEngine, sampleSize int) {
	fmt.Println("Validating read correctness...")

	// Pick random keys from the written keys to validate
	keyMutex.Lock()
	sampleKeys := getSampleKeys(sampleSize)
	keyMutex.Unlock()

	// Validate each key
	for _, key := range sampleKeys {
		var data string
		session := e.GetRandomSession()
		err := session.Query(fmt.Sprintf("SELECT data FROM %s WHERE id = ?", e.Config.Table), key).Scan(&data)
		if err != nil {
			fmt.Printf("Read validation failed for key: %s\n", key)
		} else {
			fmt.Printf("Read validation success for key: %s | data: %s\n", key, data)
		}
	}
}

// Helper function to get a sample of keys for validation
func getSampleKeys(sampleSize int) []string {
	sampleKeys := make([]string, 0, sampleSize)
	for i := 0; i < sampleSize; i++ {
		if len(writtenKeys) > 0 {
			key := writtenKeys[rand.Intn(len(writtenKeys))]
			sampleKeys = append(sampleKeys, key)
		}
	}
	return sampleKeys
}
