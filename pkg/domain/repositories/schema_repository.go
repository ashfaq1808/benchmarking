package repositories

import (
	"context"
	"cassandra-benchmark/pkg/domain/entities"
)

type SchemaRepository interface {
	CreateKeyspace(ctx context.Context, keyspaceName string, config entities.SchemaConfig) error
	CreateTable(ctx context.Context, keyspaceName, tableName string, config entities.SchemaConfig) error
	CreateIndexes(ctx context.Context, keyspaceName, tableName string, indexes []entities.IndexConfig) error
	KeyspaceExists(ctx context.Context, keyspaceName string) (bool, error)
	TableExists(ctx context.Context, keyspaceName, tableName string) (bool, error)
}