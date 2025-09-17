package config

type Config struct {
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
			Enabled         bool   `yaml:"enabled"`
			Mode            string `yaml:"mode"`
			MinRate         int    `yaml:"min_rate"`
			MaxRate         int    `yaml:"max_rate"`
			PeakDuration    int    `yaml:"peak_duration_seconds"`
			ValleyDuration  int    `yaml:"valley_duration_seconds"`
			ChangeInterval  float64 `yaml:"change_interval_seconds"`
		} `yaml:"rate_pattern"`
	} `yaml:"benchmark"`
}
