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

	if err := sessions[0].Query(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id text PRIMARY KEY,
		name text,
		dept text, 
		salary int
	);`, table)).Exec(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	startTime := time.Now()
	warmupCutoff := startTime.Add(time.Duration(cfg.Benchmark.WarmupSeconds) * time.Second)
	endTime := startTime.Add(time.Duration(cfg.Benchmark.DurationSeconds) * time.Second)

	if cfg.Benchmark.Mode == "open-loop" && cfg.Benchmark.RequestsPerSecond > 0 {
		ticker := time.NewTicker(time.Second / time.Duration(cfg.Benchmark.RequestsPerSecond))
		defer ticker.Stop()

		for time.Now().Before(endTime) {
			<-ticker.C
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				nodeIndex := rand.Intn(len(sessions))
				session := sessions[nodeIndex]

				if rand.Float64() < cfg.Benchmark.WriteRatio {
					uuid := gocql.TimeUUID().String()
					empRaw := GenerateEmployee(rand.Int())
					empRaw.ID = uuid

					emp := result.Employee{
						ID:     empRaw.ID,
						Name:   empRaw.Name,
						Dept:   empRaw.Dept,
						Salary: empRaw.Salary,
					}

					start := time.Now()
					err := session.Query(
						fmt.Sprintf(`INSERT INTO %s (id, name, dept, salary) VALUES (?, ?, ?, ?)`, table),
						emp.ID, emp.Name, emp.Dept, emp.Salary,
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
					var name, dept string
					var salary int
					err := session.Query(
						fmt.Sprintf(`SELECT name, dept, salary FROM %s WHERE id = ?`, table),
						id,
					).Scan(&name, &dept, &salary)
					duration := time.Since(start)

					if time.Now().After(warmupCutoff) {
						result.LogRead(workerID, nodeIndex, id, name, dept, salary, duration, err)
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
						uuid := gocql.TimeUUID().String()
						empRaw := GenerateEmployee(rand.Int())
						empRaw.ID = uuid

						emp := result.Employee{
							ID:     empRaw.ID,
							Name:   empRaw.Name,
							Dept:   empRaw.Dept,
							Salary: empRaw.Salary,
						}

						start := time.Now()
						err := session.Query(
							fmt.Sprintf(`INSERT INTO %s (id, name, dept, salary) VALUES (?, ?, ?, ?)`, table),
							emp.ID, emp.Name, emp.Dept, emp.Salary,
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
						var name, dept string
						var salary int
						err := session.Query(
							fmt.Sprintf(`SELECT name, dept, salary FROM %s WHERE id = ?`, table),
							id,
						).Scan(&name, &dept, &salary)
						duration := time.Since(start)

						if time.Now().After(warmupCutoff) {
							result.LogRead(workerID, nodeIndex, id, name, dept, salary, duration, err)
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
