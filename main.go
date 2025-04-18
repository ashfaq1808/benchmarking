// cassandra-benchmark/main.go
package main

import (
	"cassandra-benchmark/client"
	"cassandra-benchmark/config"
	"cassandra-benchmark/result"
	"cassandra-benchmark/workload"
	"fmt"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"
)

func loadConfig(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg config.Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	cfg, err := loadConfig("config.yaml")
	if err != nil {
		fmt.Println("Failed to load config:", err)
		os.Exit(1)
	}

	result.InitializeLogger(10000)
	result.StartFlusher()

	sessions := client.ConnectToAll(cfg.Cassandra.Nodes, cfg.Cassandra.Keyspace)
	defer func() {
		for _, s := range sessions {
			s.Close()
		}
	}()

	err = workload.RunBenchmark(cfg, sessions)
	if err != nil {
		fmt.Println("Benchmark failed:", err)
		os.Exit(1)
	}
	result.StopFlusher()
	fmt.Println("Benchmarking complete ✅")
	fmt.Println("Launching Streamlit dashboard at http://localhost:8501 ...")

	cmd := exec.Command("streamlit", "run", "visualize.py")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Streamlit error:", err)
	}
}
