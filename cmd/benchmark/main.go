package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"cassandra-benchmark/pkg/application/benchmark"
	"cassandra-benchmark/pkg/application/workload"
	"cassandra-benchmark/pkg/domain/repositories"
	"cassandra-benchmark/pkg/infrastructure/config"
	"cassandra-benchmark/pkg/infrastructure/database"
	"cassandra-benchmark/pkg/infrastructure/logging"
	"cassandra-benchmark/pkg/infrastructure/monitoring"
	"cassandra-benchmark/pkg/interfaces/cli"
)

func main() {
	configLoader := config.NewYamlConfigLoader()
	cfg, err := configLoader.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Setup schema before connecting
	fmt.Println("🔧 Setting up Cassandra schema...")
	err = setupSchema(cfg)
	if err != nil {
		fmt.Printf("Failed to setup schema: %v\n", err)
		os.Exit(1)
	}

	sessionManager, err := database.NewCassandraSessionManager(cfg.Cassandra)
	if err != nil {
		fmt.Printf("Failed to connect to Cassandra: %v\n", err)
		os.Exit(1)
	}
	defer sessionManager.Close()

	employeeRepo := database.NewCassandraEmployeeRepository(sessionManager, cfg.Cassandra.Table)
	loggingRepo := logging.NewJsonLoggingRepository(10000)
	dataRepo := database.NewCsvDataRepository("employees_data.csv")

	// Setup monitoring if enabled
	var metricsCollector repositories.MetricsCollector
	if cfg.Benchmark.Monitoring.Enabled {
		fmt.Println("🔧 Initializing system monitoring...")

		systemMonitor := monitoring.NewSystemMonitor()

		// Initialize Docker monitor if enabled
		var dockerMonitor repositories.DockerMonitor
		if cfg.Benchmark.Monitoring.EnableDocker {
			dockerMon, err := monitoring.NewDockerMonitor()
			if err != nil {
				fmt.Printf("Warning: Failed to initialize Docker monitor: %v\n", err)
			} else {
				dockerMonitor = dockerMon
				fmt.Println("✅ Docker monitor initialized")
			}
		}

		metricsRepo := monitoring.NewJsonMetricsRepository(cfg.Benchmark.Monitoring.OutputFile)
		metricsCollector = monitoring.NewMetricsCollector(systemMonitor, dockerMonitor, metricsRepo)

		// Start monitoring
		err = metricsCollector.Start(context.Background(), cfg.Benchmark.Monitoring)
		if err != nil {
			fmt.Printf("Warning: Failed to start monitoring: %v\n", err)
		} else {
			fmt.Println("✅ System monitoring started")

			// Setup signal handling to ensure proper cleanup
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			// Cleanup function
			cleanup := func() {
				fmt.Println("📊 Stopping system monitoring...")
				metricsCollector.Stop()
				fmt.Println("📊 System monitoring stopped")
			}

			// Handle signals in a goroutine
			go func() {
				<-sigChan
				cleanup()
				os.Exit(0)
			}()

			defer cleanup()
		}
	}

	workloadService := workload.NewWorkloadService()
	benchmarkService := benchmark.NewBenchmarkService(workloadService, employeeRepo, loggingRepo, dataRepo)

	cli := cli.NewCLI(benchmarkService, configLoader)

	err = cli.Run("config.yaml")
	if err != nil {
		fmt.Printf("CLI error: %v\n", err)
		os.Exit(1)
	}
}

func setupSchema(cfg *config.Config) error {
	ctx := context.Background()
	schemaRepo := database.NewCassandraSchemaRepository(cfg.Cassandra.Hosts)

	// Check and create keyspace
	keyspaceExists, err := schemaRepo.KeyspaceExists(ctx, cfg.Cassandra.Keyspace)
	if err != nil {
		return fmt.Errorf("failed to check keyspace: %w", err)
	}

	if !keyspaceExists {
		fmt.Printf("Creating keyspace '%s'...\n", cfg.Cassandra.Keyspace)
		err = schemaRepo.CreateKeyspace(ctx, cfg.Cassandra.Keyspace, cfg.Schema)
		if err != nil {
			return fmt.Errorf("failed to create keyspace: %w", err)
		}
	}

	// Check and create table
	tableExists, err := schemaRepo.TableExists(ctx, cfg.Cassandra.Keyspace, cfg.Cassandra.Table)
	if err != nil {
		return fmt.Errorf("failed to check table: %w", err)
	}

	if !tableExists {
		fmt.Printf("Creating table '%s.%s'...\n", cfg.Cassandra.Keyspace, cfg.Cassandra.Table)
		err = schemaRepo.CreateTable(ctx, cfg.Cassandra.Keyspace, cfg.Cassandra.Table, cfg.Schema)
		if err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Create indexes
	if len(cfg.Schema.Indexes) > 0 {
		fmt.Printf("Setting up indexes...\n")
		err = schemaRepo.CreateIndexes(ctx, cfg.Cassandra.Keyspace, cfg.Cassandra.Table, cfg.Schema.Indexes)
		if err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}
	}

	fmt.Println("✅ Schema setup completed")
	return nil
}
