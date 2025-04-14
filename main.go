package main

import (
	"cassandra-benchmark/client"
	"cassandra-benchmark/config"
	"cassandra-benchmark/result"
	"cassandra-benchmark/workload"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Cassandra struct {
		Nodes    []string `yaml:"nodes"`
		Keyspace string   `yaml:"keyspace"`
		Table    string   `yaml:"table"`
	} `yaml:"cassandra"`
	Benchmark struct {
		DurationSeconds   int     `yaml:"duration_seconds"`
		RequestsPerSecond int     `yaml:"requests_per_second"`
		WarmupSeconds     int     `yaml:"warmup_seconds"`
		Concurrency       int     `yaml:"concurrency"`
		ReadRatio         float64 `yaml:"read_ratio"`
		WriteRatio        float64 `yaml:"write_ratio"`
		Mode              string  `yaml:"mode"`
		LogFile           string  `yaml:"log_file"`
	} `yaml:"benchmark"`
}

func loadConfig(path string) (*config.Config, error) {
	data, err := ioutil.ReadFile(path)
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
	result.InitLogFile()
	defer result.CloseLogFile()

	session := client.Connect(cfg.Cassandra.Nodes, cfg.Cassandra.Keyspace)
	defer session.Close()

	err = workload.RunBenchmark(cfg, session)
	if err != nil {
		fmt.Println("Benchmark failed:", err)
		os.Exit(1)
	} else {
		result.CloseLogFile()

		fmt.Println("Launching Streamlit dashboard at http://localhost:8501 ...")
		cmd := exec.Command("streamlit", "run", "visualize.py")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("Streamlit error:", err)
		}
	}

}
