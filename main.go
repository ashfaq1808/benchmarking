package main

import (
	"benchmarking/client"
	"benchmarking/workload"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Cassandra struct {
		Nodes    []string `yaml:"nodes"`
		Keyspace string   `yaml:"keyspace"`
		Table    string   `yaml:"table"`
	} `yaml:"cassandra"`
	Benchmark struct {
		DurationSeconds int     `yaml:"duration_seconds"`
		WarmupSeconds   int     `yaml:"warmup_seconds"`
		Concurrency     int     `yaml:"concurrency"`
		ReadRatio       float64 `yaml:"read_ratio"`
		WriteRatio      float64 `yaml:"write_ratio"`
		Mode            string  `yaml:"mode"`
		LogFile         string  `yaml:"log_file"`
	} `yaml:"benchmark"`
}

func loadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
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

	session := client.Connect(cfg.Cassandra.Nodes, cfg.Cassandra.Keyspace)
	defer session.Close()

	err = workload.RunBenchmark(cfg, session)
	if err != nil {
		fmt.Println("Benchmark failed:", err)
		os.Exit(1)
	}
}
