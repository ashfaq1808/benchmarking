package database

import (
	"context"
	"fmt"
	"strings"

	"cassandra-benchmark/pkg/domain/entities"
	"github.com/gocql/gocql"
)

type CassandraSchemaRepository struct {
	cluster *gocql.ClusterConfig
}

func NewCassandraSchemaRepository(hosts []string) *CassandraSchemaRepository {
	cluster := gocql.NewCluster(hosts...)
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 4
	return &CassandraSchemaRepository{cluster: cluster}
}

func (csr *CassandraSchemaRepository) KeyspaceExists(ctx context.Context, keyspaceName string) (bool, error) {
	session, err := csr.cluster.CreateSession()
	if err != nil {
		return false, err
	}
	defer session.Close()

	var name string
	query := "SELECT keyspace_name FROM system_schema.keyspaces WHERE keyspace_name = ?"
	err = session.Query(query, keyspaceName).Scan(&name)
	if err != nil {
		if err == gocql.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (csr *CassandraSchemaRepository) TableExists(ctx context.Context, keyspaceName, tableName string) (bool, error) {
	// First check if keyspace exists
	keyspaceExists, err := csr.KeyspaceExists(ctx, keyspaceName)
	if err != nil || !keyspaceExists {
		return false, err
	}

	// Connect without specifying keyspace to query system tables
	session, err := csr.cluster.CreateSession()
	if err != nil {
		return false, err
	}
	defer session.Close()

	var name string
	query := "SELECT table_name FROM system_schema.tables WHERE keyspace_name = ? AND table_name = ?"
	err = session.Query(query, keyspaceName, tableName).Scan(&name)
	if err != nil {
		if err == gocql.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (csr *CassandraSchemaRepository) CreateKeyspace(ctx context.Context, keyspaceName string, config entities.SchemaConfig) error {
	session, err := csr.cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var replicationConfig string
	if config.ReplicationStrategy == "NetworkTopologyStrategy" && len(config.NetworkTopology) > 0 {
		var parts []string
		parts = append(parts, "'class': 'NetworkTopologyStrategy'")
		for datacenter, factor := range config.NetworkTopology {
			parts = append(parts, fmt.Sprintf("'%s': %d", datacenter, factor))
		}
		replicationConfig = strings.Join(parts, ", ")
	} else {
		// Default to SimpleStrategy
		replicationFactor := config.ReplicationFactor
		if replicationFactor == 0 {
			replicationFactor = 3
		}
		replicationConfig = fmt.Sprintf("'class': 'SimpleStrategy', 'replication_factor': %d", replicationFactor)
	}

	query := fmt.Sprintf(`
		CREATE KEYSPACE IF NOT EXISTS %s
		WITH REPLICATION = {%s}`,
		keyspaceName, replicationConfig)

	err = session.Query(query).Exec()
	if err != nil {
		return fmt.Errorf("failed to create keyspace %s: %w", keyspaceName, err)
	}

	fmt.Printf("✅ Keyspace '%s' created successfully\n", keyspaceName)
	return nil
}

func (csr *CassandraSchemaRepository) CreateTable(ctx context.Context, keyspaceName, tableName string, config entities.SchemaConfig) error {
	// Connect to the specific keyspace now that it exists
	cluster := *csr.cluster
	cluster.Keyspace = keyspaceName
	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create session for keyspace %s: %w", keyspaceName, err)
	}
	defer session.Close()

	// Build the base table creation query
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id UUID PRIMARY KEY,
			category TEXT,
			data TEXT,
			timestamp TIMESTAMP,
			value BIGINT
		)`, tableName)

	// Add table options
	var options []string
	
	if config.TableOptions.BloomFilterFpChance > 0 {
		options = append(options, fmt.Sprintf("bloom_filter_fp_chance = %.3f", config.TableOptions.BloomFilterFpChance))
	}
	
	if len(config.TableOptions.Caching) > 0 {
		cachingParts := []string{}
		for k, v := range config.TableOptions.Caching {
			cachingParts = append(cachingParts, fmt.Sprintf("'%s': '%s'", k, v))
		}
		options = append(options, fmt.Sprintf("caching = {%s}", strings.Join(cachingParts, ", ")))
	}
	
	if config.TableOptions.Comment != "" {
		options = append(options, fmt.Sprintf("comment = '%s'", config.TableOptions.Comment))
	}
	
	if config.TableOptions.CompactionStrategy != "" {
		options = append(options, fmt.Sprintf("compaction = {'class': '%s'}", config.TableOptions.CompactionStrategy))
	}
	
	if config.TableOptions.CompressionAlgorithm != "" {
		options = append(options, fmt.Sprintf("compression = {'class': '%s'}", config.TableOptions.CompressionAlgorithm))
	}
	
	if config.TableOptions.GcGraceSeconds > 0 {
		options = append(options, fmt.Sprintf("gc_grace_seconds = %d", config.TableOptions.GcGraceSeconds))
	}

	if len(options) > 0 {
		query += " WITH " + strings.Join(options, "\nAND ")
	}

	err = session.Query(query).Exec()
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	fmt.Printf("✅ Table '%s.%s' created successfully\n", keyspaceName, tableName)
	return nil
}

func (csr *CassandraSchemaRepository) CreateIndexes(ctx context.Context, keyspaceName, tableName string, indexes []entities.IndexConfig) error {
	if len(indexes) == 0 {
		return nil
	}

	// Connect to the specific keyspace now that it exists
	cluster := *csr.cluster
	cluster.Keyspace = keyspaceName
	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to create session for keyspace %s: %w", keyspaceName, err)
	}
	defer session.Close()

	for _, index := range indexes {
		query := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", index.Name, tableName, index.Column)
		err = session.Query(query).Exec()
		if err != nil {
			return fmt.Errorf("failed to create index %s: %w", index.Name, err)
		}
		fmt.Printf("✅ Index '%s' created successfully on column '%s'\n", index.Name, index.Column)
	}

	return nil
}