package monitoring

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"cassandra-benchmark/pkg/domain/entities/monitoring"
	"cassandra-benchmark/pkg/domain/repositories"
)

type DockerMonitorImpl struct {
	lastCPUTime     time.Time
	lastCPUUsage    map[string]uint64
	lastSystemUsage map[string]uint64
}

type dockerStatsJSON struct {
	ContainerID string `json:"container"`
	Name        string `json:"name"`
	CPUPerc     string `json:"cpu"`
	MemUsage    string `json:"mem_usage"`
	MemPerc     string `json:"mem_perc"`
	NetIO       string `json:"net_io"`
	BlockIO     string `json:"block_io"`
	PIDs        string `json:"pids"`
}

func NewDockerMonitor() (repositories.DockerMonitor, error) {
	// Check if docker command is available
	_, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("docker command not found: %w", err)
	}

	return &DockerMonitorImpl{
		lastCPUTime:     time.Now(),
		lastCPUUsage:    make(map[string]uint64),
		lastSystemUsage: make(map[string]uint64),
	}, nil
}

func (dm *DockerMonitorImpl) GetDockerMetrics(ctx context.Context, containerID string) (*monitoring.DockerMetrics, error) {
	// Get container name
	containerName, err := dm.GetContainerName(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container name: %w", err)
	}

	// Get container stats using docker stats command
	cmd := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--format",
		"{{.Container}}|{{.Name}}|{{.CPUPerc}}|{{.MemUsage}}|{{.MemPerc}}|{{.NetIO}}|{{.BlockIO}}|{{.PIDs}}", containerID)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}

	// Parse the output
	stats := strings.TrimSpace(string(output))
	if stats == "" {
		return nil, fmt.Errorf("no stats returned for container %s", containerID)
	}

	parts := strings.Split(stats, "|")
	if len(parts) != 8 {
		return nil, fmt.Errorf("unexpected stats format: %s", stats)
	}

	// Parse CPU percentage
	cpuPercStr := strings.TrimSuffix(parts[2], "%")
	cpuPerc, _ := strconv.ParseFloat(cpuPercStr, 64)

	// Parse memory usage and percentage
	memUsageParts := strings.Split(parts[3], " / ")
	var memUsage, memLimit uint64
	if len(memUsageParts) == 2 {
		memUsage = dm.parseMemoryValue(memUsageParts[0])
		memLimit = dm.parseMemoryValue(memUsageParts[1])
	}

	memPercStr := strings.TrimSuffix(parts[4], "%")
	memPerc, _ := strconv.ParseFloat(memPercStr, 64)

	// Parse network I/O
	netIOParts := strings.Split(parts[5], " / ")
	var netRx, netTx uint64
	if len(netIOParts) == 2 {
		netRx = dm.parseMemoryValue(netIOParts[0])
		netTx = dm.parseMemoryValue(netIOParts[1])
	}

	// Parse block I/O
	blockIOParts := strings.Split(parts[6], " / ")
	var blockRead, blockWrite uint64
	if len(blockIOParts) == 2 {
		blockRead = dm.parseMemoryValue(blockIOParts[0])
		blockWrite = dm.parseMemoryValue(blockIOParts[1])
	}

	// Parse PIDs (not used in current implementation but keeping for future use)
	_, _ = strconv.Atoi(parts[7])

	return &monitoring.DockerMetrics{
		ContainerID:   containerID,
		ContainerName: containerName,
		CPU: monitoring.DockerCPUMetrics{
			UsagePercent: cpuPerc,
		},
		Memory: monitoring.DockerMemoryMetrics{
			Usage:   memUsage,
			Limit:   memLimit,
			Percent: memPerc,
		},
		Network: monitoring.DockerNetworkMetrics{
			RxBytes: netRx,
			TxBytes: netTx,
		},
		BlockIO: monitoring.DockerBlockIOMetrics{
			ReadBytes:  blockRead,
			WriteBytes: blockWrite,
		},
	}, nil
}

func (dm *DockerMonitorImpl) ListContainers(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "docker", "ps", "-q")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	containerIDs := strings.Fields(strings.TrimSpace(string(output)))
	return containerIDs, nil
}

func (dm *DockerMonitorImpl) GetContainerName(ctx context.Context, containerID string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.Name}}", containerID)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get container name: %w", err)
	}

	name := strings.TrimSpace(string(output))
	return strings.TrimPrefix(name, "/"), nil
}

func (dm *DockerMonitorImpl) parseMemoryValue(value string) uint64 {
	value = strings.TrimSpace(value)
	if value == "0B" || value == "" {
		return 0
	}

	// Remove unit and convert
	var multiplier uint64 = 1
	if strings.HasSuffix(value, "kB") {
		multiplier = 1024
		value = strings.TrimSuffix(value, "kB")
	} else if strings.HasSuffix(value, "MB") {
		multiplier = 1024 * 1024
		value = strings.TrimSuffix(value, "MB")
	} else if strings.HasSuffix(value, "GB") {
		multiplier = 1024 * 1024 * 1024
		value = strings.TrimSuffix(value, "GB")
	} else if strings.HasSuffix(value, "TB") {
		multiplier = 1024 * 1024 * 1024 * 1024
		value = strings.TrimSuffix(value, "TB")
	} else if strings.HasSuffix(value, "B") {
		value = strings.TrimSuffix(value, "B")
	}

	if num, err := strconv.ParseFloat(value, 64); err == nil {
		return uint64(num * float64(multiplier))
	}

	return 0
}

