package entities

import (
	"time"
	"cassandra-benchmark/pkg/domain/entities/monitoring"
)

type BenchmarkConfig struct {
	DurationSeconds   int
	WarmupSeconds     int
	Concurrency       int
	ReadRatio         float64
	WriteRatio        float64
	Mode              string
	RequestsPerSecond int
	RatePattern       RatePatternConfig
	Monitoring        monitoring.MonitoringConfig
}

type RatePatternConfig struct {
	Enabled         bool
	Mode            string
	MinRate         int
	MaxRate         int
	PeakDuration    int
	ValleyDuration  int
	ChangeInterval  float64
}

type DatabaseConfig struct {
	Hosts    []string
	Keyspace string
	Table    string
}

type OperationResult struct {
	WorkerID  int
	NodeID    int
	Operation string
	Duration  time.Duration
	Success   bool
	Error     error
	Timestamp time.Time
}

type WriteResult struct {
	OperationResult
	Employee Employee
}

type ReadResult struct {
	OperationResult
	EmployeeID string
	Employee   *Employee
}