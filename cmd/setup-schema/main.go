package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cassandra-benchmark/pkg/infrastructure/config"
	"cassandra-benchmark/pkg/infrastructure/database"
)

func main() {
	if len(os.Args) > 2 {
		fmt.Println("Usage: setup-schema [config-file]")
		fmt.Println("Default config file: config.yaml")
		os.Exit(1)
	}

	configPath := "config.yaml"
	if len(os.Args) == 2 {
		configPath = os.Args[1]
	}

	fmt.Printf("🚀 Setting up Cassandra schema from config: %s\n", configPath)
	
	// Load configuration
	configLoader := config.NewYamlConfigLoader()
	cfg, err := configLoader.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Create schema repository
	schemaRepo := database.NewCassandraSchemaRepository(cfg.Cassandra.Hosts)

	ctx := context.Background()

	// Check if keyspace exists
	fmt.Printf("🔍 Checking if keyspace '%s' exists...\n", cfg.Cassandra.Keyspace)
	keyspaceExists, err := schemaRepo.KeyspaceExists(ctx, cfg.Cassandra.Keyspace)
	if err != nil {
		log.Fatalf("❌ Failed to check keyspace existence: %v", err)
	}

	// Create keyspace if it doesn't exist
	if !keyspaceExists {
		fmt.Printf("🏗️ Creating keyspace '%s'...\n", cfg.Cassandra.Keyspace)
		err = schemaRepo.CreateKeyspace(ctx, cfg.Cassandra.Keyspace, cfg.Schema)
		if err != nil {
			log.Fatalf("❌ Failed to create keyspace: %v", err)
		}
	} else {
		fmt.Printf("✅ Keyspace '%s' already exists\n", cfg.Cassandra.Keyspace)
	}

	// Check if table exists
	fmt.Printf("🔍 Checking if table '%s.%s' exists...\n", cfg.Cassandra.Keyspace, cfg.Cassandra.Table)
	tableExists, err := schemaRepo.TableExists(ctx, cfg.Cassandra.Keyspace, cfg.Cassandra.Table)
	if err != nil {
		log.Fatalf("❌ Failed to check table existence: %v", err)
	}

	// Create table if it doesn't exist
	if !tableExists {
		fmt.Printf("🏗️ Creating table '%s.%s'...\n", cfg.Cassandra.Keyspace, cfg.Cassandra.Table)
		err = schemaRepo.CreateTable(ctx, cfg.Cassandra.Keyspace, cfg.Cassandra.Table, cfg.Schema)
		if err != nil {
			log.Fatalf("❌ Failed to create table: %v", err)
		}
	} else {
		fmt.Printf("✅ Table '%s.%s' already exists\n", cfg.Cassandra.Keyspace, cfg.Cassandra.Table)
	}

	// Create indexes
	if len(cfg.Schema.Indexes) > 0 {
		fmt.Printf("🏗️ Creating indexes...\n")
		err = schemaRepo.CreateIndexes(ctx, cfg.Cassandra.Keyspace, cfg.Cassandra.Table, cfg.Schema.Indexes)
		if err != nil {
			log.Fatalf("❌ Failed to create indexes: %v", err)
		}
	}

	fmt.Println("\n🎉 Schema setup completed successfully!")
	fmt.Printf("📊 Database Configuration:\n")
	fmt.Printf("   • Keyspace: %s\n", cfg.Cassandra.Keyspace)
	fmt.Printf("   • Table: %s\n", cfg.Cassandra.Table)
	fmt.Printf("   • Replication Strategy: %s\n", cfg.Schema.ReplicationStrategy)
	fmt.Printf("   • Replication Factor: %d\n", cfg.Schema.ReplicationFactor)
	fmt.Printf("   • Indexes: %d created\n", len(cfg.Schema.Indexes))
	fmt.Printf("   • Hosts: %v\n", cfg.Cassandra.Hosts)
}