package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"cassandra-benchmark/pkg/domain/entities/monitoring"
	"cassandra-benchmark/pkg/domain/repositories"
)

type JsonMetricsRepository struct {
	filename   string
	file       *os.File
	encoder    *json.Encoder
	mutex      sync.Mutex
	metrics    []monitoring.SystemMetrics
	needsComma bool
}

func NewJsonMetricsRepository(filename string) repositories.MetricsRepository {
	return &JsonMetricsRepository{
		filename: filename,
		metrics:  make([]monitoring.SystemMetrics, 0),
	}
}

func (jmr *JsonMetricsRepository) SaveMetrics(ctx context.Context, metrics monitoring.SystemMetrics) error {
	jmr.mutex.Lock()
	defer jmr.mutex.Unlock()

	// Initialize file if not opened
	if jmr.file == nil {
		// Create fresh file (clears any existing content)
		file, err := os.Create(jmr.filename)
		if err != nil {
			return fmt.Errorf("failed to create metrics file: %w", err)
		}
		jmr.file = file
		jmr.needsComma = false

		// Write opening bracket for JSON array
		_, err = jmr.file.WriteString("[\n")
		if err != nil {
			return fmt.Errorf("failed to write opening bracket: %w", err)
		}
	}

	// Store in memory for queries
	jmr.metrics = append(jmr.metrics, metrics)

	// Write comma if needed (not for first entry)
	if jmr.needsComma {
		_, err := jmr.file.WriteString(",\n")
		if err != nil {
			return fmt.Errorf("failed to write comma: %w", err)
		}
	}

	// Write the metrics JSON
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	_, err = jmr.file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write metrics data: %w", err)
	}

	// Flush the data to disk immediately to prevent file locking issues
	err = jmr.file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	jmr.needsComma = true
	return nil
}

func (jmr *JsonMetricsRepository) GetMetrics(ctx context.Context, from, to time.Time) ([]monitoring.SystemMetrics, error) {
	jmr.mutex.Lock()
	defer jmr.mutex.Unlock()

	var filtered []monitoring.SystemMetrics
	for _, metric := range jmr.metrics {
		if metric.Timestamp.After(from) && metric.Timestamp.Before(to) {
			filtered = append(filtered, metric)
		}
	}

	return filtered, nil
}

func (jmr *JsonMetricsRepository) Close() error {
	jmr.mutex.Lock()
	defer jmr.mutex.Unlock()

	if jmr.file != nil {
		// Write closing bracket for JSON array
		_, err := jmr.file.WriteString("\n]")
		if err != nil {
			return fmt.Errorf("failed to write closing bracket: %w", err)
		}

		// Flush all data to disk before closing
		err = jmr.file.Sync()
		if err != nil {
			return fmt.Errorf("failed to sync file before closing: %w", err)
		}

		err = jmr.file.Close()
		if err != nil {
			return fmt.Errorf("failed to close file: %w", err)
		}

		jmr.file = nil
	}
	return nil
}

type MetricsCollectorImpl struct {
	systemMonitor  repositories.SystemMonitor
	dockerMonitor  repositories.DockerMonitor
	metricsRepo    repositories.MetricsRepository
	config         monitoring.MonitoringConfig
	stopChan       chan struct{}
	running        bool
	mutex          sync.Mutex
	currentMetrics *monitoring.SystemMetrics
}

func NewMetricsCollector(
	systemMonitor repositories.SystemMonitor,
	dockerMonitor repositories.DockerMonitor,
	metricsRepo repositories.MetricsRepository,
) repositories.MetricsCollector {
	return &MetricsCollectorImpl{
		systemMonitor: systemMonitor,
		dockerMonitor: dockerMonitor,
		metricsRepo:   metricsRepo,
		stopChan:      make(chan struct{}),
	}
}

func (mc *MetricsCollectorImpl) Start(ctx context.Context, config monitoring.MonitoringConfig) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if mc.running {
		return fmt.Errorf("metrics collector is already running")
	}

	mc.config = config
	mc.running = true

	go mc.collectLoop(ctx)
	return nil
}

func (mc *MetricsCollectorImpl) Stop() error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if !mc.running {
		return nil
	}

	close(mc.stopChan)
	mc.running = false

	// Close the metrics repository to properly close the JSON array
	if mc.metricsRepo != nil {
		return mc.metricsRepo.Close()
	}

	return nil
}

func (mc *MetricsCollectorImpl) GetCurrentMetrics() (*monitoring.SystemMetrics, error) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if mc.currentMetrics == nil {
		return nil, fmt.Errorf("no metrics collected yet")
	}

	return mc.currentMetrics, nil
}

func (mc *MetricsCollectorImpl) collectLoop(ctx context.Context) {
	ticker := time.NewTicker(mc.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-mc.stopChan:
			return
		case <-ticker.C:
			if err := mc.collectMetrics(ctx); err != nil {
				fmt.Printf("Error collecting metrics: %v\n", err)
			}
		}
	}
}

func (mc *MetricsCollectorImpl) collectMetrics(ctx context.Context) error {
	timestamp := time.Now()

	// Collect CPU metrics
	cpuMetrics, err := mc.systemMonitor.GetCPUMetrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to collect CPU metrics: %w", err)
	}

	// Collect memory metrics
	memoryMetrics, err := mc.systemMonitor.GetMemoryMetrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to collect memory metrics: %w", err)
	}

	// Collect network metrics
	networkMetrics, err := mc.systemMonitor.GetNetworkMetrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to collect network metrics: %w", err)
	}

	metrics := monitoring.SystemMetrics{
		Timestamp: timestamp,
		CPU:       *cpuMetrics,
		Memory:    *memoryMetrics,
		Network:   *networkMetrics,
	}

	// Collect GPU metrics if enabled
	if mc.config.EnableGPU {
		gpuMetrics, err := mc.systemMonitor.GetGPUMetrics(ctx)
		if err == nil && len(gpuMetrics.Devices) > 0 {
			metrics.GPU = gpuMetrics
		}
	}

	// Collect Docker metrics if enabled
	if mc.config.EnableDocker && mc.dockerMonitor != nil {
		var containerID string
		if mc.config.DockerContainerID != "" {
			containerID = mc.config.DockerContainerID
		} else {
			// Auto-detect running containers
			containers, err := mc.dockerMonitor.ListContainers(ctx)
			if err == nil && len(containers) > 0 {
				containerID = containers[0] // Use first running container
			}
		}

		if containerID != "" {
			dockerMetrics, err := mc.dockerMonitor.GetDockerMetrics(ctx, containerID)
			if err == nil {
				metrics.Docker = dockerMetrics
			}
		}
	}

	// Store current metrics
	mc.mutex.Lock()
	mc.currentMetrics = &metrics
	mc.mutex.Unlock()

	// Save to repository
	return mc.metricsRepo.SaveMetrics(ctx, metrics)
}
