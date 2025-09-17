package monitoring

import (
	"time"
)

type SystemMetrics struct {
	Timestamp time.Time     `json:"timestamp"`
	CPU       CPUMetrics    `json:"cpu"`
	Memory    MemoryMetrics `json:"memory"`
	Network   NetworkMetrics `json:"network"`
	GPU       *GPUMetrics   `json:"gpu,omitempty"`
	Docker    *DockerMetrics `json:"docker,omitempty"`
}

type CPUMetrics struct {
	UsagePercent float64            `json:"usage_percent"`
	LoadAverage  LoadAverageMetrics `json:"load_average"`
	Cores        int                `json:"cores"`
	PerCoreUsage []float64          `json:"per_core_usage"`
}

type LoadAverageMetrics struct {
	Load1  float64 `json:"load_1m"`
	Load5  float64 `json:"load_5m"`
	Load15 float64 `json:"load_15m"`
}

type MemoryMetrics struct {
	TotalBytes     uint64  `json:"total_bytes"`
	UsedBytes      uint64  `json:"used_bytes"`
	FreeBytes      uint64  `json:"free_bytes"`
	UsagePercent   float64 `json:"usage_percent"`
	SwapTotal      uint64  `json:"swap_total_bytes"`
	SwapUsed       uint64  `json:"swap_used_bytes"`
	SwapFree       uint64  `json:"swap_free_bytes"`
	SwapPercent    float64 `json:"swap_usage_percent"`
	BuffersBytes   uint64  `json:"buffers_bytes"`
	CachedBytes    uint64  `json:"cached_bytes"`
	AvailableBytes uint64  `json:"available_bytes"`
}

type NetworkMetrics struct {
	Interfaces []NetworkInterface `json:"interfaces"`
	TotalRxBytes uint64           `json:"total_rx_bytes"`
	TotalTxBytes uint64           `json:"total_tx_bytes"`
	TotalRxRate  float64          `json:"total_rx_rate_mbps"`
	TotalTxRate  float64          `json:"total_tx_rate_mbps"`
}

type NetworkInterface struct {
	Name      string  `json:"name"`
	RxBytes   uint64  `json:"rx_bytes"`
	TxBytes   uint64  `json:"tx_bytes"`
	RxPackets uint64  `json:"rx_packets"`
	TxPackets uint64  `json:"tx_packets"`
	RxErrors  uint64  `json:"rx_errors"`
	TxErrors  uint64  `json:"tx_errors"`
	RxRate    float64 `json:"rx_rate_mbps"`
	TxRate    float64 `json:"tx_rate_mbps"`
}

type GPUMetrics struct {
	Devices []GPUDevice `json:"devices"`
}

type GPUDevice struct {
	Index            int     `json:"index"`
	Name             string  `json:"name"`
	DriverVersion    string  `json:"driver_version"`
	UsagePercent     float64 `json:"usage_percent"`
	MemoryTotal      uint64  `json:"memory_total_mb"`
	MemoryUsed       uint64  `json:"memory_used_mb"`
	MemoryFree       uint64  `json:"memory_free_mb"`
	MemoryPercent    float64 `json:"memory_usage_percent"`
	Temperature      float64 `json:"temperature_celsius"`
	PowerUsage       float64 `json:"power_usage_watts"`
	PowerLimit       float64 `json:"power_limit_watts"`
	FanSpeed         float64 `json:"fan_speed_percent"`
	ClockGraphics    int     `json:"clock_graphics_mhz"`
	ClockMemory      int     `json:"clock_memory_mhz"`
	ProcessCount     int     `json:"process_count"`
}

type DockerMetrics struct {
	ContainerID     string              `json:"container_id"`
	ContainerName   string              `json:"container_name"`
	CPU             DockerCPUMetrics    `json:"cpu"`
	Memory          DockerMemoryMetrics `json:"memory"`
	Network         DockerNetworkMetrics `json:"network"`
	BlockIO         DockerBlockIOMetrics `json:"block_io"`
}

type DockerCPUMetrics struct {
	UsagePercent    float64 `json:"usage_percent"`
	UsageInUsermode uint64  `json:"usage_in_usermode"`
	UsageInKernelmode uint64  `json:"usage_in_kernelmode"`
	SystemCPUUsage  uint64  `json:"system_cpu_usage"`
	ThrottledTime   uint64  `json:"throttled_time"`
}

type DockerMemoryMetrics struct {
	Usage     uint64  `json:"usage_bytes"`
	MaxUsage  uint64  `json:"max_usage_bytes"`
	Limit     uint64  `json:"limit_bytes"`
	Percent   float64 `json:"usage_percent"`
	Cache     uint64  `json:"cache_bytes"`
	RSS       uint64  `json:"rss_bytes"`
	Swap      uint64  `json:"swap_bytes"`
}

type DockerNetworkMetrics struct {
	RxBytes   uint64 `json:"rx_bytes"`
	TxBytes   uint64 `json:"tx_bytes"`
	RxPackets uint64 `json:"rx_packets"`
	TxPackets uint64 `json:"tx_packets"`
	RxErrors  uint64 `json:"rx_errors"`
	TxErrors  uint64 `json:"tx_errors"`
	RxDropped uint64 `json:"rx_dropped"`
	TxDropped uint64 `json:"tx_dropped"`
}

type DockerBlockIOMetrics struct {
	ReadBytes  uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
	ReadOps    uint64 `json:"read_ops"`
	WriteOps   uint64 `json:"write_ops"`
}

type MonitoringConfig struct {
	Enabled           bool          `json:"enabled"`
	Interval          time.Duration `json:"interval"`
	EnableGPU         bool          `json:"enable_gpu"`
	EnableDocker      bool          `json:"enable_docker"`
	DockerContainerID string        `json:"docker_container_id"`
	OutputFile        string        `json:"output_file"`
}