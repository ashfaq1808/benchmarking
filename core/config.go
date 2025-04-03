package core

import (
    "flag"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "strings"
)

type Config struct {
    Hosts       []string `yaml:"hosts"`
    Keyspace    string   `yaml:"keyspace"`
    Table       string   `yaml:"table"`
    TotalReads  int      `yaml:"reads"`
    TotalWrites int      `yaml:"writes"`
    Concurrency int      `yaml:"concurrency"`
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
