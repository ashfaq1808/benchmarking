package config

type Config struct {
	Cassandra struct {
		Nodes    []string `yaml:"nodes"`
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
	} `yaml:"benchmark"`
}
