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

func RunBenchmark(cfg *config.Config, session *gocql.Session) error {
	table := cfg.Cassandra.Table
	createTable := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id text PRIMARY KEY,
		name text,
		dept text,
		salary int
	);`, table)
	if err := session.Query(createTable).Exec(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	endTime := time.Now().Add(time.Duration(cfg.Benchmark.DurationSeconds) * time.Second)

	if cfg.Benchmark.Mode == "open-loop" && cfg.Benchmark.RequestsPerSecond > 0 {
		// open-loop with request rate limiting
		ticker := time.NewTicker(time.Second / time.Duration(cfg.Benchmark.RequestsPerSecond))
		defer ticker.Stop()

		for time.Now().Before(endTime) {
			<-ticker.C
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				id := rand.Int()
				empRaw := GenerateEmployee(id)
				emp := result.Employee{
					ID:     empRaw.ID,
					Name:   empRaw.Name,
					Dept:   empRaw.Dept,
					Salary: empRaw.Salary,
				}
				if rand.Float64() < cfg.Benchmark.WriteRatio {
					start := time.Now()
					err := session.Query(
						fmt.Sprintf(`INSERT INTO %s (id, name, dept, salary) VALUES (?, ?, ?, ?)`, table),
						emp.ID, emp.Name, emp.Dept, emp.Salary,
					).Exec()
					duration := time.Since(start)
					result.LogWrite(workerID, emp, duration, err)
				} else {
					start := time.Now()
					var name, dept string
					var salary int
					err := session.Query(
						fmt.Sprintf(`SELECT name, dept, salary FROM %s WHERE id = ?`, table),
						emp.ID,
					).Scan(&name, &dept, &salary)
					duration := time.Since(start)
					result.LogRead(workerID, emp.ID, name, dept, salary, duration, err)
				}
			}(rand.Intn(cfg.Benchmark.Concurrency))
		}
	} else {
		// closed-loop or unthrottled open-loop
		for i := 0; i < cfg.Benchmark.Concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				id := 0
				for time.Now().Before(endTime) {
					id++
					empRaw := GenerateEmployee(id)
					emp := result.Employee{
						ID:     empRaw.ID,
						Name:   empRaw.Name,
						Dept:   empRaw.Dept,
						Salary: empRaw.Salary,
					}
					if rand.Float64() < cfg.Benchmark.WriteRatio {
						start := time.Now()
						err := session.Query(
							fmt.Sprintf(`INSERT INTO %s (id, name, dept, salary) VALUES (?, ?, ?, ?)`, table),
							emp.ID, emp.Name, emp.Dept, emp.Salary,
						).Exec()
						duration := time.Since(start)
						result.LogWrite(workerID, emp, duration, err)
					} else {
						start := time.Now()
						var name, dept string
						var salary int
						err := session.Query(
							fmt.Sprintf(`SELECT name, dept, salary FROM %s WHERE id = ?`, table),
							emp.ID,
						).Scan(&name, &dept, &salary)
						duration := time.Since(start)
						result.LogRead(workerID, emp.ID, name, dept, salary, duration, err)
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
