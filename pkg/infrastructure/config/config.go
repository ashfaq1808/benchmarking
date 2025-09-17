package config

import (
	"os"
	"time"
	"cassandra-benchmark/pkg/domain/entities"
	"cassandra-benchmark/pkg/domain/entities/monitoring"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Cassandra entities.DatabaseConfig    `yaml:"cassandra"`
	Benchmark entities.BenchmarkConfig   `yaml:"benchmark"`
	Schema    entities.SchemaConfig      `yaml:"schema"`
}

type ConfigLoader interface {
	LoadConfig(path string) (*Config, error)
}

type YamlConfigLoader struct{}

func NewYamlConfigLoader() *YamlConfigLoader {
	return &YamlConfigLoader{}
}

func (y *YamlConfigLoader) LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var yamlConfig struct {
		Cassandra struct {
			Hosts    []string `yaml:"hosts"`
			Keyspace string   `yaml:"keyspace"`
			Table    string   `yaml:"table"`
		} `yaml:"cassandra"`
		Benchmark struct {
			DurationSeconds   int     `yaml:"duration_seconds"`
			WarmupSeconds     int     `yaml:"warmup_seconds"`
			Concurrency       int     `yaml:"concurrency"`
			ReadRatio         float64 `yaml:"read_ratio"`
			WriteRatio        float64 `yaml:"write_ratio"`
			Mode              string  `yaml:"mode"`
			LogFile           string  `yaml:"log_file"`
			RequestsPerSecond int     `yaml:"requests_per_second"`
			RatePattern       struct {
				Enabled         bool    `yaml:"enabled"`
				Mode            string  `yaml:"mode"`
				MinRate         int     `yaml:"min_rate"`
				MaxRate         int     `yaml:"max_rate"`
				PeakDuration    int     `yaml:"peak_duration_seconds"`
				ValleyDuration  int     `yaml:"valley_duration_seconds"`
				ChangeInterval  float64 `yaml:"change_interval_seconds"`
			} `yaml:"rate_pattern"`
		} `yaml:"benchmark"`
		Schema struct {
			ReplicationStrategy   string                 `yaml:"replication_strategy"`
			ReplicationFactor     int                    `yaml:"replication_factor"`
			NetworkTopology       map[string]int         `yaml:"network_topology"`
			TableOptions          struct {
				BloomFilterFpChance   float64           `yaml:"bloom_filter_fp_chance"`
				Caching               map[string]string `yaml:"caching"`
				Comment               string            `yaml:"comment"`
				CompactionStrategy    string            `yaml:"compaction_strategy"`
				CompressionAlgorithm  string            `yaml:"compression_algorithm"`
				GcGraceSeconds        int               `yaml:"gc_grace_seconds"`
			} `yaml:"table_options"`
			Indexes []struct {
				Name   string `yaml:"name"`
				Column string `yaml:"column"`
			} `yaml:"indexes"`
		} `yaml:"schema"`
		Monitoring struct {
			Enabled           bool   `yaml:"enabled"`
			IntervalSeconds   int    `yaml:"interval_seconds"`
			EnableGPU         bool   `yaml:"enable_gpu"`
			EnableDocker      bool   `yaml:"enable_docker"`
			DockerContainerID string `yaml:"docker_container_id"`
			OutputFile        string `yaml:"output_file"`
		} `yaml:"monitoring"`
	}
	
	err = yaml.Unmarshal(data, &yamlConfig)
	if err != nil {
		return nil, err
	}
	
	// Convert indexes
	indexes := make([]entities.IndexConfig, len(yamlConfig.Schema.Indexes))
	for i, idx := range yamlConfig.Schema.Indexes {
		indexes[i] = entities.IndexConfig{
			Name:   idx.Name,
			Column: idx.Column,
		}
	}

	config := &Config{
		Cassandra: entities.DatabaseConfig{
			Hosts:    yamlConfig.Cassandra.Hosts,
			Keyspace: yamlConfig.Cassandra.Keyspace,
			Table:    yamlConfig.Cassandra.Table,
		},
		Benchmark: entities.BenchmarkConfig{
			DurationSeconds:   yamlConfig.Benchmark.DurationSeconds,
			WarmupSeconds:     yamlConfig.Benchmark.WarmupSeconds,
			Concurrency:       yamlConfig.Benchmark.Concurrency,
			ReadRatio:         yamlConfig.Benchmark.ReadRatio,
			WriteRatio:        yamlConfig.Benchmark.WriteRatio,
			Mode:              yamlConfig.Benchmark.Mode,
			RequestsPerSecond: yamlConfig.Benchmark.RequestsPerSecond,
			RatePattern: entities.RatePatternConfig{
				Enabled:         yamlConfig.Benchmark.RatePattern.Enabled,
				Mode:            yamlConfig.Benchmark.RatePattern.Mode,
				MinRate:         yamlConfig.Benchmark.RatePattern.MinRate,
				MaxRate:         yamlConfig.Benchmark.RatePattern.MaxRate,
				PeakDuration:    yamlConfig.Benchmark.RatePattern.PeakDuration,
				ValleyDuration:  yamlConfig.Benchmark.RatePattern.ValleyDuration,
				ChangeInterval:  yamlConfig.Benchmark.RatePattern.ChangeInterval,
			},
			Monitoring: monitoring.MonitoringConfig{
				Enabled:           yamlConfig.Monitoring.Enabled,
				Interval:          time.Duration(yamlConfig.Monitoring.IntervalSeconds) * time.Second,
				EnableGPU:         yamlConfig.Monitoring.EnableGPU,
				EnableDocker:      yamlConfig.Monitoring.EnableDocker,
				DockerContainerID: yamlConfig.Monitoring.DockerContainerID,
				OutputFile:        yamlConfig.Monitoring.OutputFile,
			},
		},
		Schema: entities.SchemaConfig{
			ReplicationStrategy: yamlConfig.Schema.ReplicationStrategy,
			ReplicationFactor:   yamlConfig.Schema.ReplicationFactor,
			NetworkTopology:     yamlConfig.Schema.NetworkTopology,
			TableOptions: entities.TableOptions{
				BloomFilterFpChance:  yamlConfig.Schema.TableOptions.BloomFilterFpChance,
				Caching:              yamlConfig.Schema.TableOptions.Caching,
				Comment:              yamlConfig.Schema.TableOptions.Comment,
				CompactionStrategy:   yamlConfig.Schema.TableOptions.CompactionStrategy,
				CompressionAlgorithm: yamlConfig.Schema.TableOptions.CompressionAlgorithm,
				GcGraceSeconds:       yamlConfig.Schema.TableOptions.GcGraceSeconds,
			},
			Indexes: indexes,
		},
	}
	
	return config, nil
}