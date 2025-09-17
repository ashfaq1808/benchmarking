package workload

import (
	"cassandra-benchmark/config"
	"cassandra-benchmark/result"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gocql/gocql"
)

var (
	writtenIDs []string
	idMux      sync.Mutex
)

func addWrittenID(id string) {
	idMux.Lock()
	defer idMux.Unlock()
	writtenIDs = append(writtenIDs, id)
}

func getRandomWrittenID() string {
	idMux.Lock()
	defer idMux.Unlock()
	if len(writtenIDs) == 0 {
		return ""
	}
	return writtenIDs[rand.Intn(len(writtenIDs))]
}

func RunBenchmark(cfg *config.Config, sessions []*gocql.Session) error {
	table := cfg.Cassandra.Table

	// Table already exists with schema: id uuid PRIMARY KEY, category text, data text, timestamp timestamp, value bigint
	// No need to create table as it already exists

	var wg sync.WaitGroup
	startTime := time.Now()
	warmupCutoff := startTime.Add(time.Duration(cfg.Benchmark.WarmupSeconds) * time.Second)
	endTime := warmupCutoff.Add(time.Duration(cfg.Benchmark.DurationSeconds) * time.Second)

	if cfg.Benchmark.Mode == "open-loop" && cfg.Benchmark.RequestsPerSecond > 0 {
		var currentRate int
		var ticker *time.Ticker
		
		if cfg.Benchmark.RatePattern.Enabled {
			currentRate = cfg.Benchmark.RatePattern.MinRate
		} else {
			currentRate = cfg.Benchmark.RequestsPerSecond
		}
		
		ticker = time.NewTicker(time.Second / time.Duration(currentRate))
		defer ticker.Stop()
		
		var phaseStart time.Time
		var inPeakPhase bool
		var lastRateChange time.Time
		
		if cfg.Benchmark.RatePattern.Enabled {
			if cfg.Benchmark.RatePattern.Mode == "cycles" {
				phaseStart = time.Now()
				inPeakPhase = false
			} else if cfg.Benchmark.RatePattern.Mode == "random" {
				lastRateChange = time.Now()
			}
		}

		for time.Now().Before(endTime) {
			if cfg.Benchmark.RatePattern.Enabled {
				now := time.Now()
				
				if cfg.Benchmark.RatePattern.Mode == "cycles" {
					phaseDuration := now.Sub(phaseStart)
					
					var shouldSwitchPhase bool
					var newRate int
					
					if inPeakPhase {
						shouldSwitchPhase = phaseDuration >= time.Duration(cfg.Benchmark.RatePattern.PeakDuration)*time.Second
						newRate = cfg.Benchmark.RatePattern.MinRate
					} else {
						shouldSwitchPhase = phaseDuration >= time.Duration(cfg.Benchmark.RatePattern.ValleyDuration)*time.Second
						newRate = cfg.Benchmark.RatePattern.MaxRate
					}
					
					if shouldSwitchPhase {
						inPeakPhase = !inPeakPhase
						phaseStart = now
						currentRate = newRate
						ticker.Stop()
						ticker = time.NewTicker(time.Second / time.Duration(currentRate))
					}
				} else if cfg.Benchmark.RatePattern.Mode == "random" {
					timeSinceLastChange := now.Sub(lastRateChange)
					
					if timeSinceLastChange >= time.Duration(cfg.Benchmark.RatePattern.ChangeInterval*1000)*time.Millisecond {
						rateRange := cfg.Benchmark.RatePattern.MaxRate - cfg.Benchmark.RatePattern.MinRate
						newRate := cfg.Benchmark.RatePattern.MinRate + rand.Intn(rateRange+1)
						
						if newRate != currentRate {
							currentRate = newRate
							ticker.Stop()
							ticker = time.NewTicker(time.Second / time.Duration(currentRate))
						}
						lastRateChange = now
					}
				}
			}
			
			<-ticker.C
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				nodeIndex := rand.Intn(len(sessions))
				session := sessions[nodeIndex]

				if rand.Float64() < cfg.Benchmark.WriteRatio {
					uuid := gocql.TimeUUID()
					empRaw := GenerateEmployee(rand.Int())
					
					emp := result.Employee{
						ID:        uuid.String(),
						Category:  empRaw.Category,
						Data:      empRaw.Data,
						Timestamp: empRaw.Timestamp,
						Value:     empRaw.Value,
					}

					start := time.Now()
					err := session.Query(
						fmt.Sprintf(`INSERT INTO %s (id, category, data, timestamp, value) VALUES (?, ?, ?, ?, ?)`, table),
						uuid, emp.Category, emp.Data, time.Now(), emp.Value,
					).Exec()
					duration := time.Since(start)

					if time.Now().After(warmupCutoff) {
						result.LogWrite(workerID, nodeIndex, emp, duration, err)
					}
					if err == nil {
						addWrittenID(emp.ID)
					}
				} else {
					id := getRandomWrittenID()
					if id == "" {
						return
					}
					start := time.Now()
					var category, data string
					var timestamp time.Time
					var value int64
					
					uuid, parseErr := gocql.ParseUUID(id)
					if parseErr != nil {
						return
					}
					
					err := session.Query(
						fmt.Sprintf(`SELECT category, data, timestamp, value FROM %s WHERE id = ?`, table),
						uuid,
					).Scan(&category, &data, &timestamp, &value)
					duration := time.Since(start)

					if time.Now().After(warmupCutoff) {
						result.LogRead(workerID, nodeIndex, id, category, data, timestamp.Format(time.RFC3339), value, duration, err)
					}
				}
			}(rand.Intn(cfg.Benchmark.Concurrency))
		}
	} else {
		for i := 0; i < cfg.Benchmark.Concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for time.Now().Before(endTime) {
					nodeIndex := rand.Intn(len(sessions))
					session := sessions[nodeIndex]

					if rand.Float64() < cfg.Benchmark.WriteRatio {
						uuid := gocql.TimeUUID()
						empRaw := GenerateEmployee(rand.Int())
						
						emp := result.Employee{
							ID:        uuid.String(),
							Category:  empRaw.Category,
							Data:      empRaw.Data,
							Timestamp: empRaw.Timestamp,
							Value:     empRaw.Value,
						}

						start := time.Now()
						err := session.Query(
							fmt.Sprintf(`INSERT INTO %s (id, category, data, timestamp, value) VALUES (?, ?, ?, ?, ?)`, table),
							uuid, emp.Category, emp.Data, time.Now(), emp.Value,
						).Exec()
						duration := time.Since(start)

						if time.Now().After(warmupCutoff) {
							result.LogWrite(workerID, nodeIndex, emp, duration, err)
						}
						if err == nil {
							addWrittenID(emp.ID)
						}
					} else {
						id := getRandomWrittenID()
						if id == "" {
							continue
						}
						start := time.Now()
						var category, data string
						var timestamp time.Time
						var value int64
						
						uuid, parseErr := gocql.ParseUUID(id)
						if parseErr != nil {
							continue
						}
						
						err := session.Query(
							fmt.Sprintf(`SELECT category, data, timestamp, value FROM %s WHERE id = ?`, table),
							uuid,
						).Scan(&category, &data, &timestamp, &value)
						duration := time.Since(start)

						if time.Now().After(warmupCutoff) {
							result.LogRead(workerID, nodeIndex, id, category, data, timestamp.Format(time.RFC3339), value, duration, err)
						}
					}

					if cfg.Benchmark.Mode == "closed-loop" {
						time.Sleep(10 * time.Millisecond)
					}
				}
			}(i)
		}
	}

	wg.Wait()
	return nil
}
