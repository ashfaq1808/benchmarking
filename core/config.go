package core

import (
	"flag"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Hosts            []string `yaml:"hosts"`
	Keyspace         string   `yaml:"keyspace"`
	Table            string   `yaml:"table"`
	TotalReads       int      `yaml:"reads"`
	TotalWrites      int      `yaml:"writes"`
	Concurrency      int      `yaml:"concurrency"`
	WarmupTime       int      `yaml:"warmup_time"`
	Duration         int      `yaml:"duration"`
	LoopType         string   `yaml:"loop_type"` // open or closed
	MaxConcurrency   int      `yaml:"max_concurrency"`
	FailureThreshold float64  `yaml:"failure_threshold"` // 0.05 means 5% failure rate allowed
}

func LoadConfigFromFlags() *Config {
	var hosts string
	cfg := &Config{}

	flag.StringVar(&hosts, "hosts", "127.0.0.1", "Comma-separated Cassandra node IPs")
	flag.StringVar(&cfg.Keyspace, "keyspace", "benchmark", "Cassandra keyspace")
	flag.StringVar(&cfg.Table, "table", "benchmark_table", "Cassandra table")
	flag.IntVar(&cfg.TotalReads, "reads", 500, "Total read operations")
	flag.IntVar(&cfg.TotalWrites, "writes", 500, "Total write operations")
	flag.IntVar(&cfg.Concurrency, "concurrency", 50, "Number of concurrent goroutines")
	flag.IntVar(&cfg.WarmupTime, "warmup_time", 10, "Warmup time in seconds")
	flag.IntVar(&cfg.Duration, "duration", 60, "Benchmark duration in seconds")
	flag.StringVar(&cfg.LoopType, "loop_type", "open", "Benchmark type: open or closed")
	flag.IntVar(&cfg.MaxConcurrency, "max_concurrency", 1000, "Max concurrency for load test")
	flag.Float64Var(&cfg.FailureThreshold, "failure_threshold", 0.05, "Failure rate threshold (5% = 0.05)")
	flag.Parse()

	cfg.Hosts = strings.Split(hosts, ",")
	return cfg
}

func LoadConfigFromYAML(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
