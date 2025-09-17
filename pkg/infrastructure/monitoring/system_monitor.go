package monitoring

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"cassandra-benchmark/pkg/domain/entities/monitoring"
	"cassandra-benchmark/pkg/domain/repositories"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type SystemMonitorImpl struct {
	lastNetworkStats map[string]net.IOCountersStat
	lastCollectTime  time.Time
}

func NewSystemMonitor() repositories.SystemMonitor {
	return &SystemMonitorImpl{
		lastNetworkStats: make(map[string]net.IOCountersStat),
		lastCollectTime:  time.Now(),
	}
}

func (sm *SystemMonitorImpl) GetCPUMetrics(ctx context.Context) (*monitoring.CPUMetrics, error) {
	// Get overall CPU usage
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	// Get per-core usage
	perCoreUsage, err := cpu.Percent(time.Second, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get per-core CPU usage: %w", err)
	}

	// Get load average
	loadAvg, err := load.Avg()
	if err != nil {
		return nil, fmt.Errorf("failed to get load average: %w", err)
	}

	return &monitoring.CPUMetrics{
		UsagePercent: percentages[0],
		LoadAverage: monitoring.LoadAverageMetrics{
			Load1:  loadAvg.Load1,
			Load5:  loadAvg.Load5,
			Load15: loadAvg.Load15,
		},
		Cores:        runtime.NumCPU(),
		PerCoreUsage: perCoreUsage,
	}, nil
}

func (sm *SystemMonitorImpl) GetMemoryMetrics(ctx context.Context) (*monitoring.MemoryMetrics, error) {
	// Get virtual memory stats
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}

	// Get swap memory stats
	swapStat, err := mem.SwapMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get swap stats: %w", err)
	}

	return &monitoring.MemoryMetrics{
		TotalBytes:     vmStat.Total,
		UsedBytes:      vmStat.Used,
		FreeBytes:      vmStat.Free,
		UsagePercent:   vmStat.UsedPercent,
		SwapTotal:      swapStat.Total,
		SwapUsed:       swapStat.Used,
		SwapFree:       swapStat.Free,
		SwapPercent:    swapStat.UsedPercent,
		BuffersBytes:   vmStat.Buffers,
		CachedBytes:    vmStat.Cached,
		AvailableBytes: vmStat.Available,
	}, nil
}

func (sm *SystemMonitorImpl) GetNetworkMetrics(ctx context.Context) (*monitoring.NetworkMetrics, error) {
	// Get network I/O counters
	netIOCounters, err := net.IOCounters(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get network stats: %w", err)
	}

	currentTime := time.Now()
	timeDiff := currentTime.Sub(sm.lastCollectTime).Seconds()

	var interfaces []monitoring.NetworkInterface
	var totalRxBytes, totalTxBytes uint64
	var totalRxRate, totalTxRate float64

	for _, netIO := range netIOCounters {
		// Skip loopback and inactive interfaces
		if netIO.Name == "lo" || netIO.Name == "lo0" {
			continue
		}

		// Calculate rates if we have previous data
		var rxRate, txRate float64
		if lastStats, exists := sm.lastNetworkStats[netIO.Name]; exists && timeDiff > 0 {
			rxDiff := float64(netIO.BytesRecv - lastStats.BytesRecv)
			txDiff := float64(netIO.BytesSent - lastStats.BytesSent)
			rxRate = (rxDiff * 8) / (timeDiff * 1024 * 1024) // Convert to Mbps
			txRate = (txDiff * 8) / (timeDiff * 1024 * 1024) // Convert to Mbps
		}

		interfaces = append(interfaces, monitoring.NetworkInterface{
			Name:      netIO.Name,
			RxBytes:   netIO.BytesRecv,
			TxBytes:   netIO.BytesSent,
			RxPackets: netIO.PacketsRecv,
			TxPackets: netIO.PacketsSent,
			RxErrors:  netIO.Errin,
			TxErrors:  netIO.Errout,
			RxRate:    rxRate,
			TxRate:    txRate,
		})

		totalRxBytes += netIO.BytesRecv
		totalTxBytes += netIO.BytesSent
		totalRxRate += rxRate
		totalTxRate += txRate

		// Store current stats for next calculation
		sm.lastNetworkStats[netIO.Name] = netIO
	}

	sm.lastCollectTime = currentTime

	return &monitoring.NetworkMetrics{
		Interfaces:   interfaces,
		TotalRxBytes: totalRxBytes,
		TotalTxBytes: totalTxBytes,
		TotalRxRate:  totalRxRate,
		TotalTxRate:  totalTxRate,
	}, nil
}

func (sm *SystemMonitorImpl) GetGPUMetrics(ctx context.Context) (*monitoring.GPUMetrics, error) {
	// Check if nvidia-smi is available
	_, err := exec.LookPath("nvidia-smi")
	if err != nil {
		return &monitoring.GPUMetrics{Devices: []monitoring.GPUDevice{}}, nil
	}

	// Run nvidia-smi to get GPU information
	cmd := exec.CommandContext(ctx, "nvidia-smi", 
		"--query-gpu=index,name,driver_version,utilization.gpu,memory.total,memory.used,memory.free,temperature.gpu,power.draw,power.limit,fan.speed,clocks.gr,clocks.mem,pids",
		"--format=csv,noheader,nounits")
	
	output, err := cmd.Output()
	if err != nil {
		return &monitoring.GPUMetrics{Devices: []monitoring.GPUDevice{}}, nil
	}

	var devices []monitoring.GPUDevice
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		fields := strings.Split(line, ", ")
		if len(fields) < 14 {
			continue
		}

		device := monitoring.GPUDevice{}
		
		if index, err := strconv.Atoi(strings.TrimSpace(fields[0])); err == nil {
			device.Index = index
		}
		
		device.Name = strings.TrimSpace(fields[1])
		device.DriverVersion = strings.TrimSpace(fields[2])
		
		if usage, err := strconv.ParseFloat(strings.TrimSpace(fields[3]), 64); err == nil {
			device.UsagePercent = usage
		}
		
		if memTotal, err := strconv.ParseUint(strings.TrimSpace(fields[4]), 10, 64); err == nil {
			device.MemoryTotal = memTotal
		}
		
		if memUsed, err := strconv.ParseUint(strings.TrimSpace(fields[5]), 10, 64); err == nil {
			device.MemoryUsed = memUsed
		}
		
		if memFree, err := strconv.ParseUint(strings.TrimSpace(fields[6]), 10, 64); err == nil {
			device.MemoryFree = memFree
		}
		
		if device.MemoryTotal > 0 {
			device.MemoryPercent = float64(device.MemoryUsed) / float64(device.MemoryTotal) * 100
		}
		
		if temp, err := strconv.ParseFloat(strings.TrimSpace(fields[7]), 64); err == nil {
			device.Temperature = temp
		}
		
		if power, err := strconv.ParseFloat(strings.TrimSpace(fields[8]), 64); err == nil {
			device.PowerUsage = power
		}
		
		if powerLimit, err := strconv.ParseFloat(strings.TrimSpace(fields[9]), 64); err == nil {
			device.PowerLimit = powerLimit
		}
		
		if fanSpeed, err := strconv.ParseFloat(strings.TrimSpace(fields[10]), 64); err == nil {
			device.FanSpeed = fanSpeed
		}
		
		if clockGfx, err := strconv.Atoi(strings.TrimSpace(fields[11])); err == nil {
			device.ClockGraphics = clockGfx
		}
		
		if clockMem, err := strconv.Atoi(strings.TrimSpace(fields[12])); err == nil {
			device.ClockMemory = clockMem
		}
		
		// Count processes
		pidsField := strings.TrimSpace(fields[13])
		if pidsField != "[Not Supported]" && pidsField != "" {
			device.ProcessCount = strings.Count(pidsField, ",") + 1
		}

		devices = append(devices, device)
	}

	return &monitoring.GPUMetrics{Devices: devices}, nil
}