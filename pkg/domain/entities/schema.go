package entities

type SchemaConfig struct {
	ReplicationStrategy string                 `yaml:"replication_strategy"`
	ReplicationFactor   int                    `yaml:"replication_factor"`
	NetworkTopology     map[string]int         `yaml:"network_topology"`
	TableOptions        TableOptions           `yaml:"table_options"`
	Indexes             []IndexConfig          `yaml:"indexes"`
}

type TableOptions struct {
	BloomFilterFpChance      float64           `yaml:"bloom_filter_fp_chance"`
	Caching                  map[string]string `yaml:"caching"`
	Comment                  string            `yaml:"comment"`
	CompactionStrategy       string            `yaml:"compaction_strategy"`
	CompressionAlgorithm     string            `yaml:"compression_algorithm"`
	GcGraceSeconds          int               `yaml:"gc_grace_seconds"`
}

type IndexConfig struct {
	Name   string `yaml:"name"`
	Column string `yaml:"column"`
}