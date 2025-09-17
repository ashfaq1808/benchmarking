package repositories

import (
	"context"
	"time"
	"cassandra-benchmark/pkg/domain/entities/monitoring"
)

type SystemMonitor interface {
	GetCPUMetrics(ctx context.Context) (*monitoring.CPUMetrics, error)
	GetMemoryMetrics(ctx context.Context) (*monitoring.MemoryMetrics, error)
	GetNetworkMetrics(ctx context.Context) (*monitoring.NetworkMetrics, error)
	GetGPUMetrics(ctx context.Context) (*monitoring.GPUMetrics, error)
}

type DockerMonitor interface {
	GetDockerMetrics(ctx context.Context, containerID string) (*monitoring.DockerMetrics, error)
	ListContainers(ctx context.Context) ([]string, error)
	GetContainerName(ctx context.Context, containerID string) (string, error)
}

type MetricsRepository interface {
	SaveMetrics(ctx context.Context, metrics monitoring.SystemMetrics) error
	GetMetrics(ctx context.Context, from, to time.Time) ([]monitoring.SystemMetrics, error)
	Close() error
}

type MetricsCollector interface {
	Start(ctx context.Context, config monitoring.MonitoringConfig) error
	Stop() error
	GetCurrentMetrics() (*monitoring.SystemMetrics, error)
}