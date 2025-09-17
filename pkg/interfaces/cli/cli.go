package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"cassandra-benchmark/pkg/domain/services"
	"cassandra-benchmark/pkg/infrastructure/config"
)

type CLI struct {
	benchmarkService services.BenchmarkService
	configLoader     config.ConfigLoader
}

func NewCLI(benchmarkService services.BenchmarkService, configLoader config.ConfigLoader) *CLI {
	return &CLI{
		benchmarkService: benchmarkService,
		configLoader:     configLoader,
	}
}

func (c *CLI) Run(configPath string) error {
	cfg, err := c.configLoader.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx := context.Background()
	err = c.benchmarkService.RunBenchmark(ctx, cfg.Benchmark)
	if err != nil {
		return fmt.Errorf("benchmark failed: %w", err)
	}

	fmt.Println("Benchmarking complete ✅")
	
	// Check if enhanced visualization exists
	var vizFile string
	if _, err := os.Stat("visualize_enhanced.py"); err == nil {
		vizFile = "visualize_enhanced.py"
		fmt.Println("Launching Enhanced Dashboard with System Monitoring at http://localhost:8501 ...")
	} else {
		vizFile = "visualize.py"
		fmt.Println("Launching Streamlit dashboard at http://localhost:8501 ...")
	}

	cmd := exec.Command("streamlit", "run", vizFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Printf("Streamlit error: %v", err)
	}

	return nil
}